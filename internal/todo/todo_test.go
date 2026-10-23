package todo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = `pick 1111111 one
pick 2222222 two
pick 3333333 three

# Rebase ...
`

func run(t *testing.T, p Plan) string {
	t.Helper()
	plan, err := Write(p)
	if err != nil {
		t.Fatal(err)
	}
	todo := filepath.Join(t.TempDir(), "git-rebase-todo")
	os.WriteFile(todo, []byte(sample), 0o644)
	if err := Apply(plan, todo); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(todo)
	return string(data)
}

func TestEdits(t *testing.T) {
	got := run(t, Plan{Edits: []Item{{Action: "drop", Commit: "2222222abcdef"}, {Action: "reword", Commit: "3333333ffff", Message: "new msg"}}})
	if !strings.Contains(got, "pick 1111111 one\ndrop 2222222 two\npick 3333333 three\nexec git commit --amend") {
		t.Fatalf("got:\n%s", got)
	}
}

func TestFullReorder(t *testing.T) {
	got := run(t, Plan{Full: []Item{{"pick", "3333333x", ""}, {"pick", "1111111x", ""}, {"fixup", "2222222x", ""}}})
	if !strings.HasPrefix(got, "pick 3333333 three\npick 1111111 one\nfixup 2222222 two\n") {
		t.Fatalf("got:\n%s", got)
	}
}

func TestMismatchFails(t *testing.T) {
	plan, _ := Write(Plan{Full: []Item{{"pick", "1111111x", ""}}})
	todo := filepath.Join(t.TempDir(), "todo")
	os.WriteFile(todo, []byte(sample), 0o644)
	if err := Apply(plan, todo); err == nil {
		t.Fatal("expected error for incomplete plan")
	}
}

func TestFullRefusesUpdateRef(t *testing.T) {
	plan, _ := Write(Plan{Full: []Item{{"pick", "1111111x", ""}}})
	todo := filepath.Join(t.TempDir(), "todo")
	os.WriteFile(todo, []byte("pick 1111111 one\nupdate-ref refs/heads/stack1\n"), 0o644)
	if err := Apply(plan, todo); err == nil {
		t.Fatal("expected refusal")
	}
}
