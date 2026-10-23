package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

func numbered(n int, edit map[int]string) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		if s, ok := edit[i]; ok {
			b.WriteString(s)
		} else {
			fmt.Fprintf(&b, "line %d\n", i)
		}
	}
	return b.String()
}

// twoHunks commits a 20-line f.txt, edits lines 2 and 18, and loads the TUI.
func twoHunks(t *testing.T) (*testrepo.Repo, *model) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", numbered(20, nil))
	r.Write("f.txt", numbered(20, map[int]string{2: "two\n", 18: "eighteen\n"}))
	m := &model{repo: r.Repo}
	m.load()
	return r, m
}

func TestHunkModeStagesOneHunk(t *testing.T) {
	r, m := twoHunks(t)
	m.key("enter")
	if m.hunks == nil || len(m.hunks.fd.Hunks) != 2 {
		t.Fatalf("hunks = %+v (msg %q)", m.hunks, m.msg)
	}
	m.key("space")
	if !strings.Contains(r.Git("diff", "--cached", "--stat"), "f.txt") {
		t.Fatal("nothing staged")
	}
	if d := r.Git("diff", "-U0"); !strings.Contains(d, "+eighteen") || strings.Contains(d, "+two") {
		t.Fatalf("unstaged:\n%s", d)
	}
	if m.hunks == nil || len(m.hunks.fd.Hunks) != 1 {
		t.Fatalf("hunks after stage = %+v", m.hunks)
	}
}

func TestHunkModeUnstageSide(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", numbered(20, nil))
	r.Write("f.txt", numbered(20, map[int]string{2: "two\n"}))
	r.Git("add", "f.txt")
	m := &model{repo: r.Repo}
	m.load()
	m.key("enter")
	if m.hunks == nil || !m.hunks.staged {
		t.Fatalf("hunks = %+v", m.hunks)
	}
	m.key("space")
	if d := r.Git("diff", "--cached"); d != "" {
		t.Fatalf("still staged:\n%s", d)
	}
	if m.hunks != nil {
		t.Fatal("should be back on the file list")
	}
}

func TestHunkModeUntrackedFile(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "a.txt", "a\n")
	r.Write("new.txt", "n\n")
	m := &model{repo: r.Repo}
	m.load()
	m.key("enter")
	if m.hunks != nil || !strings.Contains(m.msg, "New file") {
		t.Fatalf("hunks=%v msg=%q", m.hunks, m.msg)
	}
}

func TestHunkModeQGoesBack(t *testing.T) {
	_, m := twoHunks(t)
	m.key("enter")
	if m.hunks == nil {
		t.Fatal("not in hunk mode")
	}
	m.key("q")
	if m.action.Kind == "quit" || m.hunks != nil {
		t.Fatalf("action=%+v hunks=%v", m.action, m.hunks)
	}
}

func TestLineModeStagesPickedLine(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", numbered(10, nil))
	r.Write("f.txt", numbered(10, map[int]string{5: "line 5\na\nb\n"}))
	m := &model{repo: r.Repo}
	m.load()
	m.key("enter")
	m.key("v")
	m.key("enter")
	if !strings.Contains(m.msg, "Pick lines") {
		t.Fatalf("msg = %q", m.msg)
	}
	m.key("space")
	m.key("enter")
	if idx := r.Git("show", ":f.txt"); !strings.Contains(idx, "\na\n") || strings.Contains(idx, "\nb\n") {
		t.Fatalf("index:\n%s", idx)
	}
	if d := r.Git("diff", "-U0"); !strings.Contains(d, "+b") || strings.Contains(d, "+a") {
		t.Fatalf("unstaged:\n%s", d)
	}
}

func TestLineModeRefusesNoNewlineAtEOF(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "a\nb")
	r.Write("f.txt", "a\nc")
	m := &model{repo: r.Repo}
	m.load()
	m.key("enter")
	m.key("v")
	if m.hunks == nil || m.hunks.lineMode || !strings.Contains(m.msg, "newline") {
		t.Fatalf("lineMode=%v msg=%q", m.hunks != nil && m.hunks.lineMode, m.msg)
	}
}

func TestHunkModeIgnoresUserDiffConfig(t *testing.T) {
	for _, cfg := range [][2]string{{"diff.noprefix", "true"}, {"color.diff", "always"}, {"diff.suppressBlankEmpty", "true"}, {"diff.external", "false"}} {
		t.Run(cfg[0], func(t *testing.T) {
			r := testrepo.New(t)
			r.Commit("base", "f.txt", "1\n\n3\n"+numbered(20, nil))
			r.Git("config", cfg[0], cfg[1])
			r.Write("f.txt", "1\n\n33\n"+numbered(20, nil))
			m := &model{repo: r.Repo}
			m.load()
			m.key("enter")
			if m.hunks == nil {
				t.Fatalf("no hunks: %q", m.msg)
			}
			m.key("space")
			if !strings.Contains(r.Git("diff", "--cached", "--no-color", "--no-ext-diff"), "+33") {
				t.Fatalf("not staged (msg %q)", m.msg)
			}
		})
	}
}

func TestGraphNotLoadedOffCommitsTab(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	m := &model{repo: r.Repo}
	m.load()
	if m.rows != nil {
		t.Fatal("graph loaded on Changes tab")
	}
	m.key("2")
	if m.count(tabCommits) != 1 {
		t.Fatalf("count = %d after switching to Commits", m.count(tabCommits))
	}
}

func TestHunkModeStagedRenameStagesWhole(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "old.txt", numbered(20, nil))
	r.Git("mv", "old.txt", "new.txt")
	m := &model{repo: r.Repo}
	m.load()
	m.key("enter")
	if m.hunks != nil || !strings.Contains(m.msg, "renamed file") {
		t.Fatalf("hunks=%v msg=%q", m.hunks, m.msg)
	}
}

func TestLineModeRefusedForNewAndDeletedFiles(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "gone.txt", "1\n2\n")
	r.Write("added.txt", "x\ny\n")
	r.Git("add", "added.txt")
	if err := os.Remove(filepath.Join(r.Top, "gone.txt")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"added.txt", "gone.txt"} {
		m := &model{repo: r.Repo}
		m.load()
		for i, f := range m.files {
			if f.path == path {
				m.cursor[tabChanges] = i
			}
		}
		m.key("enter")
		if m.hunks == nil {
			t.Fatalf("%s: no hunk view (msg %q)", path, m.msg)
		}
		m.key("v")
		if m.hunks.lineMode || !strings.Contains(m.msg, "whole") {
			t.Fatalf("%s: lineMode=%v msg=%q", path, m.hunks.lineMode, m.msg)
		}
	}
}

func TestLineModeMarksPickedLinesWithCheckbox(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", numbered(10, nil))
	r.Write("f.txt", numbered(10, map[int]string{5: "line 5\na\nb\n"}))
	m := &model{repo: r.Repo, w: 100, h: 30}
	m.load()
	m.key("enter")
	m.key("v")
	m.key("space") // picks "+a" under the cursor
	out := strings.Join(m.hunkDetail(80, 20), "\n")
	if strings.Count(out, "[x]") != 1 || !strings.Contains(out, "[ ]") {
		t.Fatalf("want one [x] and some [ ]:\n%s", out)
	}
}

func TestNextHintsHiddenInHunkMode(t *testing.T) {
	_, m := twoHunks(t)
	m.w, m.h = 100, 30
	if !strings.Contains(m.render(), "Next:") {
		t.Fatal("file list should show Next: hints")
	}
	m.key("enter")
	if out := m.render(); strings.Contains(out, "Next:") {
		t.Fatalf("hunk mode should not show Next: hints:\n%s", out)
	}
}
