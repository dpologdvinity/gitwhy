package search

import (
	"os"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/kb"
	"github.com/dpologdvinity/gitwhy/internal/recipe"
)

// TestSearchFindsRecipes proves that everyday phrasings, including a typo ("undo comit") and git
// jargon ("cherry pick"), find the intended recipe in the top 3, and that a dashed query still
// returns a result.
func TestSearchFindsRecipes(t *testing.T) {
	fsys := os.DirFS("../..")
	k, err := kb.Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	rs, err := recipe.Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	ix := New(k, rs)
	cases := map[string]string{ // query -> recipe expected in top 3
		"undo last commit":             "undo-last-commit",
		"undo comit":                   "undo-last-commit",
		"rename commit":                "reword-commit",
		"reorder commits":              "edit-recent-commits",
		"squash":                       "squash-last-commits",
		"split commit":                 "split-commit",
		"make branch same as master":   "reset-branch-to",
		"make master same as remote":   "sync-with-remote",
		"move commits to other branch": "move-commits-to-branch",
		"delete branch":                "delete-branch",
		"merge conflict":               "resolve-conflicts",
		"big files lfs":                "lfs-track",
		"remove password from history": "remove-file-from-history",
		"cherry pick":                  "copy-commit",
		"commit staged":                "commit-staged",
	}
	for q, want := range cases {
		found := false
		for _, r := range ix.Search(q, 3) {
			if r.Kind == Recipe && r.Recipe.ID == want {
				found = true
			}
		}
		if !found {
			var got []string
			for _, r := range ix.Search(q, 3) {
				got = append(got, r.Title)
			}
			t.Errorf("%q: %s not in top 3: %q", q, want, got)
		}
	}
	if r := ix.Search("cherry-pick", 1); len(r) == 0 {
		t.Error("no result for cherry-pick")
	}
}

// TestSafeRecipeFirst proves that for "reset file" (git's way to unstage a file) the first
// recipe found is "Unstage files", not the destructive "Throw away changes to files".
func TestSafeRecipeFirst(t *testing.T) {
	fsys := os.DirFS("../..")
	k, err := kb.Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	rs, err := recipe.Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range New(k, rs).Search("reset file", 15) {
		if r.Kind == Recipe {
			if r.Recipe.ID != "unstage-files" {
				t.Errorf("first recipe for \"reset file\" is %s, want unstage-files", r.Recipe.ID)
			}
			return
		}
	}
	t.Error("no recipe for \"reset file\"")
}

// TestCloseEnough proves that one missing letter, one swap of neighbors or one changed letter is a
// typo match, and that different words of similar length are not.
func TestCloseEnough(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{{"comit", "commit", true}, {"rebsae", "rebase", true}, {"merge", "marge", true}, {"stash", "squash", false}} {
		if got := closeEnough(c.a, c.b); got != c.want {
			t.Errorf("closeEnough(%q,%q) = %v", c.a, c.b, got)
		}
	}
}
