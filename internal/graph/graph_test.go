package graph

import (
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestParseGraphSkipsConnectors feeds Parse hand-written --graph output and checks that connector
// lines become rows without a commit, that graph art keeps its shape minus trailing spaces, and
// that each commit field lands in the right place. The TUI relies on a nil Commit to make
// connector rows unselectable.
func TestParseGraphSkipsConnectors(t *testing.T) {
	out := "* \x00id1\x00a\x00s1\x00x\x00y\x00HEAD -> main\n" +
		"|\\  \n" +
		"| * \x00id2\x00b\x00s2\x00x\x00y\x00\n" +
		"|/  \n" +
		"* \x00id3\x00c\x00s3\x00x\x00y\x00"
	rows := Parse(out)
	if len(rows) != 5 {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[1].Commit != nil || rows[1].Graph != "|\\" {
		t.Fatalf("row 1 = %+v", rows[1])
	}
	if rows[2].Graph != "| *" || rows[2].Commit.ID != "id2" {
		t.Fatalf("row 2 = %+v", rows[2])
	}
	if c := rows[0].Commit; c.Refs != "HEAD -> main" || c.Short != "a" || c.Subject != "s1" {
		t.Fatalf("row 0 = %+v", c)
	}
}

// TestLoadWithMerge runs Load on a real repo with a --no-ff merge and checks that git's actual
// output parses into at least three commits and at least one connector line. This guards the
// --format string in Load against drifting from what Parse expects.
func TestLoadWithMerge(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "a.txt", "a\n")
	r.Git("switch", "-q", "-c", "side")
	r.Commit("side", "b.txt", "b\n")
	r.Git("switch", "-q", "main")
	r.Commit("main", "c.txt", "c\n")
	r.Git("merge", "-q", "--no-ff", "-m", "merge", "side")
	rows, err := Load(r.Repo, false, 50)
	if err != nil {
		t.Fatal(err)
	}
	commits, connectors := 0, 0
	for _, row := range rows {
		if row.Commit != nil {
			commits++
		} else {
			connectors++
		}
	}
	if commits < 3 || connectors < 1 {
		t.Fatalf("commits=%d connectors=%d rows=%+v", commits, connectors, rows)
	}
}
