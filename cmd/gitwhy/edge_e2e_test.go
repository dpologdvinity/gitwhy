package main

import (
	"strings"
	"testing"
)

// TestRenameFileWithDashName checks that a file whose name starts with a dash can be renamed:
// without "--" git reads "-n.txt" as an option and the recipe fails.
func TestRenameFileWithDashName(t *testing.T) {
	r := newRepo(t)
	r.commit("dash", "-n.txt", "x\n")
	out := r.do("rename-file", pick("-n.txt"), text("renamed.txt"), yes)
	expect(t, "renamed file", r.read("renamed.txt"), "x\n", out)
	expect(t, "old name gone", r.read("-n.txt"), "<missing>", out)
}

// mergeRepo is a repository whose last commit is a merge of a side branch that added side.txt.
func mergeRepo(t *testing.T) *repo {
	t.Helper()
	r := newRepo(t)
	r.git("switch", "-q", "-c", "side")
	r.commit("side work", "side.txt", "s\n")
	r.git("switch", "-q", "main")
	r.commit("main work", "m.txt", "m\n")
	r.git("merge", "-q", "--no-ff", "-m", "merge side", "side")
	return r
}

// TestRevertMergeCommit checks that a merge commit can be reverted: git refuses without a
// mainline parent, so the recipe passes -m 1 for merges. The side branch's file must go away.
func TestRevertMergeCommit(t *testing.T) {
	r := mergeRepo(t)
	out := r.do("revert-commit", commit("merge side"), yes)
	expect(t, "side file removed", r.read("side.txt"), "<missing>", out)
	expect(t, "main file kept", r.read("m.txt"), "m\n", out)
	if !strings.HasPrefix(r.git("log", "-1", "--format=%s"), "Revert") {
		t.Fatalf("no revert commit:\n%s", out)
	}
}

// TestMoveCommitsRefusesTheCurrentBranch checks that picking the branch you are already on as the
// target stops with a plain message and changes nothing, instead of cherry-picking commits onto
// themselves and then resetting the branch.
func TestMoveCommitsRefusesTheCurrentBranch(t *testing.T) {
	r := newRepo(t)
	before := r.state()
	out := r.do("move-commits-to-branch", commit("c3"), pick("main"))
	if !strings.Contains(out, "already on") {
		t.Fatalf("want a refusal mentioning the branch you are on:\n%s", out)
	}
	expect(t, "state", r.state(), before, out)
}

// snapshotRefs counts gitwhy's undo points in the repository.
func snapshotRefs(r *repo) int {
	out := r.git("for-each-ref", "--format=%(refname)", "refs/gitwhy/snapshots")
	if out == "" {
		return 0
	}
	return len(strings.Split(out, "\n"))
}

// TestPartialUndoKeepsTheUndoPoint checks that when an undo has to leave HEAD and the files alone
// (the branch moved after the action), gitwhy keeps that undo point and says so, instead of
// deleting it and claiming the change was undone.
func TestPartialUndoKeepsTheUndoPoint(t *testing.T) {
	r := newRepo(t)
	drive(t, r.dir, []string{"run", "git", "reset", "--hard", "HEAD~1"}, yes)
	r.commit("later", "later.txt", "l\n") // the branch moves after the action
	before := snapshotRefs(r)
	out := drive(t, r.dir, []string{"undo"}, yes)
	if strings.Contains(out, "Undone.") {
		t.Fatalf("a partial undo must not claim it undid everything:\n%s", out)
	}
	if !strings.Contains(out, "kept") {
		t.Fatalf("expected a note that the undo point is kept:\n%s", out)
	}
	// The partial undo adds a "before undo" point and must not delete the one it used.
	if got := snapshotRefs(r); got != before+1 {
		t.Fatalf("snapshot refs: got %d, want %d (the used undo point kept)\n%s", got, before+1, out)
	}
}

// TestUndoWarnsWhenTheAfterStateWasNeverRecorded checks that undoing a snapshot whose "after"
// step never ran (the shell hook was killed between __pre and __post) says that everything,
// including later work, will be put back, because there is no record of what the command changed.
func TestUndoWarnsWhenTheAfterStateWasNeverRecorded(t *testing.T) {
	r := newRepo(t)
	r.write("f1.txt", "dirty\n")
	out := drive(t, r.dir, []string{"__pre", "reset", "--hard"}) // snapshot taken, __post never called
	if strings.TrimSpace(out) == "" {
		t.Fatalf("__pre took no snapshot:\n%s", out)
	}
	r.git("reset", "--hard")
	out = drive(t, r.dir, []string{"undo"}, yes)
	if !strings.Contains(out, "couldn't record") {
		t.Fatalf("expected a note that the after-state was never recorded:\n%s", out)
	}
	expect(t, "undo still restores the edit", r.read("f1.txt"), "dirty\n", out)
}

// TestPostWithoutTokenDoesNotSearch checks that invoking `__post` with no token returns silently
// instead of falling through to natural language search.
func TestPostWithoutTokenDoesNotSearch(t *testing.T) {
	r := newRepo(t)
	out := drive(t, r.dir, []string{"__post"})
	if strings.Contains(out, "Recipes") || strings.Contains(out, "Commands") {
		t.Fatalf("__post without token fell through to search:\n%s", out)
	}
}
