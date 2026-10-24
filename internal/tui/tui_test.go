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

// TestSpaceKeepsCursorOnFile is a regression test: staging a new file moves it up the list
// (porcelain v2 lists untracked files last), and the cursor must follow the file, not stay on
// its old row, so a second space unstages the same file.
func TestSpaceKeepsCursorOnFile(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "z.txt", "z\n")
	r.Write("z.txt", "edited\n")
	r.Write("a.txt", "a\n")
	r.Write("b.txt", "b\n")
	m := &model{repo: r.Repo}
	m.load()
	m.key("down")
	if m.files[m.cursor[tabChanges]].path != "a.txt" {
		t.Fatalf("setup: cursor on %q, files %+v", m.files[m.cursor[tabChanges]].path, m.files)
	}
	m.key("space")
	if f := m.files[m.cursor[tabChanges]]; f.path != "a.txt" || !f.staged() {
		t.Fatalf("after staging, cursor on %+v (files %+v)", f, m.files)
	}
	m.key("space")
	if st := r.Git("status", "--porcelain"); !strings.Contains(st, " M z.txt") || !strings.Contains(st, "?? a.txt") {
		t.Fatalf("second space touched the wrong file:\n%s", st)
	}
}

// TestReloadKeepsCursorOnBranch proves that a reload that reorders the Branches tab (it is
// sorted by commit date) keeps the cursor on the same branch.
func TestReloadKeepsCursorOnBranch(t *testing.T) {
	r := testrepo.New(t)
	at := func(date, msg string) {
		r.Git("add", "-A")
		if _, err := r.With("GIT_COMMITTER_DATE="+date).Output("commit", "-q", "--allow-empty", "-m", msg); err != nil {
			t.Fatal(err)
		}
	}
	at("2000-01-01T00:00:00Z", "main one")
	r.Git("switch", "-q", "-c", "feat")
	at("2001-01-01T00:00:00Z", "feat one")
	r.Git("switch", "-q", "main")
	m := &model{repo: r.Repo, tab: tabBranches}
	m.load()
	m.key("down")
	if b := m.brs[m.cursor[tabBranches]].name; b != "main" {
		t.Fatalf("setup: cursor on %q (%+v)", b, m.brs)
	}
	at("2002-01-01T00:00:00Z", "main two") // main is now the newest: the list reorders
	m.key("ctrl+r")
	if b := m.brs[m.cursor[tabBranches]].name; b != "main" {
		t.Fatalf("after reload cursor on %q (%+v)", b, m.brs)
	}
}

// TestUnstageInRepoWithoutCommits is a regression test: space on a staged file must unstage it
// even before the first commit, where "git restore --staged" fails for lack of HEAD.
func TestUnstageInRepoWithoutCommits(t *testing.T) {
	r := testrepo.New(t)
	r.Write("a.txt", "a\n")
	r.Git("add", "a.txt")
	m := &model{repo: r.Repo}
	m.load()
	m.key("space")
	if len(m.files) != 1 || !m.files[0].untracked() {
		t.Fatalf("file still staged: %+v (msg %q)", m.files, m.msg)
	}
}

// TestDeleteUntrackedFilePresetsIt proves that d on one untracked file deletes only that file
// (delete-untracked with a files preset), not every untracked file.
func TestDeleteUntrackedFilePresetsIt(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Write("important-notes.txt", "keep\n")
	r.Write("scratch.txt", "x\n")
	m := &model{repo: r.Repo}
	m.load()
	m.key("down")
	a, ok := m.tabKey("d", m.cursor[tabChanges])
	files, _ := a.Preset["files"].([]string)
	if !ok || a.Recipe != "delete-untracked" || len(files) != 1 || files[0] != "scratch.txt" {
		t.Fatalf("d on scratch.txt = %+v", a)
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
