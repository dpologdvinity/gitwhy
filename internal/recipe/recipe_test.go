package recipe

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestLoadRealRecipes proves that every shipped recipe loads and validates, that every template
// in it (step args, "when", question texts and sources, predict and explain) parses and runs, and
// that every "suggest" names a recipe that exists.
func TestLoadRealRecipes(t *testing.T) {
	rs, err := Load(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) < 40 {
		t.Fatalf("got %d recipes", len(rs))
	}
	// Every template in every recipe must parse.
	a := Answers{"branch": "main", "head": "abc", "op": "merge", "remote": "origin", "upstream": "origin/main"}
	for _, r := range rs {
		for _, s := range r.Steps {
			if _, err := Truthy(s.When, a); err != nil {
				t.Errorf("%s when: %v", r.ID, err)
			}
			if _, err := Args(s, a); err != nil {
				t.Errorf("%s: %v", r.ID, err)
			}
		}
		for _, q := range r.Questions {
			for _, tm := range []string{q.Ask, q.Default, q.When, q.Source} {
				if _, err := Render(tm, a); err != nil {
					t.Errorf("%s question %s: %v", r.ID, q.ID, err)
				}
			}
		}
		for _, tm := range []string{r.Predict, r.Explain} {
			if _, err := Render(tm, a); err != nil {
				t.Errorf("%s predict/explain: %v", r.ID, err)
			}
		}
		for _, c := range r.Checks {
			if c.Suggest != "" && !hasID(rs, c.Suggest) {
				t.Errorf("%s suggests unknown recipe %q", r.ID, c.Suggest)
			}
		}
	}
}

// TestAliasesUnique proves that no alias belongs to two recipes, unless it is listed in allowed
// with a reason. A shared alias let search rank a destructive recipe above a safe one: "reset
// file" (which in git means unstage) found "Throw away changes to files" first.
func TestAliasesUnique(t *testing.T) {
	allowed := map[string]string{ // alias -> why sharing it is fine
		"committed to wrong branch": "both recipes move commits off the branch, to a new or an existing branch; neither is safer",
		"fixup":                     "both rewrite history the same way: into one past commit, or by planning several",
		"squash commits":            "both rewrite history the same way: the last commits, or a chosen plan",
	}
	rs, err := Load(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	owner := map[string]string{}
	for _, r := range rs {
		for _, al := range r.Aliases {
			al = strings.ToLower(strings.TrimSpace(al))
			if prev, ok := owner[al]; ok && allowed[al] == "" {
				t.Errorf("alias %q is in both %s and %s", al, prev, r.ID)
			}
			owner[al] = r.ID
		}
	}
}

// TestDeleteUntrackedPreset proves the contract with the TUI: delete-untracked given a preset
// "files" list skips its "Delete what?" question and runs "git clean -f -- <files>", which deletes
// only those files; without the preset it still cleans everything with the chosen flags.
func TestDeleteUntrackedPreset(t *testing.T) {
	rs, err := Load(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	var r *Recipe
	for _, x := range rs {
		if x.ID == "delete-untracked" {
			r = x
		}
	}
	if r == nil {
		t.Fatal("no delete-untracked recipe")
	}
	steps := func(a Answers) [][]string {
		var out [][]string
		for _, s := range r.Steps {
			if ok, err := Truthy(s.When, a); err != nil || !ok {
				continue
			}
			args, err := Args(s, a)
			if err != nil {
				t.Fatal(err)
			}
			out = append(out, args)
		}
		return out
	}
	if got := steps(Answers{"what": "-fd"}); !reflect.DeepEqual(got, [][]string{{"clean", "-fd"}}) {
		t.Errorf("without preset: %q", got)
	}
	preset := Answers{"files": []string{"b.txt"}}
	if ask, _ := Truthy(r.Question("what").When, preset); ask {
		t.Error("asks \"Delete what?\" although files are given")
	}
	got := steps(preset)
	if !reflect.DeepEqual(got, [][]string{{"clean", "-f", "--", "b.txt"}}) {
		t.Fatalf("with preset: %q", got)
	}
	repo := testrepo.New(t)
	repo.Commit("one", "t.txt", "t\n")
	repo.Write("a.txt", "a\n")
	repo.Write("b.txt", "b\n")
	repo.Git(got[0]...)
	if repo.Read("a.txt") != "a\n" || repo.Read("b.txt") != "<missing>" {
		t.Errorf("a.txt %q, b.txt %q: want only b.txt deleted", repo.Read("a.txt"), repo.Read("b.txt"))
	}
}

// hasID reports whether a recipe with the given ID is in rs.
func hasID(rs []*Recipe, id string) bool {
	for _, r := range rs {
		if r.ID == id {
			return true
		}
	}
	return false
}

// TestArgsSplatAndDrop proves the Args rules: a lone "{{.files}}" list answer becomes several
// arguments (spaces kept inside each), an empty "{{if ...}}" argument is dropped, and any other
// empty value stays as "" so a missing answer cannot shift the other arguments.
func TestArgsSplatAndDrop(t *testing.T) {
	s := Step{Git: []string{"merge", "{{if .mode}}{{.mode}}{{end}}", "--", "{{.files}}", "x{{.n}}", "-m", "{{.msg}}"}}
	got, err := Args(s, Answers{"mode": "", "files": []string{"a b", "c"}, "n": "1", "msg": ""})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"merge", "--", "a b", "c", "x1", "-m", ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}

// TestTruthy proves that an empty "when" is true, a template is true or false by its output, and
// a missing answer is false.
func TestTruthy(t *testing.T) {
	for tmpl, want := range map[string]bool{"": true, `{{eq .a "x"}}`: true, `{{eq .a "y"}}`: false, "{{.missing}}": false} {
		got, err := Truthy(tmpl, Answers{"a": "x"})
		if err != nil || got != want {
			t.Errorf("Truthy(%q) = %v, %v", tmpl, got, err)
		}
	}
}
