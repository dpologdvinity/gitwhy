package hunk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

func lines(n int, edit map[int]string) string {
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

// twoEdits commits a 20-line file, then edits lines 2 and 18.
func twoEdits(t *testing.T) *testrepo.Repo {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", lines(20, nil))
	r.Write("f.txt", lines(20, map[int]string{2: "two\n", 18: "eighteen\n"}))
	return r
}

func parse(t *testing.T, r *testrepo.Repo, args ...string) *FileDiff {
	t.Helper()
	out, err := r.OutputIn(nil, append(append([]string{"diff"}, args...), "--", "f.txt")...)
	if err != nil {
		t.Fatal(err)
	}
	fd, err := Parse(out)
	if err != nil {
		t.Fatal(err)
	}
	return fd
}

func apply(t *testing.T, r *testrepo.Repo, patch string, reverse bool) {
	t.Helper()
	args := []string{"apply", "--cached", "--recount"}
	if reverse {
		args = append(args, "-R")
	}
	if _, err := r.OutputIn(strings.NewReader(patch), args...); err != nil {
		t.Fatalf("%v\npatch:\n%s", err, patch)
	}
}

func TestParseTwoHunks(t *testing.T) {
	r := twoEdits(t)
	fd := parse(t, r)
	if len(fd.Hunks) != 2 {
		t.Fatalf("hunks = %d, want 2", len(fd.Hunks))
	}
	if !strings.Contains(strings.Join(fd.Header, "\n"), "+++ b/f.txt") {
		t.Fatalf("header = %q", fd.Header)
	}
	if a, d := fd.Hunks[0].Changes(); a != 1 || d != 1 {
		t.Fatalf("changes = +%d -%d", a, d)
	}
}

func TestPatchWholeHunkStagesOnlyThatHunk(t *testing.T) {
	r := twoEdits(t)
	apply(t, r, parse(t, r).Patch(0, nil, false), false)
	staged := r.Git("diff", "--cached")
	if !strings.Contains(staged, "+two") || strings.Contains(staged, "eighteen") {
		t.Fatalf("staged:\n%s", staged)
	}
	if !strings.Contains(r.Git("diff"), "+eighteen") {
		t.Fatal("line-18 change should stay unstaged")
	}
}

func TestPatchLinesForwardKeepsUnselectedDeletions(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "a\nold\nc\n")
	r.Write("f.txt", "a\nnew\nc\n")
	fd := parse(t, r)
	h := fd.Hunks[0]
	sel := make([]bool, len(h.Lines))
	for i, l := range h.Lines {
		sel[i] = l == "+new"
	}
	apply(t, r, fd.Patch(0, sel, false), false)
	if got := r.Git("show", ":f.txt"); got != "a\nold\nnew\nc" {
		t.Fatalf("index = %q", got)
	}
}

func TestPatchReverseUnstagesHunk(t *testing.T) {
	r := twoEdits(t)
	r.Git("add", "f.txt")
	apply(t, r, parse(t, r, "--cached").Patch(1, nil, true), true)
	staged := r.Git("diff", "--cached")
	if !strings.Contains(staged, "+two") || strings.Contains(staged, "eighteen") {
		t.Fatalf("staged:\n%s", staged)
	}
}

func TestParseBinaryAndModeOnly(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "b.bin", "\x00\x01", "f.txt", "x\n")
	r.Write("b.bin", "\x00\x02")
	r.Git("update-index", "--chmod=+x", "f.txt")
	for _, args := range [][]string{{"diff", "--", "b.bin"}, {"diff", "--cached", "--", "f.txt"}, {"diff", "--", "nothing"}} {
		out := r.Git(args...)
		fd, err := Parse(out)
		if err != nil || len(fd.Hunks) != 0 {
			t.Fatalf("%v: hunks=%d err=%v\n%s", args, len(fd.Hunks), err, out)
		}
	}
}
