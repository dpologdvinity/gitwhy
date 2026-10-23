package hook

import (
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/snapshot"
	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

func TestPreSkips(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Write("f.txt", "dirty\n")
	cases := map[string][]string{
		t.TempDir(): {"reset", "--hard"},
		r.Top:       {"-C", "x", "reset", "--hard"},
	}
	for dir, args := range cases {
		if id := Pre(dir, args); id != "" {
			t.Errorf("Pre(%s, %v) = %q", dir, args, id)
		}
	}
	for _, args := range [][]string{{"status"}, {"push", "--force"}, {}} {
		if id := Pre(r.Top, args); id != "" {
			t.Errorf("Pre(%v) = %q", args, id)
		}
	}
	t.Setenv("GITWHY_HOOK_OFF", "1")
	if id := Pre(r.Top, []string{"reset", "--hard"}); id != "" {
		t.Errorf("hook off: %q", id)
	}
}

func TestPreThenUndo(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Write("f.txt", "dirty\n")
	id := Pre(r.Top, []string{"reset", "--hard"})
	if id == "" {
		t.Fatal("no snapshot for reset --hard")
	}
	r.Git("reset", "--hard")
	Post(r.Top, id)
	s, err := snapshot.Get(r.Repo, id)
	if err != nil {
		t.Fatal(err)
	}
	if s.After == nil {
		t.Fatal("Post did not finish the snapshot")
	}
	if _, err := snapshot.Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if got := r.Read("f.txt"); got != "dirty\n" {
		t.Fatalf("f.txt = %q", got)
	}
}
