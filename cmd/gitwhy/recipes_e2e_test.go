package main

// End-to-end tests for recipes that stop part way: the advice gitwhy gives must be right, and
// following it must get the user back to a good state.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// runInput runs gitwhy in dir with input as its whole stdin, then end of input. It returns the
// output and exit code, and fails the test if gitwhy is still running after 30 seconds.
func runInput(t *testing.T, dir, input string, args ...string) (string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Env = env(t.TempDir())
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("gitwhy %v still running after 30s; output starts:\n%.2000s", args, out)
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

// inProgress reports whether a rebase, merge or cherry-pick is still going on in r.
func (r *repo) inProgress() bool {
	for _, p := range []string{"rebase-merge", "rebase-apply", "MERGE_HEAD", "CHERRY_PICK_HEAD"} {
		if _, err := os.Stat(r.git("rev-parse", "--path-format=absolute", "--git-path", p)); err == nil {
			return true
		}
	}
	return false
}

// mustSay fails the test unless out contains every string in want and none in not.
func mustSay(t *testing.T, out string, want, not []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q", w)
		}
	}
	for _, n := range not {
		if strings.Contains(out, n) {
			t.Errorf("output has %q", n)
		}
	}
	if t.Failed() {
		t.Fatalf("--- output ---\n%s", out)
	}
}

// TestSplitCommitCancelledMidway proves that cancelling split-commit at its mid-recipe question,
// after the commit was taken apart inside a rebase, points to "gitwhy undo" and not to
// "gitwhy continue", and that undo brings the commit back. Continuing used to finish the rebase
// without the commit, leaving its files untracked.
func TestSplitCommitCancelledMidway(t *testing.T) {
	r := newRepo(t)
	r.commit("pair", "p.txt", "p\n", "q.txt", "q\n")
	r.commit("after", "w.txt", "w\n")
	before := r.subjects()
	// Commit 2 in the list is "pair"; then both messages and "Run it?". The file question
	// comes after the reset, and end of input cancels it.
	out, code := runInput(t, r.dir, "2\np only\nq only\ny\n", "do", "split-commit")
	if code == 0 {
		t.Fatalf("exit code 0 after cancelling\n%s", out)
	}
	mustSay(t, out, []string{"gitwhy undo", "did not run"}, []string{"gitwhy continue"})
	out = drive(t, r.dir, []string{"undo"}, yes)
	expect(t, "subjects after undo", r.subjects(), before, out)
	expect(t, "status after undo", r.git("status", "--porcelain"), "", out)
	if r.inProgress() {
		t.Fatalf("still in progress after undo\n%s", out)
	}
}

// TestFailedBisectAdvice proves that when "git bisect run" fails, gitwhy says to end the bisect
// with "gitwhy abort", that "gitwhy continue" refuses a bisect, and that abort goes back to the
// branch. It used to advise "gitwhy continue", which ran the invalid "git bisect --continue".
func TestFailedBisectAdvice(t *testing.T) {
	r := newRepo(t)
	out := r.do("find-bug-commit", commit("c1"), text("exit 139"), yes)
	mustSay(t, out, []string{"gitwhy abort"}, []string{"gitwhy continue"})
	out = drive(t, r.dir, []string{"continue"})
	mustSay(t, out, []string{"bisect", "gitwhy abort"}, []string{"--continue"})
	out = drive(t, r.dir, []string{"abort"}, yes)
	expect(t, "branch after abort", r.git("branch", "--show-current"), "main", out)
}

// TestSquashMergeConflictAdvice proves that a squash merge that stops on conflicts says so and
// lists the commit step that did not run. "git merge --squash" leaves no merge in progress, so
// gitwhy used to say "Nothing is lost" and never mention the conflicts or the commit.
func TestSquashMergeConflictAdvice(t *testing.T) {
	r := newRepo(t)
	r.git("switch", "-q", "-c", "other")
	r.commit("theirs", "f1.txt", "theirs\n")
	r.git("switch", "-q", "main")
	r.commit("ours", "f1.txt", "ours\n")
	out := r.do("merge-branch", pick("other "), pick("Squash"), text("squashed"), yes)
	mustSay(t, out, []string{"gitwhy resolve", "did not run", "git commit -m squashed"}, []string{"Nothing is lost"})
}

// TestMoveCommitsConflictAdvice proves that when move-commits-to-branch stops on a conflict
// half way, gitwhy lists the steps that did not run and points to "gitwhy undo", not to
// "gitwhy continue" (which would finish only the cherry-pick and leave the commits on both
// branches), and that undo puts both branches back.
func TestMoveCommitsConflictAdvice(t *testing.T) {
	r := newRepo(t)
	r.git("switch", "-q", "-c", "other", "HEAD~2")
	r.commit("other f4", "f4.txt", "other\n")
	r.git("switch", "-q", "main")
	main, other := r.git("rev-parse", "main"), r.git("rev-parse", "other")
	out := r.do("move-commits-to-branch", commit("c4"), pick("other "), yes)
	mustSay(t, out, []string{"did not run", "git switch main", "gitwhy undo"}, []string{"gitwhy continue"})
	out = drive(t, r.dir, []string{"undo"}, yes)
	expect(t, "main", r.git("rev-parse", "main"), main, out)
	expect(t, "other", r.git("rev-parse", "other"), other, out)
	expect(t, "branch", r.git("branch", "--show-current"), "main", out)
	if r.inProgress() {
		t.Fatalf("still in progress after undo\n%s", out)
	}
}

// TestCommitStaged proves that commit-staged commits the index exactly as it is: unstaged edits,
// even to a file that has staged changes, stay out of the commit and stay in the files. With
// nothing staged it stops with a plain message and changes nothing.
func TestCommitStaged(t *testing.T) {
	r := newRepo(t)
	before := r.state()
	out, code := runEOF(t, r.dir, "do", "commit-staged")
	if code == 0 || !strings.Contains(out, "Nothing is staged") {
		t.Fatalf("want a refusal with nothing staged (exit %d):\n%s", code, out)
	}
	expect(t, "state", r.state(), before, out)

	r.write("f1.txt", "staged\n")
	r.git("add", "f1.txt")
	r.write("f1.txt", "staged\nunstaged\n")
	r.write("f2.txt", "unstaged\n")
	out = r.do("commit-staged", text("only staged"), yes)
	expect(t, "subject", r.git("log", "-1", "--format=%s"), "only staged", out)
	expect(t, "committed f1", r.git("show", "HEAD:f1.txt"), "staged", out)
	expect(t, "files in commit", r.git("show", "--name-only", "--format=", "HEAD"), "f1.txt", out)
	expect(t, "status", r.git("status", "--porcelain"), "M f1.txt\n M f2.txt", out) // r.git trims the first space
}

// TestSquashNeedsKeptCommitAbove proves that the commit editor refuses a plan where a squash or
// fixup has no kept commit above it (here the first commit is dropped). git used to stop with
// "cannot 'squash' without a previous commit" and leave a rebase in progress.
func TestSquashNeedsKeptCommitAbove(t *testing.T) {
	r := newRepo(t)
	out := r.do("edit-recent-commits", text("3"), todoItem("c2"), pick("Drop"), todoItem("c3"), pick("Squash"),
		pick("Done"), todoItem("c3"), pick("Keep as is"), pick("Done"), yes)
	mustSay(t, out, []string{"can't be squashed"}, nil)
	expect(t, "subjects", r.subjects(), "c4\nc3\nc1", out)
	if r.inProgress() {
		t.Fatalf("rebase left in progress\n%s", out)
	}
}
