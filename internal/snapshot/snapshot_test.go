package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/git"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

func setup(t *testing.T) *testrepo.Repo {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "a1\n", "dir/b.txt", "b1\n", ".gitignore", "*.log\n")
	r.Commit("two", "a.txt", "a2\n")
	r.Git("branch", "feature")
	// Mixed working state.
	r.Write("a.txt", "a-staged\n")
	r.Git("add", "a.txt")
	r.Write("a.txt", "a-unstaged\n")
	r.Write("dir/b.txt", "b-unstaged\n")
	r.Write("new/untracked.txt", "u\n")
	r.Write("debug.log", "ignored\n")
	r.Git("stash", "push", "-q", "--", "dir/b.txt")
	r.Write("dir/b.txt", "b-again\n")
	return r
}

func TestRoundTripDestructive(t *testing.T) {
	r := setup(t)
	before := r.State()
	s, err := Take(r.Repo, "test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Wreck everything.
	r.Git("reset", "-q", "--hard", "HEAD~1")
	r.Git("clean", "-fdq")
	r.Git("stash", "drop", "-q")
	r.Git("branch", "-D", "feature")
	r.Git("branch", "extra")
	r.Write("created-later.txt", "x\n")

	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if after := r.State(); after != before {
		t.Fatalf("state differs\nbefore:\n%s\n\nafter:\n%s", before, after)
	}
	for path, want := range map[string]string{
		"a.txt": "a-unstaged\n", "dir/b.txt": "b-again\n", "new/untracked.txt": "u\n",
		"debug.log": "ignored\n", "created-later.txt": "<missing>",
	} {
		if got := r.Read(path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
}

func TestUndoTheUndo(t *testing.T) {
	r := setup(t)
	s, err := Take(r.Repo, "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("reset", "-q", "--hard", "HEAD~1")
	wrecked := r.State()
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(r.Repo, res.Before); err != nil {
		t.Fatal(err)
	}
	if got := r.State(); got != wrecked {
		t.Fatalf("undo of undo differs\nwant:\n%s\n\ngot:\n%s", wrecked, got)
	}
}

func TestConflictStateRestored(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "base\n")
	r.Git("switch", "-q", "-c", "other")
	r.Commit("theirs", "f.txt", "theirs\n")
	r.Git("switch", "-q", "main")
	r.Commit("ours", "f.txt", "ours\n")
	if _, err := r.Output("merge", "other"); err == nil {
		t.Fatal("expected conflict")
	}
	before := r.State()
	s, err := Take(r.Repo, "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("merge", "--abort")
	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if after := r.State(); after != before {
		t.Fatalf("state differs\nbefore:\n%s\n\nafter:\n%s", before, after)
	}
	if _, err := os.Stat(filepath.Join(r.GitDir, "MERGE_HEAD")); err != nil {
		t.Fatal("MERGE_HEAD not restored")
	}
}

func TestExtraIgnoredFiles(t *testing.T) {
	r := setup(t)
	s, err := Take(r.Repo, "clean -x", []string{"debug.log"})
	if err != nil {
		t.Fatal(err)
	}
	r.Git("clean", "-fdxq")
	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if got := r.Read("debug.log"); got != "ignored\n" {
		t.Fatalf("debug.log = %q", got)
	}
}

func TestEmptyRepoAndPrune(t *testing.T) {
	r := testrepo.New(t)
	r.Write("x.txt", "x\n")
	for i := 0; i < Keep+3; i++ {
		if _, err := Take(r.Repo, "t", nil); err != nil {
			t.Fatal(err)
		}
	}
	all, _ := List(r.Repo)
	if len(all) != Keep {
		t.Fatalf("kept %d, want %d", len(all), Keep)
	}
	os.Remove(filepath.Join(r.Top, "x.txt"))
	if _, err := Restore(r.Repo, all[0]); err != nil {
		t.Fatal(err)
	}
	if r.Read("x.txt") != "x\n" {
		t.Fatal("x.txt not restored in repo without commits")
	}
}

func TestNestedRepoAndSubmoduleUntouched(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "a\n")
	nested := testrepo.New(t)
	nested.Commit("n1", "n.txt", "n\n")
	// An untracked nested repo inside the working tree.
	dst := filepath.Join(r.Top, "nested")
	if out, err := r.Output("clone", "-q", nested.Top, dst); err != nil {
		t.Fatal(out, err)
	}
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	g := func(args ...string) {
		c := r.Runner
		c.Dir = dst
		if out, err := c.Output(args...); err != nil {
			t.Fatal(out, err)
		}
	}
	os.WriteFile(filepath.Join(dst, "n.txt"), []byte("edited\n"), 0o644)
	g("-c", "user.name=T", "-c", "user.email=t@t", "commit", "-qam", "n2")
	os.WriteFile(filepath.Join(dst, "n.txt"), []byte("uncommitted\n"), 0o644)
	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dst, "n.txt"))
	if string(data) != "uncommitted\n" {
		t.Fatalf("nested repo file = %q", data)
	}
	if _, err := os.Stat(filepath.Join(dst, ".git")); err != nil {
		t.Fatal("nested .git deleted")
	}
}

func TestSymrefOriginHead(t *testing.T) {
	up := testrepo.New(t)
	up.Commit("one", "a.txt", "1\n")
	clone := t.TempDir()
	if out, err := up.Output("clone", "-q", up.Top, clone); err != nil {
		t.Fatal(out, err)
	}
	c, _ := git.Open(clone)
	s, err := Take(c, "fetch", nil)
	if err != nil {
		t.Fatal(err)
	}
	up.Commit("two", "a.txt", "2\n")
	if out, err := c.Output("fetch", "-q"); err != nil {
		t.Fatal(out, err)
	}
	if _, err := Restore(c, s); err != nil {
		t.Fatal(err)
	}
	head, _ := c.Output("symbolic-ref", "refs/remotes/origin/HEAD")
	if head != "refs/remotes/origin/main" {
		t.Fatalf("origin/HEAD = %q", head)
	}
}

func TestRestoreOldestDoesntPruneItself(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	for i := 0; i < Keep; i++ {
		r.Write("a.txt", strings.Repeat("x", i+1))
		if _, err := Take(r.Repo, "t", nil); err != nil {
			t.Fatal(err)
		}
	}
	all, _ := List(r.Repo)
	oldest := all[len(all)-1]
	if _, err := Restore(r.Repo, oldest); err != nil {
		t.Fatal(err)
	}
	if r.Read("a.txt") != "x" {
		t.Fatalf("a.txt = %q", r.Read("a.txt"))
	}
}

func TestIgnoredFileOverwriteIsSaved(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Write("notes.txt", "v1\n")
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Write(".git/info/exclude", "notes.txt\n") // now ignored
	r.Write("notes.txt", "v2 precious\n")
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if r.Read("notes.txt") != "v1\n" {
		t.Fatalf("notes = %q", r.Read("notes.txt"))
	}
	if _, err := Restore(r.Repo, res.Before); err != nil {
		t.Fatal(err)
	}
	if r.Read("notes.txt") != "v2 precious\n" {
		t.Fatalf("undo of undo lost the ignored edit: %q", r.Read("notes.txt"))
	}
}

func TestRefsChangedLaterAreLeftAlone(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Git("branch", "feat")
	s, err := Take(r.Repo, "commit", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Commit("by gitwhy", "a.txt", "2\n")
	if err := Finish(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	// Later, in plain git: a commit on feat and a new branch.
	feat := r.Git("rev-parse", "feat")
	r.Git("switch", "-q", "feat")
	r.Commit("later", "b.txt", "b\n")
	later := r.Git("rev-parse", "HEAD")
	r.Git("branch", "newbr")
	r.Git("switch", "-q", "main")
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Git("log", "-1", "--format=%s", "main"); got != "one" {
		t.Errorf("main = %q, want one", got)
	}
	if got := r.Git("rev-parse", "feat"); got != later || got == feat {
		t.Errorf("feat was moved back")
	}
	if !r.OK("rev-parse", "--verify", "-q", "refs/heads/newbr") {
		t.Errorf("newbr deleted")
	}
	if len(res.Skipped) != 2 {
		t.Errorf("skipped = %q", res.Skipped)
	}
}

func TestSparseCheckoutUndo(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "d1/x", "x\n", "d2/y", "y\n")
	s, err := Take(r.Repo, "sparse", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("sparse-checkout", "set", "--cone", "d1")
	if r.Read("d2/y") != "<missing>" {
		t.Fatal("sparse checkout didn't hide d2")
	}
	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if r.Read("d2/y") != "y\n" {
		t.Fatalf("d2/y = %q", r.Read("d2/y"))
	}
	if st := r.Git("status", "--porcelain"); st != "" {
		t.Fatalf("status = %q", st)
	}
}
