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
