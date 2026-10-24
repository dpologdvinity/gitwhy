package tui

import (
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestGraphActionsOutsideHistory proves that history-rewriting keys refuse a commit that is not
// on HEAD's branch, while p still offers to copy it here.
func TestGraphActionsOutsideHistory(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "a.txt", "a\n")
	r.Git("switch", "-q", "-c", "other")
	r.Commit("elsewhere", "b.txt", "b\n")
	r.Git("switch", "-q", "main")
	r.Commit("here", "c.txt", "c\n")
	m := &model{repo: r.Repo, tab: tabCommits, all: true}
	m.load()
	k := -1
	for i := 0; i < m.count(tabCommits); i++ {
		if m.commitAt(i).Subject == "elsewhere" {
			k = i
		}
	}
	if k < 0 {
		t.Fatal("commit from other branch not shown with all")
	}
	m.cursor[tabCommits] = k
	if _, ok := m.tabKey("r", k); ok || !strings.Contains(m.msg, "isn't on your branch") {
		t.Fatalf("r: ok=%v msg=%q", ok, m.msg)
	}
	a, ok := m.tabKey("p", k)
	if !ok || a.Recipe != "copy-commit" || a.Preset["commit"] != m.commitAt(k).ID {
		t.Fatalf("p = %+v", a)
	}
}

// TestGraphCursorSkipsConnectors proves that the Commits cursor counts only commit rows, not
// the graph's connector rows.
func TestGraphCursorSkipsConnectors(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "a.txt", "a\n")
	r.Git("switch", "-q", "-c", "side")
	r.Commit("side", "b.txt", "b\n")
	r.Git("switch", "-q", "main")
	r.Commit("main", "c.txt", "c\n")
	r.Git("merge", "-q", "--no-ff", "-m", "merge", "side")
	m := &model{repo: r.Repo, tab: tabCommits}
	m.load()
	commits := 0
	for _, row := range m.rows {
		if row.Commit != nil {
			commits++
		}
	}
	if n := m.count(tabCommits); n != commits || n == len(m.rows) {
		t.Fatalf("count=%d commits=%d rows=%d", n, commits, len(m.rows))
	}
	for k := 0; k < commits; k++ {
		if m.commitAt(k) == nil {
			t.Fatalf("commitAt(%d) = nil", k)
		}
	}
}

// TestCopyKeyRefusedForCommitOnBranch proves that p refuses to copy a commit that is already
// on the current branch.
func TestCopyKeyRefusedForCommitOnBranch(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	m := &model{repo: r.Repo, tab: tabCommits}
	m.load()
	if _, ok := m.tabKey("p", 0); ok || !strings.Contains(m.msg, "already on your branch") {
		t.Fatalf("ok=%v msg=%q", ok, m.msg)
	}
}

// TestAllToggleKeepsSelectedCommit proves that a (show all branches) keeps the same commit
// selected though the graph changes.
func TestAllToggleKeepsSelectedCommit(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "a.txt", "a\n")
	r.Git("switch", "-q", "-c", "other")
	r.Commit("elsewhere", "b.txt", "b\n")
	r.Git("switch", "-q", "main")
	r.Commit("here", "c.txt", "c\n")
	m := &model{repo: r.Repo, tab: tabCommits}
	m.load()
	m.cursor[tabCommits] = 1 // "base"
	want := m.commitAt(1).ID
	m.tabKey("a", 1)
	if got := m.commitAt(m.cursor[tabCommits]); got == nil || got.ID != want {
		t.Fatalf("selection moved: %+v", got)
	}
}
