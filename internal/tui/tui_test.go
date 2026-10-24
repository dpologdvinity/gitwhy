package tui

import (
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

func TestLoadAndSuggest(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Commit("two", "b.txt", "2\n")
	r.Write("a.txt", "edited\n")
	r.Write("new.txt", "n\n")
	r.Write("b.txt", "staged\n")
	r.Git("add", "b.txt")

	m := &model{repo: r.Repo, tab: tabCommits}
	m.load()
	if m.head != "main" || m.count(tabCommits) != 2 || len(m.brs) != 1 {
		t.Fatalf("head=%q commits=%d branches=%d", m.head, m.count(tabCommits), len(m.brs))
	}
	got := map[string]string{}
	for _, f := range m.files {
		got[f.path] = f.label()
	}
	want := map[string]string{"a.txt": "changed", "b.txt": "staged", "new.txt": "new"}
	for p, l := range want {
		if got[p] != l {
			t.Errorf("%s = %q, want %q (all: %v)", p, got[p], l, got)
		}
	}
	if s := strings.Join(m.suggest(), "|"); !strings.Contains(s, "commit 1 staged") || !strings.Contains(s, "publish") {
		t.Errorf("suggestions = %q", s)
	}
}

func TestConflictAndOperationShown(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "base\n")
	r.Git("switch", "-q", "-c", "other")
	r.Commit("theirs", "f.txt", "theirs\n")
	r.Git("switch", "-q", "main")
	r.Commit("ours", "f.txt", "ours\n")
	_, _ = r.Output("merge", "other") // conflicts

	m := &model{repo: r.Repo}
	m.load()
	if m.op != "merge" || len(m.files) != 1 || !m.files[0].conflict() {
		t.Fatalf("op=%q files=%+v", m.op, m.files)
	}
	a, ok := m.tabKey("enter", 0)
	if !ok || a.Recipe != "resolve-conflicts" || a.Preset["files"].([]string)[0] != "f.txt" {
		t.Fatalf("enter on conflict = %+v", a)
	}
	if s := strings.Join(m.suggest(), "|"); !strings.Contains(s, "1 conflict") {
		t.Errorf("suggestions = %q", s)
	}
}

// TestCommitKeyCommitsTheIndex proves that c with staged files commits the index as it is
// (commit-staged, no file list), so a partly staged file or a staged rename is committed exactly
// as staged; with nothing staged it still offers commit-changes.
func TestCommitKeyCommitsTheIndex(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "f.txt", "1\n2\n", "old.txt", "o\n")
	r.Write("f.txt", "1 edited\n2\n")
	m := &model{repo: r.Repo}
	m.load()
	if a, ok := m.tabKey("c", 0); !ok || a.Recipe != "commit-changes" || a.Preset != nil {
		t.Fatalf("c with nothing staged = %+v", a)
	}
	r.Git("mv", "old.txt", "new.txt")
	m.load()
	a, ok := m.tabKey("c", 0)
	if !ok || a.Kind != "recipe" || a.Recipe != "commit-staged" || a.Preset != nil {
		t.Fatalf("c with a staged rename = %+v", a)
	}
}

func TestCommitKeysPresetCommit(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	m := &model{repo: r.Repo, tab: tabCommits}
	m.load()
	a, ok := m.tabKey("r", 0)
	if !ok || a.Recipe != "reword-commit" || a.Preset["commit"] != m.commitAt(0).ID {
		t.Fatalf("r = %+v", a)
	}
}
