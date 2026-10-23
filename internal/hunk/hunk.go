// Package hunk splits a one-file diff into hunks and builds patches that
// stage or unstage part of it.
package hunk

import (
	"fmt"
	"strings"
)

// Hunk is one "@@" section of a diff.
type Hunk struct {
	Header string   // the full "@@ ... @@ context" line
	Lines  []string // body lines, each starting with ' ', '+', '-' or '\'
}

// FileDiff is the diff of a single file.
type FileDiff struct {
	Header []string // every line before the first "@@"
	Hunks  []Hunk
}

// Parse reads the output of `git diff -- <one path>`. Binary, mode-only
// and empty diffs give no hunks.
func Parse(diff string) (*FileDiff, error) {
	fd := &FileDiff{}
	for _, line := range strings.Split(strings.TrimRight(diff, "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "@@"):
			fd.Hunks = append(fd.Hunks, Hunk{Header: line})
		case len(fd.Hunks) == 0:
			if line != "" {
				fd.Header = append(fd.Header, line)
			}
		case line != "" && strings.ContainsRune(" +-\\", rune(line[0])):
			h := &fd.Hunks[len(fd.Hunks)-1]
			h.Lines = append(h.Lines, line)
		default:
			return nil, fmt.Errorf("unexpected diff line %q (one file only)", line)
		}
	}
	return fd, nil
}

// Changes counts the hunk's added and removed lines.
func (h Hunk) Changes() (added, removed int) {
	for _, l := range h.Lines {
		switch l[0] {
		case '+':
			added++
		case '-':
			removed++
		}
	}
	return added, removed
}

var keepHeader = []string{"diff --git", "index", "---", "+++", "old mode", "new mode", "new file mode", "deleted file mode"}

// Patch builds a patch for hunk h alone, for `git apply --cached --recount`
// (add -R when reverse). sel picks change lines by index into Lines; nil
// means all. Unpicked lines become context on the side that has them.
func (fd *FileDiff) Patch(h int, sel []bool, reverse bool) string {
	var b strings.Builder
	for _, l := range fd.Header {
		for _, p := range keepHeader {
			if strings.HasPrefix(l, p) {
				b.WriteString(l + "\n")
				break
			}
		}
	}
	hk := fd.Hunks[h]
	b.WriteString(hk.Header + "\n")
	for i, l := range hk.Lines {
		keep := sel == nil || sel[i]
		switch l[0] {
		case '+':
			if keep {
				b.WriteString(l + "\n")
			} else if reverse {
				b.WriteString(" " + l[1:] + "\n")
			}
		case '-':
			if keep {
				b.WriteString(l + "\n")
			} else if !reverse {
				b.WriteString(" " + l[1:] + "\n")
			}
		default:
			b.WriteString(l + "\n")
		}
	}
	return b.String()
}
