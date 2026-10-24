// Package graph reads `git log --graph` into rows the TUI can draw.
//
// Each output line becomes a Row. Lines that carry a commit get its details; connector lines
// (only graph art, such as "|\") have no commit, so the TUI draws them but never selects them.
package graph

import (
	"strconv"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Row is one output line: graph art, plus a commit unless it's a connector line.
type Row struct {
	Graph  string  // graph art such as "| *", with trailing spaces removed
	Commit *Commit // nil on a connector line
}

// Commit is the commit drawn on a row.
type Commit struct {
	// ID is the full hash, Short the abbreviated one, When the relative committer date ("2 days
	// ago"), Author the author name, and Refs the ref names pointing at it ("HEAD -> main, tag:
	// v1"), or "" if none.
	ID, Short, Subject, When, Author, Refs string
}

// Load reads up to n commits of HEAD's history (all refs when all is set). n limits commits, not
// rows: connector lines come on top.
func Load(r *git.Repo, all bool, n int) ([]Row, error) {
	// Every field starts with a NUL, so the graph art is everything before the first NUL and a
	// line without one is a connector. --color=never keeps escape codes out of the graph art.
	args := []string{"log", "--graph", "--color=never", "-n", strconv.Itoa(n), "--format=%x00%H%x00%h%x00%s%x00%cr%x00%an%x00%D"}
	if all {
		args = append(args, "--all")
	}
	out, err := r.Output(append(args, "--")...)
	if err != nil {
		return nil, err
	}
	return Parse(out), nil
}

// Parse splits Load's raw output into rows, one per line. A line gets a Commit only when it has
// exactly the six NUL-separated fields Load asks for; any other line is a connector.
func Parse(out string) []Row {
	if out == "" {
		return nil
	}
	var rows []Row
	for _, l := range strings.Split(out, "\n") {
		g, fields, ok := strings.Cut(l, "\x00")
		row := Row{Graph: strings.TrimRight(g, " ")}
		if f := strings.Split(fields, "\x00"); ok && len(f) == 6 {
			row.Commit = &Commit{f[0], f[1], f[2], f[3], f[4], f[5]}
		}
		rows = append(rows, row)
	}
	return rows
}
