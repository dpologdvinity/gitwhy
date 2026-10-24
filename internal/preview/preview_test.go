package preview

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

func text(r Result) string { return strings.Join(r.Lines, "\n") }

func diverged(t *testing.T, conflict bool) *testrepo.Repo {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "base\n", "g.txt", "g\n")
	r.Git("switch", "-q", "-c", "feature")
	r.Commit("feature work", "f.txt", "feature\n")
	r.Git("switch", "-q", "main")
	if conflict {
		r.Commit("main work", "f.txt", "main\n")
	} else {
		r.Commit("main work", "g.txt", "g2\n")
	}
	return r
}

func TestResetPreview(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Commit("two", "a.txt", "2\n")
	r.Write("a.txt", "dirty\n")
	got := Preview(r.Repo, []string{"reset", "--hard", "HEAD~1"})
	s := text(got)
	if !got.Exact || !strings.Contains(s, "Commits leaving main (1)") || !strings.Contains(s, "two") || !strings.Contains(s, "RESET (1)") {
		t.Fatalf("unexpected preview:\n%s", s)
	}
	if r.Git("rev-parse", "HEAD~0") == "" || r.Read("a.txt") != "dirty\n" {
		t.Fatal("preview changed the repo")
	}
}

func TestCleanPreviewSavesFiles(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", ".gitignore", "*.log\n")
	r.Write("x.log", "x\n")
	r.Write("u.txt", "u\n")
	got := Preview(r.Repo, []string{"clean", "-fdx"})
	if !got.Exact || len(got.Save) != 2 {
		t.Fatalf("got %+v", got)
	}
	if r.Read("x.log") != "x\n" || r.Read("u.txt") != "u\n" {
		t.Fatal("preview deleted files")
	}
}

func TestMergePredictsConflicts(t *testing.T) {
	r := diverged(t, true)
	if s := text(Preview(r.Repo, []string{"merge", "feature"})); !strings.Contains(s, "CONFLICTS expected in (1)") || !strings.Contains(s, "f.txt") {
		t.Fatalf("merge preview:\n%s", s)
	}
	r2 := diverged(t, false)
	if s := text(Preview(r2.Repo, []string{"merge", "feature"})); !strings.Contains(s, "No conflicts") {
		t.Fatalf("clean merge preview:\n%s", s)
	}
}

func TestRebaseAndCherryPickPredictConflicts(t *testing.T) {
	r := diverged(t, true)
	r.Git("switch", "-q", "feature")
	if s := text(Preview(r.Repo, []string{"rebase", "main"})); !strings.Contains(s, "CONFLICT expected at") {
		t.Fatalf("rebase preview:\n%s", s)
	}
	r2 := diverged(t, false)
	r2.Git("switch", "-q", "feature")
	if s := text(Preview(r2.Repo, []string{"rebase", "main"})); !strings.Contains(s, "No conflicts expected") {
		t.Fatalf("clean rebase preview:\n%s", s)
	}
	if s := text(Preview(r2.Repo, []string{"cherry-pick", "main"})); !strings.Contains(s, "No conflicts expected") {
		t.Fatalf("cherry-pick preview:\n%s", s)
	}
}

func TestBranchDeleteShowsOrphanedCommits(t *testing.T) {
	r := diverged(t, false)
	s := text(Preview(r.Repo, []string{"branch", "-D", "feature"}))
	if !strings.Contains(s, "NO branch") || !strings.Contains(s, "(1):") || !strings.Contains(s, "feature work") {
		t.Fatalf("branch preview:\n%s", s)
	}
}

// inDir returns a copy of repo that runs git from the given folder under the repo root, as
// `gitwhy run` does when the user is inside a subfolder.
func inDir(r *testrepo.Repo, sub string) *git.Repo {
	c := *r.Repo
	c.Dir = filepath.Join(r.Top, sub)
	return &c
}

// TestCleanPreviewSaveIsRelativeToTheRepoRoot checks that the files a clean will delete are
// listed from the repo root even when git runs in a subfolder. snapshot.Take reads them from the
// root: a folder-relative "debug.log" would save the top-level file instead of sub/debug.log.
func TestCleanPreviewSaveIsRelativeToTheRepoRoot(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", ".gitignore", "*.log\n", "sub/keep.txt", "k\n")
	r.Write("sub/debug.log", "precious\n")
	r.Write("debug.log", "top\n")
	got := Preview(inDir(r, "sub"), []string{"clean", "-fdx"})
	if len(got.Save) != 1 || got.Save[0] != "sub/debug.log" {
		t.Fatalf("Save = %q, want [sub/debug.log]", got.Save)
	}
}

// TestCleanPreviewSaveHandlesQuotedNames checks that names git writes in quotes or escapes
// (non-ASCII, a double quote) come back as the real file names, so they can be saved.
func TestCleanPreviewSaveHandlesQuotedNames(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", ".gitignore", "*.log\n")
	names := []string{"secret-é.log", `q"uote.log`, "sp ace.log"}
	for _, n := range names {
		r.Write(n, "x\n")
	}
	got := Preview(r.Repo, []string{"clean", "-fX"})
	have := map[string]bool{}
	for _, s := range got.Save {
		have[s] = true
	}
	for _, n := range names {
		if !have[n] {
			t.Errorf("Save %q is missing %q", got.Save, n)
		}
	}
}

// TestCleanPreviewDoubleForceListsNestedRepos checks that `clean -ffd` is previewed with its
// second -f: only then does git delete a nested repository, so dropping the flag showed
// "Nothing to do" for a command that deletes a whole repo.
func TestCleanPreviewDoubleForceListsNestedRepos(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "a.txt", "a\n")
	r.Git("init", "-q", "nested")
	r.Write("nested/f.txt", "x\n")
	if s := text(Preview(r.Repo, []string{"clean", "-ffd"})); !strings.Contains(s, "nested/") {
		t.Fatalf("preview should list nested/:\n%s", s)
	}
}

// TestCleanPreviewKeepsAttachedValues checks that a pattern attached to -e is passed to git
// untouched (it used to lose every "f" in it).
func TestCleanPreviewKeepsAttachedValues(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "a.txt", "a\n")
	r.Write("keepfile.txt", "x\n")
	r.Write("other.txt", "x\n")
	got := Preview(r.Repo, []string{"clean", "-fd", "-ekeepfile.txt"})
	if s := text(got); strings.Contains(s, "keepfile.txt") || !strings.Contains(s, "other.txt") {
		t.Fatalf("pattern keepfile.txt should be excluded:\n%s", s)
	}
}

// TestSwitchForcePreviewSaysChangesAreDiscarded checks that `checkout -f` and
// `switch --discard-changes` do not claim your uncommitted changes "come along".
func TestSwitchForcePreviewSaysChangesAreDiscarded(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Git("branch", "other")
	r.Write("f.txt", "dirty\n")
	for _, args := range [][]string{{"checkout", "-f", "other"}, {"switch", "--discard-changes", "other"}} {
		s := text(Preview(r.Repo, args))
		if strings.Contains(s, "come along") || !strings.Contains(s, "DISCARDED") || !strings.Contains(s, "f.txt") {
			t.Errorf("%v:\n%s", args, s)
		}
	}
}

// TestMergePreviewHonoursFastForwardFlags checks the merge preview against what git really does
// with --no-ff, --ff-only and --no-commit.
func TestMergePreviewHonoursFastForwardFlags(t *testing.T) {
	ff := testrepo.New(t)
	ff.Commit("base", "f.txt", "1\n")
	ff.Git("switch", "-q", "-c", "ahead")
	ff.Commit("ahead work", "g.txt", "g\n")
	ff.Git("switch", "-q", "main")
	if s := text(Preview(ff.Repo, []string{"merge", "--no-ff", "ahead"})); strings.Contains(s, "Fast-forward") || !strings.Contains(s, "merge commit") {
		t.Errorf("--no-ff on a fast-forwardable branch:\n%s", s)
	}
	if s := text(Preview(ff.Repo, []string{"merge", "ahead"})); !strings.Contains(s, "Fast-forward") {
		t.Errorf("plain merge should fast-forward:\n%s", s)
	}

	div := diverged(t, false)
	if s := text(Preview(div.Repo, []string{"merge", "--ff-only", "feature"})); !strings.Contains(s, "fast-forward") || strings.Contains(s, "A merge commit is created") {
		t.Errorf("--ff-only on diverged branches should say git refuses:\n%s", s)
	}
	if s := text(Preview(div.Repo, []string{"merge", "--no-commit", "feature"})); !strings.Contains(s, "not committed") {
		t.Errorf("--no-commit should say the merge is staged, not committed:\n%s", s)
	}
}

// TestMvPreviewListsOnlyTheMoves checks that git's progress lines ("Checking rename of ...") are
// not shown as files.
func TestMvPreviewListsOnlyTheMoves(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "h.txt", "h\n")
	got := Preview(r.Repo, []string{"mv", "h.txt", "h2.txt"})
	s := text(got)
	if strings.Contains(s, "Checking") || !strings.Contains(s, "h.txt to h2.txt") || !strings.Contains(s, "(1)") {
		t.Fatalf("unexpected mv preview:\n%s", s)
	}
}

// TestCheckoutWithEmptyDashDashIsABranchSwitch checks that `checkout <branch> --` is previewed as
// a branch switch, not as overwriting an empty list of files.
func TestCheckoutWithEmptyDashDashIsABranchSwitch(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Git("branch", "other")
	if s := text(Preview(r.Repo, []string{"checkout", "other", "--"})); strings.Contains(s, "overwritten") || !strings.Contains(s, "Your files change to") {
		t.Fatalf("empty -- is not a file list; expected the branch-switch preview:\n%s", s)
	}
}
