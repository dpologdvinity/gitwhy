package hunk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// lines returns n lines "line 1\n" to "line n\n", with line i replaced by edit[i] where set.
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

// twoEdits commits a 20-line file, then edits lines 2 and 18. The edits are far enough apart
// that git diff shows them as two hunks.
func twoEdits(t *testing.T) *testrepo.Repo {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", lines(20, nil))
	r.Write("f.txt", lines(20, map[int]string{2: "two\n", 18: "eighteen\n"}))
	return r
}

// parse runs `git diff <args> -- f.txt` in r and parses the output, failing the test on error.
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

// apply applies patch to r's index the way the TUI does (`git apply --cached --recount`, with
// -R when reverse), failing the test with the patch shown if git rejects it.
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

// TestParseTwoHunks checks that a real git diff splits into its two hunks, keeps the file header
// and counts each hunk's added and removed lines.
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

// TestPatchWholeHunkStagesOnlyThatHunk checks that staging the first hunk leaves the second one
// unstaged, which is the point of hunk staging.
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

// TestPatchLinesForwardKeepsUnselectedDeletions checks line staging: staging only the "+new" line
// of a changed line keeps the unpicked "-old" line in the index as context, so the index gets
// both lines rather than losing "old".
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

// TestPatchReverseUnstagesHunk checks that a reverse patch built from `git diff --cached` unstages
// only the chosen hunk and leaves the other one staged.
func TestPatchReverseUnstagesHunk(t *testing.T) {
	r := twoEdits(t)
	r.Git("add", "f.txt")
	apply(t, r, parse(t, r, "--cached").Patch(1, nil, true), true)
	staged := r.Git("diff", "--cached")
	if !strings.Contains(staged, "+two") || strings.Contains(staged, "eighteen") {
		t.Fatalf("staged:\n%s", staged)
	}
}

// TestParseBinaryAndModeOnly checks that binary, mode-only and empty diffs parse without error
// and give no hunks, so the TUI shows nothing to pick instead of failing.
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

// TestPatchReverseWithSelectionUnstagesOnlyPickedLines checks line unstaging: unstaging only the
// "-2" line puts "2" back in the index while the staged "X" and "Y" stay, so unpicked "+" lines
// must become context when reversing.
func TestPatchReverseWithSelectionUnstagesOnlyPickedLines(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n2\n4\n")
	r.Write("f.txt", "1\nX\nY\n4\n")
	r.Git("add", "f.txt")
	fd := parse(t, r, "--cached")
	sel := make([]bool, len(fd.Hunks[0].Lines))
	for i, l := range fd.Hunks[0].Lines {
		sel[i] = l == "-2"
	}
	apply(t, r, fd.Patch(0, sel, true), true)
	if got := r.Git("show", ":f.txt"); got != "1\n2\nX\nY\n4" {
		t.Fatalf("index = %q", got)
	}
}
