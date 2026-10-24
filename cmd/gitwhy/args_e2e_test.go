package main

// End-to-end tests for how gitwhy reads its own arguments and reports git's exit status.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestFlagsOnlyBeforeCommand proves that gitwhy reads -q, -h and version only before the first
// word that is not a flag. They used to be taken from anywhere, so the git command after
// "gitwhy run" lost its own -q and -h.
func TestFlagsOnlyBeforeCommand(t *testing.T) {
	r := newRepo(t)
	r.commit("hay", "hay.txt", "needle\n")

	out := drive(t, r.dir, []string{"run", "git", "grep", "-q", "needle"})
	if strings.Contains(out, "hay.txt:needle") {
		t.Errorf("git grep -q printed matches, so -q was dropped:\n%s", out)
	}

	r.write("-q", "file named -q\n")
	out = drive(t, r.dir, []string{"run", "git", "add", "--", "-q"}, yes)
	expect(t, "staged", r.git("status", "--porcelain", "--", "./-q"), "A  -q", out)

	out = drive(t, r.dir, []string{"run", "git", "commit", "-h"}, yes)
	if strings.Contains(out, "gitwhy: do anything with git") || !strings.Contains(out, "usage: git commit") {
		t.Errorf("want git's own help for git commit -h:\n%s", out)
	}

	out = drive(t, r.dir, []string{"-q", "version"})
	expect(t, "-q version", strings.TrimSpace(out), "gitwhy dev", out)
}

// TestTypedCommandsGlob proves that a read-only git command the user types after "gitwhy run"
// keeps git's normal pathspec globbing ('*.txt' matches files) even when
// GIT_LITERAL_PATHSPECS=1 is set for gitwhy's own git calls. TestTypedRunnerGlobs in
// internal/ui covers commands that change the repository.
func TestTypedCommandsGlob(t *testing.T) {
	t.Setenv("GIT_LITERAL_PATHSPECS", "1")
	r := newRepo(t)
	out := drive(t, r.dir, []string{"run", "git", "ls-files", "*.txt"})
	if !strings.Contains(out, "f1.txt") {
		t.Errorf("read-only command: '*.txt' matched nothing:\n%s", out)
	}
}

// TestNeedsGit240 proves that gitwhy refuses git older than 2.40 with a plain message. Previews
// use "git merge-tree --write-tree --merge-base", which git 2.39 rejects, so conflict checks
// used to fail or vanish silently there.
func TestNeedsGit240(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake git is a shell script")
	}
	fake := t.TempDir()
	if err := os.WriteFile(filepath.Join(fake, "git"), []byte("#!/bin/sh\necho 'git version 2.39.5'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(bin, "explain", "git", "status")
	cmd.Env = append(env(t.TempDir()), "PATH="+fake+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "needs git 2.40 or newer; you have 2.39") {
		t.Fatalf("want a refusal for git 2.39 (err %v):\n%s", err, out)
	}
}

// TestGitExitCodeKept proves that when git fails in a command gitwhy passes straight through,
// gitwhy exits with git's status and git's message is printed once. It used to exit 1 and print
// git's message a second time.
func TestGitExitCodeKept(t *testing.T) {
	r := newRepo(t)
	out, code := runEOF(t, r.dir, "run", "git", "log", "nosuch")
	expect(t, "exit code", fmt.Sprint(code), "128", out)
	if n := strings.Count(out, "ambiguous argument 'nosuch'"); n != 1 {
		t.Errorf("git's message printed %d times, want 1:\n%s", n, out)
	}
}
