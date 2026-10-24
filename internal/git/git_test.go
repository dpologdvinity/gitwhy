package git_test

import (
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestFileNamesAreLiteral proves that a file name with glob characters only matches that file:
// staging or restoring "[ab].txt" must not touch a.txt. A user's own global pathspec setting
// must not make git refuse to run.
func TestFileNamesAreLiteral(t *testing.T) {
	t.Setenv("GIT_GLOB_PATHSPECS", "1")
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "a1\n", "[ab].txt", "g1\n")
	r.Write("a.txt", "a2\n")
	r.Write("[ab].txt", "g2\n")
	r.Git("add", "-A", "--", "[ab].txt")
	if got := r.Git("diff", "--cached", "--name-only"); got != "[ab].txt" {
		t.Fatalf("staged %q, want only [ab].txt", got)
	}
	r.Git("restore", "--staged", "--worktree", "--", "[ab].txt")
	if got := r.Read("a.txt"); got != "a2\n" {
		t.Fatalf("a.txt = %q, restore of [ab].txt reverted it", got)
	}
}

// TestStashIncludesUntrackedFiles proves "stash push --include-untracked" still sets untracked
// files aside: git stash uses the ":/" pathspec itself, which literal pathspecs break.
func TestStashIncludesUntrackedFiles(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Write("a.txt", "2\n")
	r.Write("new.txt", "n\n")
	r.Git("stash", "push", "--include-untracked", "-m", "wip")
	if st := r.Git("status", "--porcelain"); st != "" {
		t.Fatalf("status = %q, want clean", st)
	}
}
