package todo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestMain lets this test binary stand in for "gitwhy __todo <plan> <file>", so tests can run a
// real "git rebase -i" with Editor pointing at it.
func TestMain(m *testing.M) {
	if os.Getenv("GITWHY_TODO_TEST_EDITOR") == "1" && len(os.Args) == 4 && os.Args[1] == "__todo" {
		if err := Apply(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, "gitwhy:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// rebase runs a real "git rebase -i" with extra git args (such as "-c" options) and the plan
// applied by this test binary as the sequence editor.
func rebase(t *testing.T, r *testrepo.Repo, p Plan, args ...string) error {
	t.Helper()
	plan, err := Write(p)
	if err != nil {
		t.Fatal(err)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.With(Editor(self, plan), "GITWHY_TODO_TEST_EDITOR=1").Output(args...)
	return err
}

// TestCustomCommentChar is a regression test: with core.commentChar set to ";" the todo's help
// lines start with ";" and must be read as comments, not as steps that make the reorder refuse.
func TestCustomCommentChar(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	two := r.Commit("two", "b.txt", "2\n")
	three := r.Commit("three", "c.txt", "3\n")
	p := Plan{Full: []Item{{"pick", three, ""}, {"reword", two, "two again"}}}
	if err := rebase(t, r, p, "-c", "core.commentChar=;", "rebase", "-i", "HEAD~2"); err != nil {
		t.Fatal(err)
	}
	if got := r.Git("log", "--format=%s"); got != "two again\nthree\none" {
		t.Fatalf("log after rebase:\n%s", got)
	}
}

// TestPlanRemovedAfterRebase is a regression test: the temp directory Write makes must be gone once
// git has run the plan, including a reword whose message is needed after the editor exits.
func TestPlanRemovedAfterRebase(t *testing.T) {
	tmp := t.TempDir()
	setTempDir(t, tmp)
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	two := r.Commit("two", "b.txt", "2\n")
	r.Commit("three", "c.txt", "3\n")
	msg := "two, 'quoted' $HOME `x` 100%\n\nbody line\\n"
	if err := rebase(t, r, Plan{Edits: []Item{{Action: "reword", Commit: two, Message: msg}}}, "rebase", "-i", "HEAD~2"); err != nil {
		t.Fatal(err)
	}
	if got := r.Git("log", "-1", "--format=%B", "HEAD~1"); got != msg {
		t.Fatalf("reworded message = %q, want %q", got, msg)
	}
	if left, _ := filepath.Glob(filepath.Join(tmp, "gitwhy-todo-*")); len(left) != 0 {
		t.Fatalf("left behind: %v", left)
	}
}

// TestWriteRemovesStaleDirs proves that Write deletes gitwhy-todo-* directories older than a day
// (plans that were written but never run) and leaves recent ones and other files alone.
func TestWriteRemovesStaleDirs(t *testing.T) {
	tmp := t.TempDir()
	setTempDir(t, tmp)
	old := filepath.Join(tmp, "gitwhy-todo-old")
	recent := filepath.Join(tmp, "gitwhy-todo-recent")
	other := filepath.Join(tmp, "other-old")
	for _, d := range []string{old, recent, other} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	twoDays := time.Now().Add(-48 * time.Hour)
	for _, d := range []string{old, other} {
		if err := os.Chtimes(d, twoDays, twoDays); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Write(Plan{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("stale dir kept: %v", err)
	}
	for _, d := range []string{recent, other} {
		if _, err := os.Stat(d); err != nil {
			t.Errorf("%s removed: %v", d, err)
		}
	}
}

// setTempDir points os.TempDir at dir for one test, on Unix and Windows.
func setTempDir(t *testing.T, dir string) {
	t.Helper()
	for _, v := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(v, dir)
	}
}

const sample = `pick 1111111 one
pick 2222222 two
pick 3333333 three

# Rebase ...
`

func run(t *testing.T, p Plan) string {
	t.Helper()
	plan, err := Write(p)
	if err != nil {
		t.Fatal(err)
	}
	todo := filepath.Join(t.TempDir(), "git-rebase-todo")
	os.WriteFile(todo, []byte(sample), 0o644)
	if err := Apply(plan, todo); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(todo)
	return string(data)
}

func TestEdits(t *testing.T) {
	got := run(t, Plan{Edits: []Item{{Action: "drop", Commit: "2222222abcdef"}, {Action: "reword", Commit: "3333333ffff", Message: "new msg"}}})
	if !strings.Contains(got, "pick 1111111 one\ndrop 2222222 two\npick 3333333 three\nexec printf '%s\\n' 'new msg' | git commit --amend") {
		t.Fatalf("got:\n%s", got)
	}
}

func TestFullReorder(t *testing.T) {
	got := run(t, Plan{Full: []Item{{"pick", "3333333x", ""}, {"pick", "1111111x", ""}, {"fixup", "2222222x", ""}}})
	if !strings.HasPrefix(got, "pick 3333333 three\npick 1111111 one\nfixup 2222222 two\n") {
		t.Fatalf("got:\n%s", got)
	}
}

func TestMismatchFails(t *testing.T) {
	plan, _ := Write(Plan{Full: []Item{{"pick", "1111111x", ""}}})
	todo := filepath.Join(t.TempDir(), "todo")
	os.WriteFile(todo, []byte(sample), 0o644)
	if err := Apply(plan, todo); err == nil {
		t.Fatal("expected error for incomplete plan")
	}
}

func TestFullRefusesUpdateRef(t *testing.T) {
	plan, _ := Write(Plan{Full: []Item{{"pick", "1111111x", ""}}})
	todo := filepath.Join(t.TempDir(), "todo")
	os.WriteFile(todo, []byte("pick 1111111 one\nupdate-ref refs/heads/stack1\n"), 0o644)
	if err := Apply(plan, todo); err == nil {
		t.Fatal("expected refusal")
	}
}
