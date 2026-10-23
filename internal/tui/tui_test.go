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

	m := &model{repo: r.Repo}
	m.load()
	if m.head != "main" || len(m.commits) != 2 || len(m.brs) != 1 {
		t.Fatalf("head=%q commits=%d branches=%d", m.head, len(m.commits), len(m.brs))
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

func TestCommitKeysPresetCommit(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	m := &model{repo: r.Repo, tab: tabCommits}
	m.load()
	a, ok := m.tabKey("r", 0)
	if !ok || a.Recipe != "reword-commit" || a.Preset["commit"] != m.commits[0].id {
		t.Fatalf("r = %+v", a)
	}
}
