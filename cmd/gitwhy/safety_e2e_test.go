package main

// End-to-end regression tests for prompts, recipes and argument handling that once changed the
// repository without consent or gave wrong advice.

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
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

// TestFileNamesKeepSpaces proves that file pickers keep leading and trailing spaces in names.
// They used to trim them, so "keep" and "keep " were both listed as "keep" and picking the second
// threw away changes to the first.
func TestFileNamesKeepSpaces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows can't have a file name ending in a space")
	}
	r := newRepo(t)
	r.commit("two", "keep", "a\n", "keep ", "b\n")
	r.write("keep", "a edited\n")
	r.write("keep ", "b edited\n")
	// The list is sorted: 1 is "keep", 2 is "keep ".
	out := r.do("discard-file-changes", text("2"), text("0"), yes)
	expect(t, "keep (not picked)", r.read("keep"), "a edited\n", out)
	expect(t, "keep+space (picked)", r.read("keep "), "b\n", out)
}

// TestCommitRename proves that committing a renamed file (git mv) through the file picker commits
// the whole rename. Only the new path used to be passed, so the commit added a copy and left the
// deletion of the old path staged.
func TestCommitRename(t *testing.T) {
	for _, c := range []struct {
		recipe, rev string // rev is the commit that must hold the rename
		answers     []answer
	}{
		{"commit-changes", "HEAD", []answer{pick("new.txt"), text("0"), text("rename"), yes}},
		{"amend-last-commit", "HEAD", []answer{pick("new.txt"), text("0"), yes}},
		{"fix-past-commit", "HEAD~2", []answer{pick("new.txt"), text("0"), commit("c2"), yes}},
	} {
		t.Run(c.recipe, func(t *testing.T) {
			r := newRepo(t)
			r.git("mv", "f1.txt", "new.txt")
			out := r.do(c.recipe, c.answers...)
			expect(t, "status", r.git("status", "--porcelain"), "", out)
			expect(t, "in "+c.rev, r.git("show", "--name-status", "-M", "--format=", c.rev, "--", "f1.txt", "new.txt"),
				"R100\tf1.txt\tnew.txt", out)
		})
	}
}

// TestRemoteBranchPick proves that picking a remote branch (origin/feat) that has no local
// branch makes a local branch feat that tracks it. "git switch origin/feat" used to fail, and
// "git worktree add <dir> origin/feat" made a detached HEAD instead of the branch.
func TestRemoteBranchPick(t *testing.T) {
	up, c := withRemote(t)
	up.git("push", "-q", "origin", "HEAD~1:refs/heads/feat")
	c.git("fetch", "-q")

	out := c.do("switch-branch", pick("origin/feat"), yes)
	expect(t, "branch", c.git("branch", "--show-current"), "feat", out)
	expect(t, "upstream", c.git("rev-parse", "--abbrev-ref", "feat@{upstream}"), "origin/feat", out)

	c.git("switch", "-q", "main")
	c.git("branch", "-q", "-D", "feat")
	dir := filepath.Join(t.TempDir(), "wt")
	out = c.do("worktree-add", pick("origin/feat"), text(dir), yes)
	wt := &repo{t: t, dir: dir}
	expect(t, "worktree branch", wt.git("branch", "--show-current"), "feat", out)
	expect(t, "worktree upstream", wt.git("rev-parse", "--abbrev-ref", "feat@{upstream}"), "origin/feat", out)
}

// TestTypedPathWithSpaces proves that when there are too many files to list (over 2000) and the
// path is typed, a name with spaces is used as one path. It used to be split on spaces.
func TestTypedPathWithSpaces(t *testing.T) {
	r := newRepo(t)
	for i := range 2001 {
		r.write(fmt.Sprintf("many/%04d.txt", i), "x\n")
	}
	r.write("my file.txt", "x\n")
	out := r.do("stage-files", text("my file.txt"), yes)
	expect(t, "staged", r.git("diff", "--cached", "--name-only"), "my file.txt", out)
}
