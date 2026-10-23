package recipe

import (
	"os"
	"reflect"
	"testing"
)

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

func hasID(rs []*Recipe, id string) bool {
	for _, r := range rs {
		if r.ID == id {
			return true
		}
	}
	return false
}

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

func TestTruthy(t *testing.T) {
	for tmpl, want := range map[string]bool{"": true, `{{eq .a "x"}}`: true, `{{eq .a "y"}}`: false, "{{.missing}}": false} {
		got, err := Truthy(tmpl, Answers{"a": "x"})
		if err != nil || got != want {
			t.Errorf("Truthy(%q) = %v, %v", tmpl, got, err)
		}
	}
}
