package main

import (
	"strings"
	"testing"
)

// TestRunWarnsBeforeCheckoutOfAFile checks that `gitwhy run git checkout <file>`, the most common
// way to throw away edits, shows the "overwrites" warning, defaults to No at the prompt, and does
// not touch the edit when the user declines.
func TestRunWarnsBeforeCheckoutOfAFile(t *testing.T) {
	r := newRepo(t)
	r.write("f1.txt", "my edit\n")
	out := drive(t, r.dir, []string{"run", "git", "checkout", "f1.txt"}, text("n"))
	if !strings.Contains(out, "Overwrites changes in those files") {
		t.Fatalf("no warning before checkout of a file:\n%s", out)
	}
	expect(t, "edit kept after answering no", r.read("f1.txt"), "my edit\n", out)
}
