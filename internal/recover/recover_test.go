package recover

import (
	"strings"
	"testing"

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
