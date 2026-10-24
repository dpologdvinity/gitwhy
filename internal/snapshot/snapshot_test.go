package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

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

func TestLFSFilesRoundTrip(t *testing.T) {
	if _, err := (git.Runner{}).Output("lfs", "version"); err != nil {
		t.Skip("git-lfs not installed")
	}
	r := testrepo.New(t)
	home := t.TempDir()
	r.Runner = r.With("HOME="+home, "GIT_CONFIG_GLOBAL="+filepath.Join(home, "gitconfig"))
	r.Git("lfs", "install", "--local", "--skip-repo") // filters only; test repos have no hooks dir
	r.Git("lfs", "track", "*.bin")
	r.Commit("lfs", "big.bin", "binary v1\n")
	if ptr := r.Git("cat-file", "-p", "HEAD:big.bin"); !strings.Contains(ptr, "git-lfs") {
		t.Fatalf("not stored as LFS pointer: %q", ptr)
	}
	r.Write("big.bin", "binary v2 unstaged\n")
	r.Write("new.bin", "new untracked lfs\n")
	before := r.State()
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("reset", "-q", "--hard")
	r.Git("clean", "-fdq")
	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if got := r.Read("big.bin"); got != "binary v2 unstaged\n" {
		t.Fatalf("big.bin = %q (pointer instead of content?)", got)
	}
	if got := r.Read("new.bin"); got != "new untracked lfs\n" {
		t.Fatalf("new.bin = %q", got)
	}
	if after := r.State(); after != before {
		t.Fatalf("state differs\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestAutoCRLFRoundTrip(t *testing.T) {
	r := testrepo.New(t)
	r.Git("config", "core.autocrlf", "true")
	r.Commit("one", "a.txt", "line1\r\nline2\r\n")
	r.Write("a.txt", "edited1\r\nedited2\r\n")
	r.Write("b.txt", "new\r\n")
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("reset", "-q", "--hard")
	r.Git("clean", "-fdq")
	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if got := r.Read("a.txt"); got != "edited1\r\nedited2\r\n" {
		t.Fatalf("a.txt = %q", got)
	}
	if got := r.Read("b.txt"); got != "new\r\n" {
		t.Fatalf("b.txt = %q", got)
	}
}

// TestRestoreManyHiddenFilesIsFast proves restore brings back files hidden by a sparse
// checkout with a fixed number of git processes: it used to start one per file.
func TestRestoreManyHiddenFilesIsFast(t *testing.T) {
	r := testrepo.New(t)
	r.Write("d1/x", "x\n")
	for i := 0; i < 400; i++ {
		r.Write(fmt.Sprintf("d2/f%d", i), "y\n")
	}
	r.Commit("one")
	s, err := Take(r.Repo, "sparse", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("sparse-checkout", "set", "--cone", "d1")
	start := time.Now()
	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 8*time.Second {
		t.Errorf("Restore took %s for 400 hidden files", d)
	}
	if r.Read("d2/f399") != "y\n" {
		t.Fatalf("d2/f399 = %q", r.Read("d2/f399"))
	}
}

// TestTagsAndRemoteRefsSurviveGC proves a snapshot keeps alive what tags (including annotated
// tag objects) and remote-tracking branches point to, so gc can't break their undo.
func TestTagsAndRemoteRefsSurviveGC(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	tree := r.Git("rev-parse", "HEAD^{tree}")
	tagged := r.Git("commit-tree", tree, "-p", "HEAD", "-m", "only tagged")
	r.Git("tag", "-a", "v1", "-m", "release", tagged)
	tagObj := r.Git("rev-parse", "refs/tags/v1")
	remote := r.Git("commit-tree", tree, "-p", "HEAD", "-m", "only on the remote")
	r.Git("update-ref", "refs/remotes/origin/x", remote)
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("tag", "-d", "v1")
	r.Git("update-ref", "-d", "refs/remotes/origin/x")
	r.Git("gc", "-q", "--prune=now")
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Skipped) != 0 {
		t.Errorf("skipped = %q", res.Skipped)
	}
	if got, _ := r.Output("rev-parse", "refs/tags/v1"); got != tagObj {
		t.Errorf("v1 = %q, want tag object %s", got, tagObj)
	}
	if got, _ := r.Output("rev-parse", "refs/remotes/origin/x"); got != remote {
		t.Errorf("origin/x = %q, want %s", got, remote)
	}
	// Deleting the snapshot removes everything that pinned it.
	Delete(r.Repo, s)
	Delete(r.Repo, res.Before)
	if pins := r.Git("for-each-ref", "refs/gitwhy/"); pins != "" {
		t.Errorf("refs left after Delete:\n%s", pins)
	}
}

// TestUndoRenameIntoOwnFolder proves undo of "git branch -m foo foo/bar" works: the ref
// foo/bar must be deleted before foo can be written back.
func TestUndoRenameIntoOwnFolder(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Git("branch", "foo")
	s, err := Take(r.Repo, "git branch -m foo foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("branch", "-m", "foo", "foo/bar")
	if err := Finish(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if got := r.Git("for-each-ref", "--format=%(refname)", "refs/heads/"); got != "refs/heads/foo\nrefs/heads/main" {
		t.Fatalf("branches = %q", got)
	}
}

// TestFailedRefUpdatePutsDeletedRefsBack proves that when writing refs back fails after the
// first transaction deleted some, those refs are put back, so the error's "nothing was changed"
// stays true.
func TestFailedRefUpdatePutsDeletedRefsBack(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Git("branch", "foo/bar")
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("branch", "-D", "foo/bar")
	r.Git("branch", "extra")
	// foo is checked out in another worktree, so it is left alone and foo/bar can't come back.
	r.Git("worktree", "add", "-q", "-b", "foo", filepath.Join(t.TempDir(), "wt"))
	if _, err := Restore(r.Repo, s); err == nil || !strings.Contains(err.Error(), "nothing was changed") {
		t.Fatalf("err = %v", err)
	}
	if !r.OK("rev-parse", "--verify", "-q", "refs/heads/extra") {
		t.Fatal("extra was deleted")
	}
}

// TestRedoLeavesLaterRefsAlone proves redo (undo of an undo) leaves alone a branch moved after
// the undo: the undo's own snapshot must record the refs it left behind.
func TestRedoLeavesLaterRefsAlone(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Git("branch", "feat")
	s, err := Take(r.Repo, "commit", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Commit("c2", "a.txt", "2\n")
	if err := Finish(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("switch", "-q", "feat")
	later := r.Commit("later on feat", "b.txt", "b\n")
	r.Git("switch", "-q", "main")
	redo, err := Restore(r.Repo, res.Before)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Git("rev-parse", "feat"); got != later {
		t.Errorf("redo moved feat back")
	}
	if got := r.Git("log", "-1", "--format=%s", "main"); got != "c2" {
		t.Errorf("main = %q, want c2", got)
	}
	if len(redo.Skipped) != 1 || !strings.Contains(redo.Skipped[0], "feat") {
		t.Errorf("skipped = %q", redo.Skipped)
	}
}

// TestHeadBranchChangedLaterKeepsFiles proves that when the branch HEAD is on moved after the
// action, undo leaves HEAD, the index and the files alone instead of restoring old files under
// the new commit.
func TestHeadBranchChangedLaterKeepsFiles(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Commit("two", "a.txt", "2\n")
	s, err := Take(r.Repo, "git reset --hard HEAD~1", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("reset", "-q", "--hard", "HEAD~1")
	if err := Finish(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	later := r.Commit("later", "b.txt", "b\n")
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Git("rev-parse", "HEAD"); got != later {
		t.Errorf("HEAD moved")
	}
	if r.Read("b.txt") != "b\n" {
		t.Errorf("b.txt = %q", r.Read("b.txt"))
	}
	if st := r.Git("status", "--porcelain"); st != "" {
		t.Errorf("status = %q", st)
	}
	if len(res.Skipped) != 2 {
		t.Errorf("skipped = %q", res.Skipped)
	}
}

// TestHeadBranchInOtherWorktreeIsNotCheckedOut proves undo never checks out a branch that is
// now checked out in another worktree.
func TestHeadBranchInOtherWorktreeIsNotCheckedOut(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Git("switch", "-q", "-c", "feat")
	s, err := Take(r.Repo, "git switch main", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("switch", "-q", "main")
	if err := Finish(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	r.Git("worktree", "add", "-q", filepath.Join(t.TempDir(), "wt"), "feat")
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Git("symbolic-ref", "HEAD"); got != "refs/heads/main" {
		t.Errorf("HEAD = %q, want main", got)
	}
	if len(res.Skipped) != 1 || !strings.Contains(res.Skipped[0], "feat") {
		t.Errorf("skipped = %q", res.Skipped)
	}
}

// TestStashMadeLaterIsKept proves undo keeps a stash made after the action, while a stash the
// action itself made is still undone.
func TestStashMadeLaterIsKept(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Write("a.txt", "action wip\n")
	s, err := Take(r.Repo, "git stash", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("stash", "push", "-q", "-m", "made by the action")
	if err := Finish(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	r.Write("a.txt", "later wip\n")
	r.Git("stash", "push", "-q", "-m", "my later wip")
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Git("stash", "list", "--format=%gs"); got != "On main: my later wip" {
		t.Errorf("stash list = %q", got)
	}
	if r.Read("a.txt") != "action wip\n" {
		t.Errorf("a.txt = %q", r.Read("a.txt"))
	}
	if len(res.Skipped) != 1 || !strings.Contains(res.Skipped[0], "my later wip") {
		t.Errorf("skipped = %q", res.Skipped)
	}
}

// TestFileErrorStillRestoresOtherFiles proves one file that can't be removed doesn't leave
// other files deleted: restore carries on, writes every file back, and says how to go back.
func TestFileErrorStillRestoresOtherFiles(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs a read-only folder")
	}
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Commit("two", "a.txt", "2\n")
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("reset", "-q", "--hard", "HEAD~1")
	r.Write("ro/new.txt", "n\n")
	ro := filepath.Join(r.Top, "ro")
	if err := os.Chmod(ro, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(ro, 0o755) })
	_, err = Restore(r.Repo, s)
	if err == nil || !strings.Contains(err.Error(), "gitwhy undo") {
		t.Fatalf("err = %v", err)
	}
	if got := r.Read("a.txt"); got != "2\n" {
		t.Fatalf("a.txt = %q", got)
	}
}

// TestIgnoredFilesInReplacedFolderAreSaved proves that ignored files undo deletes, inside a
// folder or as a file where a folder must go, are saved first, so undo of the undo brings them
// back.
func TestIgnoredFilesInReplacedFolderAreSaved(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", ".gitignore", "*.env\n")
	r.Write("config", "plain file\n")
	r.Write("dir/f", "f\n")
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(filepath.Join(r.Top, "config"))
	r.Write("config/local.env", "SECRET\n")
	os.RemoveAll(filepath.Join(r.Top, "dir"))
	r.Write("dir.env", "x\n")
	if err := os.Rename(filepath.Join(r.Top, "dir.env"), filepath.Join(r.Top, "dir")); err != nil {
		t.Fatal(err)
	}
	r.Write(".git/info/exclude", "/dir\n") // dir is now an ignored file where a folder must go
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if r.Read("config") != "plain file\n" || r.Read("dir/f") != "f\n" {
		t.Fatalf("config = %q, dir/f = %q", r.Read("config"), r.Read("dir/f"))
	}
	if _, err := Restore(r.Repo, res.Before); err != nil {
		t.Fatal(err)
	}
	if got := r.Read("config/local.env"); got != "SECRET\n" {
		t.Errorf("config/local.env = %q", got)
	}
	if got := r.Read("dir"); got != "x\n" {
		t.Errorf("dir = %q", got)
	}
}

// TestSparseRedoStaysInThisWorktree proves undo and redo of a sparse checkout don't write
// sparse settings to the shared config, where every other worktree would pick them up.
func TestSparseRedoStaysInThisWorktree(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "d1/x", "x\n", "d2/y", "y\n")
	other := filepath.Join(t.TempDir(), "wt")
	r.Git("worktree", "add", "-q", "-b", "other", other)
	s, err := Take(r.Repo, "sparse", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Git("sparse-checkout", "set", "--cone", "d1")
	res, err := Restore(r.Repo, s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(r.Repo, res.Before); err != nil {
		t.Fatal(err)
	}
	if r.Read("d2/y") != "<missing>" {
		t.Errorf("redo didn't bring the sparse checkout back")
	}
	if v, _ := r.Output("config", "--file", filepath.Join(r.Common, "config"), "--get", "core.sparseCheckout"); v != "" {
		t.Errorf("shared config has core.sparseCheckout=%s", v)
	}
	o := r.Runner
	o.Dir = other
	if v, _ := o.Output("config", "--get", "core.sparseCheckout"); v == "true" {
		t.Errorf("the other worktree is now sparse")
	}
}

// TestPruneRemovesOrphans proves prune removes what an interrupted Take or a removed worktree
// leaves behind (an old folder without meta.json, refs without a folder), and keeps a folder
// that may still be in use and the snapshots of a live worktree.
func TestPruneRemovesOrphans(t *testing.T) {
	r := testrepo.New(t)
	head := r.Commit("one", "a.txt", "1\n")
	old := filepath.Join(r.GitDir, "gitwhy", "snapshots", "oldhalf")
	fresh := filepath.Join(r.GitDir, "gitwhy", "snapshots", "freshhalf")
	r.Write(".git/gitwhy/snapshots/oldhalf/index", "x")
	r.Write(".git/gitwhy/snapshots/freshhalf/index", "x")
	hour := time.Now().Add(-time.Hour)
	if err := os.Chtimes(old, hour, hour); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"oldhalf", "freshhalf", "nofolder", "nofolder-" + head} {
		r.Git("update-ref", "refs/gitwhy/snapshots/"+id, head)
	}
	wtDir := filepath.Join(t.TempDir(), "wt")
	r.Git("worktree", "add", "-q", "-b", "other", wtDir)
	wt, err := git.Open(wtDir)
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := Take(wt, "in the other worktree", nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Take(r.Repo, "t", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); err == nil {
		t.Error("old folder without meta.json kept")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("fresh folder deleted")
	}
	refs := r.Git("for-each-ref", "--format=%(refname)", "refs/gitwhy/snapshots/")
	for _, id := range []string{"oldhalf", "nofolder"} {
		if strings.Contains(refs, "/"+id) {
			t.Errorf("ref %s kept", id)
		}
	}
	for _, id := range []string{"freshhalf", theirs.ID} {
		if !strings.Contains(refs, "/"+id+"\n") && !strings.HasSuffix(refs, "/"+id) {
			t.Errorf("ref %s deleted", id)
		}
	}

	r.Git("worktree", "remove", wtDir)
	if _, err := Take(r.Repo, "t", nil); err != nil {
		t.Fatal(err)
	}
	if refs := r.Git("for-each-ref", "refs/gitwhy/snapshots/"+theirs.ID); refs != "" {
		t.Errorf("ref of a removed worktree kept: %s", refs)
	}
}

// manyBranches makes n branches, each on its own commit, with one fast-import process.
func manyBranches(t *testing.T, r *testrepo.Repo, n int) {
	t.Helper()
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "commit refs/heads/b%d\ncommitter T <t@t> %d +0000\ndata 2\nc\n", i, 1000000+i)
		if i == 0 {
			b.WriteString("from refs/heads/main\n")
		} else {
			fmt.Fprintf(&b, "from refs/heads/b%d\n", i-1)
		}
		fmt.Fprintf(&b, "M 644 inline f\ndata %d\n%d\n\n", len(strconv.Itoa(i))+1, i)
	}
	if out, err := r.OutputIn(strings.NewReader(b.String()), "fast-import", "--quiet"); err != nil {
		t.Fatal(out, err)
	}
}

// TestTakeManyBranchesIsFast proves Take checks branch tips in one git process: with hundreds
// of branches it used to start one process per branch and take tens of seconds.
func TestTakeManyBranchesIsFast(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	manyBranches(t, r, 400)
	start := time.Now()
	s, err := Take(r.Repo, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 8*time.Second {
		t.Errorf("Take took %s with 400 branches", d)
	}
	parents := strings.Fields(r.Git("log", "-1", "--format=%P", s.Work))
	if len(parents) != 401 {
		t.Errorf("work commit has %d parents, want 401 (every branch tip)", len(parents))
	}
}
