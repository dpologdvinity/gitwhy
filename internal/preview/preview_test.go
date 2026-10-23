package preview

import (
	"strings"
	"testing"

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
