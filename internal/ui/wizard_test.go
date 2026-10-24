package ui

import (
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/recipe"
	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestDeriveBranchLocalAndTrack proves which local branch a branch answer resolves to: a remote
// branch with no local branch of that name (picked as "origin/feat", or passed by the TUI as
// "feat") becomes a new local branch that tracks it, and an existing local branch is used as is.
// This matters because switching to "origin/feat" directly would fail or detach HEAD.
func TestDeriveBranchLocalAndTrack(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "a\n")
	r.Git("remote", "add", "origin", "https://example.com/x.git")
	r.Git("update-ref", "refs/remotes/origin/feat", "HEAD")
	a := &App{Repo: r.Repo}
	// check derives answers for branch v and compares target_local and target_track.
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

// TestDefaultBranch checks which branch the "default" built-in answer names: the remote's own
// default (origin/HEAD) first, then origin/main, origin/master, and a local main or master. It
// replaced a hard-coded origin/master that preselected nothing in repositories that use main.
func TestDefaultBranch(t *testing.T) {
	cases := map[string]func(r *testrepo.Repo){
		"origin/main": func(r *testrepo.Repo) {
			r.Git("update-ref", "refs/remotes/origin/main", "HEAD")
			r.Git("update-ref", "refs/remotes/origin/master", "HEAD")
			r.Git("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
		},
		"origin/master": func(r *testrepo.Repo) { r.Git("update-ref", "refs/remotes/origin/master", "HEAD") },
		"origin/trunk": func(r *testrepo.Repo) {
			r.Git("update-ref", "refs/remotes/origin/trunk", "HEAD")
			r.Git("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk")
		},
		"main": func(r *testrepo.Repo) {}, // the test repository's own branch, no remote
	}
	for want, setup := range cases {
		t.Run(want, func(t *testing.T) {
			r := testrepo.New(t)
			r.Commit("one", "a.txt", "a\n")
			setup(r)
			a := &App{Repo: r.Repo}
			if got := a.builtins()["default_branch"]; got != want {
				t.Fatalf("default_branch = %q, want %q", got, want)
			}
		})
	}
}

// TestDeriveCommitFlagsMergeCommits checks that a picked commit is marked as a merge when it has
// a second parent, so recipes that need a mainline (revert) can tell.
func TestDeriveCommitFlagsMergeCommits(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "a.txt", "a\n")
	r.Git("switch", "-q", "-c", "side")
	r.Commit("side", "s.txt", "s\n")
	r.Git("switch", "-q", "main")
	r.Commit("main work", "m.txt", "m\n")
	r.Git("merge", "-q", "--no-ff", "-m", "merge side", "side")
	a := &App{Repo: r.Repo}
	ans := recipe.Answers{}
	a.deriveCommit("commit", r.Git("rev-parse", "HEAD"), ans)
	if ans["commit_merge"] != "1" {
		t.Errorf("merge commit: commit_merge = %q, want 1", ans["commit_merge"])
	}
	a.deriveCommit("commit", r.Git("rev-parse", "HEAD~1"), ans)
	if ans["commit_merge"] != "" {
		t.Errorf("plain commit: commit_merge = %q, want empty", ans["commit_merge"])
	}
}

// TestRunRecipeWithRejectsAPresetOfTheWrongType checks that a preset answer that is not text for
// a commit or branch question is reported as an error instead of crashing the program.
func TestRunRecipeWithRejectsAPresetOfTheWrongType(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "a\n")
	a := &App{Repo: r.Repo}
	rec := &recipe.Recipe{ID: "x", Title: "x", Questions: []recipe.Question{{ID: "commit", Type: "commit"}}}
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("panicked: %v", p)
		}
	}()
	if err := a.RunRecipeWith(rec, recipe.Answers{"commit": []string{"abc"}}); err == nil {
		t.Fatal("want an error for a non-text commit preset")
	}
}
