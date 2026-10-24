package main

// End-to-end regression tests for prompts, recipes and argument handling that once changed the
// repository without consent or gave wrong advice.

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"
)

// state fingerprints everything a command could change: refs (including stashes and gitwhy's
// snapshots), the current branch, and staged, unstaged and untracked files.
func (r *repo) state() string {
	return r.git("rev-parse", "--all") + "\n" + r.git("symbolic-ref", "-q", "--short", "HEAD") + "\n" +
		r.git("status", "--porcelain", "--untracked-files=all")
}

// runEOF runs gitwhy in dir with an empty stdin and returns its output and exit code. It fails
// the test if gitwhy is still running after 20 seconds (a prompt that loops on EOF).
func runEOF(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Env = env(t.TempDir())
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("gitwhy %v still running after 20s on empty stdin; output starts:\n%.2000s", args, out)
	}
	code := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return ansi.ReplaceAllString(string(out), ""), code
}

// TestEOFCancels proves that end of input at any prompt cancels: nothing runs, the repository is
// unchanged and gitwhy exits non-zero. huh's plain-text prompts used to read EOF as "take the
// default", so scripts with no input stashed, reset, deleted branches and undid without consent.
func TestEOFCancels(t *testing.T) {
	for _, c := range []struct {
		name  string
		args  []string
		setup func(r *repo)
	}{
		{"stash-save", []string{"do", "stash-save"}, func(r *repo) { r.write("f1.txt", "wip\n") }},
		{"reset-branch-to", []string{"do", "reset-branch-to"}, func(r *repo) { r.git("switch", "-q", "-c", "feature", "HEAD~1") }},
		{"delete-branch", []string{"do", "delete-branch"}, func(r *repo) { r.git("branch", "old", "HEAD~1") }},
		{"commit-changes", []string{"do", "commit-changes"}, func(r *repo) { r.write("f1.txt", "edit\n") }},
		{"menu", []string{"menu"}, func(r *repo) { r.write("f1.txt", "edit\n") }},
		{"undo", []string{"undo"}, func(r *repo) {
			drive(r.t, r.dir, []string{"run", "git", "commit", "--allow-empty", "-m", "extra"}, yes)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t)
			c.setup(r)
			before := r.state()
			out, code := runEOF(t, r.dir, c.args...)
			if code == 0 {
				t.Errorf("exit code 0 on empty input, want non-zero\n%s", out)
			}
			expect(t, "repository state", r.state(), before, out)
		})
	}
}
