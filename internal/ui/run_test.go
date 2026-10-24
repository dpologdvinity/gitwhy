package ui

import (
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestTypedRunnerGlobs proves that a typed command (a literal job) runs with git's normal
// pathspec globbing even when GIT_LITERAL_PATHSPECS=1 is set for gitwhy's own git calls, while
// recipe steps keep literal paths. Without this, "gitwhy run git add '*.txt'" would match no
// files, and a recipe given a file named "*.txt" could touch other files.
func TestTypedRunnerGlobs(t *testing.T) {
	t.Setenv("GIT_LITERAL_PATHSPECS", "1")
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "a\n")
	a := &App{Repo: r.Repo}
	if out, _ := a.runner(job{literal: []string{"ls-files"}}).Output("ls-files", "*.txt"); out != "a.txt" {
		t.Errorf("typed command: ls-files '*.txt' = %q, want a.txt", out)
	}
	if out, _ := a.runner(job{}).Output("ls-files", "*.txt"); out != "" {
		t.Errorf("recipe step: ls-files '*.txt' = %q, want nothing (literal)", out)
	}
}
