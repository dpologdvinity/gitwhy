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
// and empty diffs give no hunks. It returns an error for any line it
// does not expect, such as a second file's header inside a hunk or an
// empty line in a hunk body.
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

// NoEOL reports whether the hunk has a "\\ No newline at end of file" line. Picking single
// lines out of such a hunk can't be expressed as a patch, so callers stage it whole.
func (h Hunk) NoEOL() bool {
	for _, l := range h.Lines {
		if l[0] == '\\' {
			return true
		}
	}
	return false
}

// keepHeader lists the prefixes of the header lines Patch copies into a patch. Other header
// lines, such as rename, copy and similarity lines, are left out.
var keepHeader = []string{"diff --git", "index", "---", "+++", "old mode", "new mode", "new file mode", "deleted file mode"}

// Patch builds a patch for hunk h alone, for `git apply --cached --recount`
// (add -R when reverse). sel picks change lines by index into Lines; nil
// means all. Unpicked lines become context on the side that has them.
// The hunk header's line counts are copied unchanged, which is why
// --recount is needed.
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
		// The patch is applied to the index. Staging (forward), an unpicked "+" line is not in
		// the index yet, so it is dropped, and an unpicked "-" line is still there, so it stays as
		// context. Unstaging (reverse, applied with -R) it is the other way round.
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
