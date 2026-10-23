package snapshot

import (
	"os"
	"path/filepath"
	"testing"

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
	cur, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(r.Repo, cur); err != nil {
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
