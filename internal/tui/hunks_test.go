package tui

import (
	"fmt"
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
