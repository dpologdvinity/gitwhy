package ui

import (
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/recipe"
	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestDeriveBranchLocalAndTrack proves which local branch a branch answer resolves to: a remote
// branch with no local branch of that name (picked as "origin/feat", or passed by the TUI as
// "feat") becomes a new local branch that tracks it, and an existing local branch is used as is.
func TestDeriveBranchLocalAndTrack(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "a\n")
	r.Git("remote", "add", "origin", "https://example.com/x.git")
	r.Git("update-ref", "refs/remotes/origin/feat", "HEAD")
	a := &App{Repo: r.Repo}
	check := func(v, local, track string) {
		t.Helper()
		ans := recipe.Answers{}
		a.deriveBranch("target", v, ans)
		if ans["target_local"] != local || ans["target_track"] != track {
			t.Errorf("%q: local %q track %q, want %q %q", v, ans["target_local"], ans["target_track"], local, track)
		}
	}
	check("main", "main", "")
	check("origin/feat", "feat", "origin/feat")
	check("feat", "feat", "origin/feat")
	r.Git("branch", "feat")
	check("origin/feat", "feat", "")
	check("feat", "feat", "")
}
