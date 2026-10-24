package recover

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

func TestEventsAndUndo(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	two := r.Commit("two", "a.txt", "2\n")
	r.Git("switch", "-q", "-c", "other")
	r.Commit("other work", "b.txt", "b\n")
	r.Git("switch", "-q", "main")
	r.Git("merge", "-q", "--no-ff", "--no-edit", "other")
	r.Git("reset", "-q", "--hard", "HEAD~1")

	evs, err := Events(r.Repo, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) < 3 {
		t.Fatalf("got %d events: %+v", len(evs), evs)
	}
	if evs[0].Kind != "reset" || evs[1].Kind != "merge" {
		t.Fatalf("kinds: %s, %s (%q, %q)", evs[0].Kind, evs[1].Kind, evs[0].Raw, evs[1].Raw)
	}
	if !strings.Contains(evs[1].Text, "1 commit") {
		t.Errorf("merge text: %q", evs[1].Text)
	}
	// Undoing the merge event goes back to "two".
	undo := evs[1].Undo(r.Repo)
	if strings.Join(undo, " ") != "reset --keep "+two {
		t.Errorf("undo = %q", undo)
	}
	if !strings.Contains(evs[0].Text, "1 commit(s) left") {
		t.Errorf("reset text: %q", evs[0].Text)
	}
}

func TestLostObjects(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	r.Git("switch", "-q", "-c", "gone")
	r.Commit("lost work", "b.txt", "b\n")
	r.Git("switch", "-q", "main")
	r.Git("branch", "-D", "gone")
	r.Write("staged.txt", "secret draft\n")
	r.Git("add", "staged.txt")
	r.Git("reset", "-q", "--hard")

	lost, err := LostObjects(r.Repo, 10)
	if err != nil {
		t.Fatal(err)
	}
	var commit, blob bool
	for _, l := range lost {
		commit = commit || (l.Kind == "commit" && l.Title == "lost work")
		blob = blob || (l.Kind == "blob" && l.Title == "secret draft")
	}
	if !commit || !blob {
		t.Fatalf("lost = %+v", lost)
	}
}

// TestLostObjectsManyBlobsIsFast proves LostObjects reads only the blobs it returns, in one
// git process, and only the start of each: it used to read every dangling blob in full with one
// process each before cutting the list to max.
func TestLostObjectsManyBlobsIsFast(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("one", "a.txt", "1\n")
	dir := t.TempDir()
	var paths []string
	big := filepath.Join(dir, "big")
	if err := os.WriteFile(big, []byte("big first line "+strings.Repeat("x", 100)+"\n"+strings.Repeat("y", 8<<20)), 0o644); err != nil {
		t.Fatal(err)
	}
	paths = append(paths, big)
	for i := 0; i < 300; i++ {
		p := filepath.Join(dir, strconv.Itoa(i))
		if err := os.WriteFile(p, []byte("small "+strconv.Itoa(i)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	if out, err := r.OutputIn(strings.NewReader(strings.Join(paths, "\n")+"\n"), "hash-object", "-w", "--stdin-paths"); err != nil {
		t.Fatal(out, err)
	}
	start := time.Now()
	lost, err := LostObjects(r.Repo, 400)
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 6*time.Second {
		t.Errorf("LostObjects took %s", d)
	}
	if len(lost) != 301 {
		t.Fatalf("got %d lost objects, want 301", len(lost))
	}
	var bigTitle string
	for _, l := range lost {
		if strings.HasPrefix(l.Title, "big") {
			bigTitle = l.Title
		}
	}
	if want := "big first line " + strings.Repeat("x", 55) + "..."; bigTitle != want {
		t.Errorf("big title = %q, want %q", bigTitle, want)
	}
	if few, _ := LostObjects(r.Repo, 5); len(few) != 5 {
		t.Errorf("max 5 gave %d", len(few))
	}
}
