package main

// End-to-end tests: build the real binary and drive recipes through accessible (plain-text) mode.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// bin is the path of the gitwhy binary that TestMain builds for every end-to-end test.
var bin string

// TestMain builds the real binary once into a temporary folder, runs the tests against it and
// removes it, so the tests exercise exactly what users run.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gitwhy-e2e-")
	if err != nil {
		panic(err)
	}
	bin = filepath.Join(dir, "gitwhy")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		panic(string(out))
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// answer is a typed line, or a menu pick chosen by the option's text.
type answer struct {
	text string // sent as is
	pick string // text of the option to choose; drive sends its number
}

// text answers a prompt with the line s.
func text(s string) answer { return answer{text: s} }

// pick answers a menu with the last listed option whose text contains s. A leading "$" means
// the option text must end with the rest of s.
func pick(s string) answer { return answer{pick: s} }

// todoItem picks a line of the rebase editor by its subject (the line ends with it).
func todoItem(s string) answer { return answer{pick: "$ " + s} }

// commit picks a commit by its subject (option text is "<hash> <subject>  <when>").
func commit(s string) answer { return answer{pick: " " + s + "  "} }

// yes answers a yes/no prompt with "y".
var yes = text("y")

// syncBuf is a buffer that the running process writes to while drive reads it.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

// Write appends p under the lock.
func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

// String returns everything written so far.
func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// ansi matches color codes, which are stripped from the output. optLine matches a numbered
// option of a plain-text menu ("3. text", with a "✓ " mark on the selected one).
var (
	ansi    = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	optLine = regexp.MustCompile(`(?m)^\s*(\d+)\.\s+(?:✓ )?(.*)$`)
)

// env isolates git config and HOME so tests never touch the user's setup.
func env(home string) []string {
	cfg := filepath.Join(home, "gitconfig")
	os.WriteFile(cfg, []byte("[user]\n\tname = Test\n\temail = test@example.com\n[init]\n\tdefaultBranch = main\n[commit]\n\tgpgsign = false\n[core]\n\tautocrlf = false\n"), 0o644)
	return append(os.Environ(), "ACCESSIBLE=1", "HOME="+home, "GIT_CONFIG_GLOBAL="+cfg, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_EDITOR=true", "LC_ALL=C", "XDG_CONFIG_HOME="+home)
}

// drive runs gitwhy in dir with args, answering prompts in order. A text answer is sent at
// once; it waits in stdin until the prompt reads it. A pick waits (up to 20 seconds) until the
// option is listed after the previous pick and the menu has asked for a number, then sends that
// number. It returns all output without color codes and fails after 60 seconds.
func drive(t *testing.T, dir string, args []string, answers ...answer) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = env(t.TempDir())
	stdin, _ := cmd.StdinPipe()
	out := &syncBuf{}
	cmd.Stdout, cmd.Stderr = out, out
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	seen := 0
	for _, a := range answers {
		line := a.text
		if a.pick != "" {
			deadline := time.Now().Add(20 * time.Second)
			for {
				s := ansi.ReplaceAllString(out.String(), "")
				recent := s[min(seen, len(s)):]
				key := strings.TrimPrefix(a.pick, "$")
				if n := find(recent, a.pick); n != "" && strings.Contains(strings.ToLower(recent[strings.LastIndex(recent, key):]), "number") {
					line = n
					seen = len(s)
					break
				}
				if time.Now().After(deadline) {
					cmd.Process.Kill()
					t.Fatalf("no option %q in:\n%s", a.pick, s)
				}
				time.Sleep(20 * time.Millisecond)
			}
		}
		io.WriteString(stdin, line+"\n")
	}
	select {
	case <-done:
	case <-time.After(60 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("timed out:\n%s", out.String())
	}
	return ansi.ReplaceAllString(out.String(), "")
}

// find returns the number of the last listed option whose text contains s.
func find(out, s string) string {
	n := ""
	for _, m := range optLine.FindAllStringSubmatch(out, -1) {
		if suffix, ok := strings.CutPrefix(s, "$"); ok {
			if strings.HasSuffix(strings.TrimSpace(m[2]), suffix) {
				n = m[1]
			}
		} else if strings.Contains(m[2], s) {
			n = m[1]
		}
	}
	return n
}

// repo is a throwaway git repository for one test.
type repo struct {
	t   *testing.T
	dir string
}

// newRepo makes a repository on main with four commits, c1 to c4, each adding f<n>.txt.
func newRepo(t *testing.T) *repo {
	r := &repo{t: t, dir: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.name", "Test")
	r.git("config", "user.email", "test@example.com")
	r.git("config", "commit.gpgsign", "false")
	r.git("config", "core.autocrlf", "false")
	for i := 1; i <= 4; i++ {
		r.commit(fmt.Sprintf("c%d", i), fmt.Sprintf("f%d.txt", i), fmt.Sprintf("%d\n", i))
	}
	return r
}

// git runs git in the repository and returns its trimmed output; it fails the test on error.
func (r *repo) git(args ...string) string {
	r.t.Helper()
	c := exec.Command("git", args...)
	c.Dir = r.dir
	c.Env = append(os.Environ(), "GIT_EDITOR=true", "LC_ALL=C")
	out, err := c.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// write creates or overwrites a file in the repository, making folders as needed.
func (r *repo) write(path, content string) {
	p := filepath.Join(r.dir, path)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// read returns a file's content, or "<missing>" when it does not exist.
func (r *repo) read(path string) string {
	b, err := os.ReadFile(filepath.Join(r.dir, path))
	if err != nil {
		return "<missing>"
	}
	return string(b)
}

// commit writes files, given as path and content pairs, and commits all changes with msg.
func (r *repo) commit(msg string, files ...string) {
	for i := 0; i+1 < len(files); i += 2 {
		r.write(files[i], files[i+1])
	}
	r.git("add", "-A")
	r.git("commit", "-q", "-m", msg)
}

// subjects returns the commit subjects of HEAD's history, newest first, one per line.
func (r *repo) subjects() string { return r.git("log", "--format=%s") }

// do runs "gitwhy do <recipe>" in the repository with the given answers.
func (r *repo) do(recipe string, answers ...answer) string {
	r.t.Helper()
	return drive(r.t, r.dir, []string{"do", recipe}, answers...)
}

// expect fails the test when got is not want, showing gitwhy's output to explain why.
func expect(t *testing.T, what, got, want, out string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q\n--- output ---\n%s", what, got, want, out)
	}
}

// TestUndoLastCommit proves each choice of "undo last commit" (keep changes staged, keep them
// unstaged, delete them) removes the commit with the right status, and that "gitwhy undo"
// brings the commit and a clean tree back. This is the most common recipe.
func TestUndoLastCommit(t *testing.T) {
	for _, c := range []struct{ choice, status string }{
		{"Yes, staged", "A  f4.txt"}, {"Yes, unstaged", "?? f4.txt"}, {"No, delete", ""},
	} {
		t.Run(c.choice, func(t *testing.T) {
			r := newRepo(t)
			out := r.do("undo-last-commit", pick(c.choice), yes)
			expect(t, "subjects", r.subjects(), "c3\nc2\nc1", out)
			expect(t, "status", r.git("status", "--porcelain"), c.status, out)
			// And gitwhy undo brings it back.
			out = drive(t, r.dir, []string{"undo"}, yes)
			expect(t, "after undo", r.subjects(), "c4\nc3\nc2\nc1", out)
			expect(t, "status after undo", r.git("status", "--porcelain"), "", out)
		})
	}
}

// TestRewordCommit proves "reword commit" changes the message of the last commit and of an
// older one (through a rebase todo list) without changing any file.
func TestRewordCommit(t *testing.T) {
	r := newRepo(t)
	out := r.do("reword-commit", commit("c4"), text("head renamed"), yes)
	expect(t, "subjects", r.subjects(), "head renamed\nc3\nc2\nc1", out)
	out = r.do("reword-commit", commit("c2"), text("old renamed"), yes)
	expect(t, "subjects", r.subjects(), "head renamed\nc3\nold renamed\nc1", out)
	expect(t, "files kept", r.read("f2.txt")+r.read("f4.txt"), "2\n4\n", out)
}

// TestDropCommit proves "drop commit" removes a commit in the middle of history and its changes.
func TestDropCommit(t *testing.T) {
	r := newRepo(t)
	out := r.do("drop-commit", commit("c3"), yes)
	expect(t, "subjects", r.subjects(), "c4\nc2\nc1", out)
	expect(t, "f3 gone", r.read("f3.txt"), "<missing>", out)
}

// TestReorderAndSquash proves the rebase todo editor: moving a commit up reorders history, and
// a fixup folds a commit into the one above it, keeping both commits' files.
func TestReorderAndSquash(t *testing.T) {
	r := newRepo(t)
	// Last 3 commits, oldest first: c2, c3, c4. Move c4 up, done.
	out := r.do("edit-recent-commits", text("3"), todoItem("c4"), pick("Move up"), pick("Done"), yes)
	expect(t, "subjects", r.subjects(), "c3\nc4\nc2\nc1", out)
	// Now c2, c4, c3: fixup c3 into c4.
	out = r.do("edit-recent-commits", text("3"), todoItem("c3"), pick("Fixup"), pick("Done"), yes)
	expect(t, "subjects", r.subjects(), "c4\nc2\nc1", out)
	expect(t, "files", r.git("show", "--name-only", "--format=", "HEAD"), "f3.txt\nf4.txt", out)
}

// TestSquashLast proves "squash last commits" turns the picked commit and every commit after it
// into one commit with the new message and leaves a clean tree.
func TestSquashLast(t *testing.T) {
	r := newRepo(t)
	out := r.do("squash-last-commits", commit("c2"), text("all in one"), yes)
	expect(t, "subjects", r.subjects(), "all in one\nc1", out)
	expect(t, "status", r.git("status", "--porcelain"), "", out)
}

// TestSplitCommit proves "split commit" splits the last commit, and an older one with commits
// on top, into two commits by the files picked, keeping the later commits and a clean tree.
func TestSplitCommit(t *testing.T) {
	r := newRepo(t)
	r.commit("both", "x.txt", "x\n", "y.txt", "y\n")
	out := r.do("split-commit", commit("both"), text("first"), text("second"), yes, pick("x.txt"), text("0"))
	expect(t, "subjects", r.subjects(), "second\nfirst\nc4\nc3\nc2\nc1", out)
	expect(t, "first has", r.git("show", "--name-only", "--format=", "HEAD~1"), "x.txt", out)
	expect(t, "second has", r.git("show", "--name-only", "--format=", "HEAD"), "y.txt", out)

	// An older commit: split "first"... after adding one more on top.
	r.commit("top", "z.txt", "z\n")
	r.commit("pair", "p.txt", "p\n", "q.txt", "q\n")
	r.commit("after", "w.txt", "w\n")
	out = r.do("split-commit", commit("pair"), text("p only"), text("q only"), yes, pick("p.txt"), text("0"))
	expect(t, "subjects", strings.Join(strings.Split(r.subjects(), "\n")[:4], "\n"), "after\nq only\np only\ntop", out)
	expect(t, "clean", r.git("status", "--porcelain"), "", out)
}

// TestFixPastCommit proves "fix past commit" puts a working-tree change into an older commit
// while keeping every commit message and leaving a clean tree.
func TestFixPastCommit(t *testing.T) {
	r := newRepo(t)
	r.write("f2.txt", "2 fixed\n")
	out := r.do("fix-past-commit", pick("f2.txt"), text("0"), commit("c2"), yes)
	expect(t, "subjects", r.subjects(), "c4\nc3\nc2\nc1", out)
	expect(t, "content in c2", r.git("show", "HEAD~2:f2.txt"), "2 fixed", out)
	expect(t, "clean", r.git("status", "--porcelain"), "", out)
}

// TestMoveCommitsToNewBranch proves the picked commit and the ones after it move to a new
// branch, which is checked out, and main goes back to before them.
func TestMoveCommitsToNewBranch(t *testing.T) {
	r := newRepo(t)
	out := r.do("move-commits-new-branch", commit("c3"), text("feat"), yes, yes)
	expect(t, "branch", r.git("branch", "--show-current"), "feat", out)
	expect(t, "feat", r.subjects(), "c4\nc3\nc2\nc1", out)
	expect(t, "main", r.git("log", "--format=%s", "main"), "c2\nc1", out)
}

// TestMoveCommitsToExistingBranch proves a commit moves onto another existing branch, leaves
// main, and the user ends up back on main.
func TestMoveCommitsToExistingBranch(t *testing.T) {
	r := newRepo(t)
	r.git("branch", "other", "HEAD~2")
	out := r.do("move-commits-to-branch", commit("c4"), pick("other "), yes)
	expect(t, "main", r.git("log", "--format=%s", "main"), "c3\nc2\nc1", out)
	expect(t, "other", r.git("log", "--format=%s", "other"), "c4\nc2\nc1", out)
	expect(t, "back on main", r.git("branch", "--show-current"), "main", out)
}

// withRemote makes a bare origin and a clone. It returns the upstream repository (which has the
// bare one as origin, so it can push new commits) and the clone.
func withRemote(t *testing.T) (*repo, *repo) {
	up := newRepo(t)
	bare := t.TempDir()
	up.git("clone", "-q", "--bare", up.dir, bare)
	clone := &repo{t: t, dir: t.TempDir()}
	// CI also runs these with GITWHY_TEST_AUTOCRLF=true, Git for Windows' default.
	autocrlf := "false"
	if v := os.Getenv("GITWHY_TEST_AUTOCRLF"); v != "" {
		autocrlf = v
	}
	clone.git("clone", "-q", "-c", "core.autocrlf="+autocrlf, bare, clone.dir)
	clone.git("config", "user.name", "Test")
	clone.git("config", "user.email", "test@example.com")
	up.git("remote", "add", "origin", bare)
	return up, clone
}

// TestMakeBranchMatchRemote proves "sync with remote", choosing to delete local work, drops a
// local commit and an uncommitted edit so the branch matches the remote exactly.
func TestMakeBranchMatchRemote(t *testing.T) {
	_, c := withRemote(t)
	c.commit("local junk", "junk.txt", "j\n")
	c.write("f1.txt", "dirty\n")
	out := c.do("sync-with-remote", pick("main"), pick("Delete them"), yes)
	expect(t, "subjects", c.subjects(), "c4\nc3\nc2\nc1", out)
	expect(t, "status", c.git("status", "--porcelain"), "", out)
}

// TestResetBranchToMaster proves "reset branch to" moves the current branch to another branch's
// commit.
func TestResetBranchToMaster(t *testing.T) {
	r := newRepo(t)
	r.git("switch", "-q", "-c", "feature", "HEAD~2")
	r.commit("feature work", "feat.txt", "f\n")
	out := r.do("reset-branch-to", pick("main"), pick("Keep them"), text("n"), yes)
	expect(t, "feature", r.git("rev-parse", "feature"), r.git("rev-parse", "main"), out)
}

// TestDeleteBranch proves "delete branch", choosing "only on my computer", deletes the local
// branch.
func TestDeleteBranch(t *testing.T) {
	r := newRepo(t)
	r.git("branch", "old", "HEAD~1")
	out := r.do("delete-branch", pick("old "), pick("Only on my computer"), text("n"), yes)
	expect(t, "branches", r.git("branch", "--format=%(refname:short)"), "main", out)
}

// TestMergeConflictResolveContinue proves the whole conflict path: the merge preview predicts
// the conflict and points to "gitwhy resolve", resolve keeps the chosen side, and continue
// finishes with a merge commit.
func TestMergeConflictResolveContinue(t *testing.T) {
	r := newRepo(t)
	r.git("switch", "-q", "-c", "other")
	r.commit("theirs", "f1.txt", "theirs\n")
	r.git("switch", "-q", "main")
	r.commit("ours", "f1.txt", "ours\n")
	out := r.do("merge-branch", pick("other "), pick("Always make a merge commit"), yes)
	if !strings.Contains(out, "CONFLICTS expected") || !strings.Contains(out, "gitwhy resolve") {
		t.Fatalf("expected predicted conflict and hint:\n%s", out)
	}
	out = drive(t, r.dir, []string{"resolve"}, pick("f1.txt"), text("0"), pick("Keep theirs"), yes)
	expect(t, "resolved", r.read("f1.txt"), "theirs\n", out)
	out = drive(t, r.dir, []string{"continue"}, yes)
	expect(t, "merge commit", r.git("log", "-1", "--format=%p"), r.git("log", "-1", "--format=%p", "HEAD"), out)
	if n := len(strings.Fields(r.git("log", "-1", "--format=%p"))); n != 2 {
		t.Fatalf("HEAD has %d parents, want a merge\n%s", n, out)
	}
}

// TestAbortRebase proves "gitwhy abort" cancels a rebase stopped on a conflict that was started
// in plain git, and puts HEAD back where it was.
func TestAbortRebase(t *testing.T) {
	r := newRepo(t)
	r.git("switch", "-q", "-c", "other", "HEAD~1")
	r.commit("conflicting", "f4.txt", "other\n")
	before := r.git("rev-parse", "HEAD")
	c := exec.Command("git", "rebase", "main")
	c.Dir = r.dir
	c.Run() // conflicts
	out := drive(t, r.dir, []string{"abort"}, yes)
	expect(t, "HEAD", r.git("rev-parse", "HEAD"), before, out)
}

// TestStashRoundTrip proves stashing saves a tracked edit and an untracked file, and popping the
// stash brings both back.
func TestStashRoundTrip(t *testing.T) {
	r := newRepo(t)
	r.write("f1.txt", "wip\n")
	r.write("new.txt", "n\n")
	out := r.do("stash-save", text("my wip"), yes, yes)
	expect(t, "clean", r.git("status", "--porcelain"), "", out)
	out = r.do("stash-restore", pick("pop"), text("0"), yes)
	expect(t, "back", r.read("f1.txt")+r.read("new.txt"), "wip\nn\n", out)
}

// TestRemoveFileFromHistory proves the git-filter-repo recipe removes a file from every commit
// and from the working tree. It is skipped when git-filter-repo is not installed.
func TestRemoveFileFromHistory(t *testing.T) {
	if exec.Command("git", "filter-repo", "--version").Run() != nil {
		t.Skip("git-filter-repo not installed")
	}
	r := newRepo(t)
	out := r.do("remove-file-from-history", text("f2.txt"), yes)
	expect(t, "f2 in history", r.git("log", "--branches", "--tags", "--format=%h", "--", "f2.txt"), "", out)
	expect(t, "f2 file", r.read("f2.txt"), "<missing>", out)
}

// TestLFSTrack proves the LFS recipe writes the right .gitattributes line. It is skipped when
// git-lfs is not installed.
func TestLFSTrack(t *testing.T) {
	if exec.Command("git", "lfs", "version").Run() != nil {
		t.Skip("git-lfs not installed")
	}
	r := newRepo(t)
	out := r.do("lfs-track", text("*.bin"), yes)
	// git-lfs writes CRLF on Windows.
	expect(t, ".gitattributes", strings.ReplaceAll(r.read(".gitattributes"), "\r\n", "\n"), "*.bin filter=lfs diff=lfs merge=lfs -text\n", out)
}

// TestRunWithPreviewAndUndo proves "gitwhy run git clean -fd" shows what it will delete, deletes
// the untracked file, and that undo brings the file back from the snapshot.
func TestRunWithPreviewAndUndo(t *testing.T) {
	r := newRepo(t)
	r.write("u.txt", "untracked\n")
	out := drive(t, r.dir, []string{"run", "git", "clean", "-fd"}, yes)
	if !strings.Contains(out, "Would remove") && !strings.Contains(out, "u.txt") {
		t.Fatalf("no preview:\n%s", out)
	}
	expect(t, "deleted", r.read("u.txt"), "<missing>", out)
	out = drive(t, r.dir, []string{"undo"}, yes)
	expect(t, "restored", r.read("u.txt"), "untracked\n", out)
}

// TestRecoverPlainGitPull proves "gitwhy recover" finds a pull done in plain git (no snapshot)
// in the reflog and goes back to before it.
func TestRecoverPlainGitPull(t *testing.T) {
	up, c := withRemote(t)
	up.commit("upstream 1", "u1.txt", "1\n")
	up.commit("upstream 2", "u2.txt", "2\n")
	up.git("push", "-q", "origin", "main")
	c.git("pull", "-q", "--no-rebase")
	out := drive(t, c.dir, []string{"recover"}, pick("Pulled"), yes)
	expect(t, "after recover", c.subjects(), "c4\nc3\nc2\nc1", out)
}

// TestExplainAndSearch proves "gitwhy explain" describes each part of a command, and that unknown words ("gitwhy undo last commit") are a search whose result runs.
func TestExplainAndSearch(t *testing.T) {
	r := newRepo(t)
	out := drive(t, r.dir, []string{"explain", "git", "reset", "--hard", "HEAD~1"})
	for _, want := range []string{"--hard", "1 commit(s) before", "Deletes every uncommitted change"} {
		if !strings.Contains(out, want) {
			t.Fatalf("explain missing %q:\n%s", want, out)
		}
	}
	out = drive(t, r.dir, []string{"undo", "last", "commit"}, pick("Undo last commit"), pick("Yes, staged"), yes)
	expect(t, "searched and ran", r.subjects(), "c3\nc2\nc1", out)
}

// TestBuilder proves the command builder: picked flags are added, a long flag's value is joined
// with "=", the command is shown quoted for copy-paste, and it runs.
func TestBuilder(t *testing.T) {
	r := newRepo(t)
	// git commit, flags: --allow-empty and --message (asks a value), nothing else.
	out := drive(t, r.dir, []string{"build", "commit"}, pick("allow-empty "), pick("-m <msg>"), text("0"), text("built it"), text(""), yes)
	expect(t, "subject", r.git("log", "-1", "--format=%s"), "built it", out)
	if !strings.Contains(out, "git commit --allow-empty '--message=built it'") {
		t.Fatalf("command not shown:\n%s", out)
	}
}

// TestRemoveFileFromHistoryThenUndo proves undo reverses a full history rewrite by
// git-filter-repo: the old HEAD and the removed file come back.
func TestRemoveFileFromHistoryThenUndo(t *testing.T) {
	if exec.Command("git", "filter-repo", "--version").Run() != nil {
		t.Skip("git-filter-repo not installed")
	}
	r := newRepo(t)
	before := r.git("rev-parse", "HEAD")
	out := r.do("remove-file-from-history", text("f2.txt"), yes)
	expect(t, "rewritten", r.git("log", "--branches", "--tags", "--format=%h", "--", "f2.txt"), "", out)
	out = drive(t, r.dir, []string{"undo"}, yes)
	expect(t, "undone", r.git("rev-parse", "HEAD"), before, out)
	expect(t, "f2 back", r.read("f2.txt"), "2\n", out)
}

// TestDeleteBranchUsesItsUpstream is a regression test: deleting a branch on the remote must
// delete the branch it tracks (team/fix), not a different remote branch with the same short
// name (fix).
func TestDeleteBranchUsesItsUpstream(t *testing.T) {
	_, c := withRemote(t)
	c.git("push", "-q", "origin", "HEAD:refs/heads/team/fix")
	c.git("push", "-q", "origin", "HEAD:refs/heads/fix") // someone else's branch with the same short name
	c.git("branch", "--track", "fix", "origin/team/fix")
	out := c.do("delete-branch", pick("fix "), pick("On my computer and on the remote"), text("y"), yes)
	remote := c.git("ls-remote", "--heads", "origin")
	if strings.Contains(remote, "refs/heads/team/fix") || !strings.Contains(remote, "refs/heads/fix") {
		t.Fatalf("wrong remote branch deleted:\n%s\n%s", remote, out)
	}
}

// TestRunFromSubfolder proves a typed command runs in the folder gitwhy was started from, as in
// plain git: "restore ." in a subfolder must not touch files at the root.
func TestRunFromSubfolder(t *testing.T) {
	r := newRepo(t)
	r.commit("sub", "sub/s.txt", "s\n")
	r.write("f1.txt", "root edit\n")
	r.write("sub/s.txt", "sub edit\n")
	out := drive(t, filepath.Join(r.dir, "sub"), []string{"run", "restore", "."}, yes)
	expect(t, "sub restored", r.read("sub/s.txt"), "s\n", out)
	expect(t, "root kept", r.read("f1.txt"), "root edit\n", out)
}

// TestRunRefusesGlobalOptions proves "gitwhy run" refuses git's global options such as -C,
// which would run the command in another repository than the one the snapshot is taken of.
func TestRunRefusesGlobalOptions(t *testing.T) {
	r := newRepo(t)
	out := drive(t, r.dir, []string{"run", "-C", "..", "reset", "--hard"})
	if !strings.Contains(out, "global options") {
		t.Fatalf("expected refusal:\n%s", out)
	}
}

// TestFixupLeavesOtherStagedChanges is a regression test: "fix past commit" must commit only
// the picked file, not other changes that were already staged, and must keep those changes.
func TestFixupLeavesOtherStagedChanges(t *testing.T) {
	r := newRepo(t)
	r.write("f1.txt", "unrelated staged\n")
	r.git("add", "f1.txt")
	r.write("f2.txt", "2 fixed\n")
	out := r.do("fix-past-commit", pick("f2.txt"), text("0"), commit("c2"), yes)
	expect(t, "c2 got fix", r.git("show", "HEAD~2:f2.txt"), "2 fixed", out)
	expect(t, "c1 untouched in history", r.git("show", "HEAD:f1.txt"), "1", out)
	// --autostash brings the unrelated change back unstaged; it must not be lost or committed.
	expect(t, "unrelated kept", r.read("f1.txt"), "unrelated staged\n", out)
}

// TestPushedCommitWarning proves rewording a commit that is already on the remote warns about
// it, and that answering No at the warning runs nothing.
func TestPushedCommitWarning(t *testing.T) {
	_, c := withRemote(t)
	c.commit("local only", "l.txt", "l\n")
	out := c.do("reword-commit", commit("c2"), text("renamed"), text("n"))
	if !strings.Contains(out, "already on the remote") {
		t.Fatalf("no pushed warning:\n%s", out)
	}
	expect(t, "nothing ran", c.git("log", "-1", "--format=%s", "HEAD~2"), "c3", out)
}

// TestNewRecipes runs four smaller recipes end to end: create a tag on an older commit, restore
// a file from a commit, stop tracking a file but keep it on disk, and change a commit's author.
func TestNewRecipes(t *testing.T) {
	r := newRepo(t)
	out := r.do("create-tag", text("v1.0.0"), commit("c3"), text("first release"), text("n"), yes)
	expect(t, "tag", r.git("rev-parse", "v1.0.0^{commit}"), r.git("rev-parse", "HEAD~1"), out)

	r.write("f2.txt", "changed\n")
	out = r.do("restore-file-from-commit", pick("f1.txt"), commit("c1"), yes)
	expect(t, "f1 from c1", r.read("f1.txt"), "1\n", out)

	out = r.do("untrack-file", pick("f4.txt"), text("0"), yes)
	expect(t, "untracked", r.git("status", "--porcelain", "--", "f4.txt"), "D  f4.txt\n?? f4.txt", out)
	expect(t, "kept on disk", r.read("f4.txt"), "4\n", out)

	out = r.do("change-commit-author", pick("Someone else"), text("Ada <ada@example.com>"), yes)
	expect(t, "author", r.git("log", "-1", "--format=%an <%ae>"), "Ada <ada@example.com>", out)
}

// TestBisectRecipe proves the bisect recipe finds the first bad commit with a test command and
// ends the bisect, so the user is back on their branch.
func TestBisectRecipe(t *testing.T) {
	r := newRepo(t) // f3.txt appears in c3: "test" fails once it exists
	out := r.do("find-bug-commit", commit("c1"), text("test ! -e f3.txt"), yes)
	if !regexp.MustCompile(`is the first '?bad'? commit`).MatchString(out) || !strings.Contains(out, "c3") {
		t.Fatalf("bisect didn't find c3:\n%s", out)
	}
	expect(t, "back on main", r.git("branch", "--show-current"), "main", out)
}

// TestHookUndoesPlainGit proves the bash hook takes a snapshot before a plain "git reset --hard",
// so "gitwhy undo" brings the lost edit back. It is skipped when bash is missing.
func TestHookUndoesPlainGit(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("no bash")
	}
	r := newRepo(t)
	script, err := exec.Command(bin, "hook", "init", "bash").Output()
	if err != nil {
		t.Fatal(err)
	}
	c := exec.Command(bash, "-c", string(script)+fmt.Sprintf("\ncd %q && echo dirty > f1.txt && git reset -q --hard", r.dir))
	c.Env = append(env(t.TempDir()), "PATH="+filepath.Dir(bin)+string(os.PathListSeparator)+os.Getenv("PATH"))
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	expect(t, "after reset", r.read("f1.txt"), "1\n", "")
	out := drive(t, r.dir, []string{"undo"}, yes)
	expect(t, "after undo", r.read("f1.txt"), "dirty\n", out)
}
