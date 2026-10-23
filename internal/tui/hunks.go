package tui

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/dpologdvinity/gitwhy/internal/hunk"
)

// hunkView is the Changes tab zoomed into one file's hunks.
type hunkView struct {
	path     string
	staged   bool // showing the staged side (space unstages)
	fd       *hunk.FileDiff
	cur      int
	lineMode bool
	line     int    // cursor in fd.Hunks[cur].Lines (line mode)
	sel      []bool // picked lines (line mode)
}

func (m *model) openHunks(path string, staged bool) error {
	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}
	out, err := m.repo.OutputIn(nil, append(args, "--", path)...)
	if err != nil {
		return err
	}
	fd, err := hunk.Parse(out)
	if err != nil {
		return err
	}
	if len(fd.Hunks) == 0 {
		return errors.New("no lines to show (binary, renamed or mode-only change): press space to stage it whole")
	}
	cur := 0
	if m.hunks != nil && m.hunks.path == path && m.hunks.staged == staged {
		cur = min(m.hunks.cur, len(fd.Hunks)-1)
	}
	m.hunks = &hunkView{path: path, staged: staged, fd: fd, cur: cur}
	return nil
}

func (m *model) hunkKey(k string) (tea.Model, tea.Cmd) {
	m.msg = ""
	v := m.hunks
	switch k {
	case "ctrl+c":
		return m.pick(Action{Kind: "quit"})
	case "esc", "q":
		m.hunks = nil
		m.loadDetail()
	case "down", "j":
		if v.cur+1 < len(v.fd.Hunks) {
			v.cur++
		}
	case "up", "k":
		if v.cur > 0 {
			v.cur--
		}
	case "space":
		if err := m.applyHunk(nil); err != nil {
			m.msg = "✗ " + strings.TrimSpace(err.Error())
		}
	case "S":
		if m.openHunks(v.path, !v.staged) != nil {
			m.msg = "Nothing on the other side."
		}
	}
	return m, nil
}

// applyHunk stages (or, on the staged side, unstages) the selected hunk, or
// only the sel lines of it, then reloads. It leaves hunk mode when that side
// has nothing left.
func (m *model) applyHunk(sel []bool) error {
	v := m.hunks
	args := []string{"apply", "--cached", "--recount"}
	verb := "staged"
	if v.staged {
		args = append(args, "-R")
		verb = "unstaged"
	}
	if _, err := m.repo.OutputIn(strings.NewReader(v.fd.Patch(v.cur, sel, v.staged)), args...); err != nil {
		return err
	}
	m.msg = "✓ " + verb + " part of " + v.path
	m.load()
	if m.openHunks(v.path, v.staged) != nil {
		m.hunks = nil
		m.loadDetail()
	}
	return nil
}

func (m *model) hunkTitle() string {
	side := "unstaged changes"
	if m.hunks.staged {
		side = "staged changes"
	}
	return bold.Render(m.hunks.path) + faint.Render(" — "+side)
}

func (m *model) hunkKeys() string {
	verb := "stage"
	if m.hunks.staged {
		verb = "unstage"
	}
	return verb + " hunk · S staged/unstaged · v pick lines · esc back"
}

func (m *model) hunkList(w, h int) []string {
	var rows []string
	for _, hk := range m.hunks.fd.Hunks {
		a, d := hk.Changes()
		rows = append(rows, hunkCol.Render(hk.Header)+"  "+addLine.Render(fmt.Sprintf("+%d", a))+" "+delLine.Render(fmt.Sprintf("−%d", d)))
	}
	return scroll(rows, m.hunks.cur, w, h)
}

func (m *model) hunkDetail(w, h int) []string {
	v := m.hunks
	return colorDiff(strings.Join(v.fd.Hunks[v.cur].Lines, "\n"), w, h)
}
