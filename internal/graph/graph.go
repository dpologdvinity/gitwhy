// Package graph reads `git log --graph` into rows the TUI can draw.
package graph

import (
	"strconv"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Row is one output line: graph art, plus a commit unless it's a connector line.
type Row struct {
	Graph  string
	Commit *Commit
}

// Commit is the commit drawn on a row.
type Commit struct {
	ID, Short, Subject, When, Author, Refs string
}

// Load reads up to n commits of HEAD's history (all refs when all is set).
func Load(r *git.Repo, all bool, n int) ([]Row, error) {
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

// Parse splits Load's raw output into rows.
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
