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
	for _, f := range m.files {
		// Limited to one path, git can't pair the rename, so the staged side would look like a whole new file.
		if f.path == path && f.moved && staged {
			return errors.New("renamed file: press space to stage it whole")
		}
	}
	// Pin the output format: user config (color, prefixes, external diff) would break parsing and apply.
	args := []string{"-c", "diff.suppressBlankEmpty=false", "diff", "--no-color", "--no-ext-diff", "--src-prefix=a/", "--dst-prefix=b/"}
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
	if v.lineMode {
		return m.lineKey(k)
	}
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
	case "v":
		for _, l := range v.fd.Header {
			if strings.HasPrefix(l, "new file mode") || strings.HasPrefix(l, "deleted file mode") {
				m.msg = "This file is entirely added or deleted: stage it whole with space."
				return m, nil
			}
		}
		if v.fd.Hunks[v.cur].NoEOL() {
			m.msg = "This hunk touches the end of a file with no final newline: stage it whole with space."
			return m, nil
		}
		v.lineMode, v.sel, v.line = true, make([]bool, len(v.fd.Hunks[v.cur].Lines)), 0
		v.line = v.nextChange(-1, 1)
	case "S":
		if m.openHunks(v.path, !v.staged) != nil {
			m.msg = "Nothing on the other side."
		}
	}
	return m, nil
}

func (m *model) lineKey(k string) (tea.Model, tea.Cmd) {
	v := m.hunks
	switch k {
	case "ctrl+c":
		return m.pick(Action{Kind: "quit"})
	case "esc", "q":
		v.lineMode = false
	case "down", "j":
		v.line = v.nextChange(v.line, 1)
	case "up", "k":
		v.line = v.nextChange(v.line, -1)
	case "space":
		v.sel[v.line] = !v.sel[v.line]
	case "enter":
		picked := false
		for _, s := range v.sel {
			picked = picked || s
		}
		if !picked {
			m.msg = "Pick lines with space first"
		} else if err := m.applyHunk(v.sel); err != nil {
			m.msg = "✗ " + strings.TrimSpace(err.Error())
		}
	}
	return m, nil
}

// nextChange returns the next +/- line after i in direction dir, or i if none.
func (v *hunkView) nextChange(i, dir int) int {
	lines := v.fd.Hunks[v.cur].Lines
	for j := i + dir; j >= 0 && j < len(lines); j += dir {
		if lines[j][0] == '+' || lines[j][0] == '-' {
			return j
		}
	}
	return i
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
	if m.hunks.lineMode {
		return "pick line · enter stage picked lines · esc back"
	}
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
	lines := v.fd.Hunks[v.cur].Lines
	if !v.lineMode {
		return colorDiff(strings.Join(lines, "\n"), w, h)
	}
	rows := make([]string, len(lines))
	for i, l := range lines {
		mark := "  "
		if v.sel[i] {
			mark = "▌ "
		}
		rows[i] = mark + colorDiff(l, w-2, 1)[0]
	}
	return scroll(rows, v.line, w, h)
}
