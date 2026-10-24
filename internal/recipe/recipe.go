// Package recipe defines goal recipes: a recipe is a YAML file with checks, questions and git
// steps. This package loads and validates the embedded recipes/*.yaml files, renders step
// arguments from the user's answers with Go templates, and runs the named situation checks.
// It has no user interface; internal/ui asks the questions and runs the steps.
package recipe

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Recipe is one goal, e.g. "Undo last commit".
type Recipe struct {
	ID        string     `yaml:"id"` // defaults to the file name without ".yaml"
	Title     string     `yaml:"title"`
	Aliases   []string   `yaml:"aliases"` // extra words search matches against
	Category  string     `yaml:"category"`
	Explain   string     `yaml:"explain"` // shown before the questions
	Predict   string     `yaml:"predict"` // template: plain-English effect
	Checks    []Check    `yaml:"checks"`
	Questions []Question `yaml:"questions"`
	Steps     []Step     `yaml:"steps"`
}

// Check is a situation test run before the recipe starts.
type Check struct {
	If      string `yaml:"if"`      // check name, "!" prefix negates
	Warn    string `yaml:"warn"`    // show and ask to continue
	Block   string `yaml:"block"`   // show and stop
	Suggest string `yaml:"suggest"` // recipe ID to offer instead
}

// Question asks the user for one answer.
type Question struct {
	ID       string   `yaml:"id"`
	Ask      string   `yaml:"ask"`
	Type     string   `yaml:"type"` // choice, confirm, text, branch, commit, files, remote, todo
	Help     string   `yaml:"help"`
	Options  []Option `yaml:"options"` // choice: the options to pick from (required)
	Default  string   `yaml:"default"`
	Optional bool     `yaml:"optional"` // an empty answer is allowed
	When     string   `yaml:"when"`     // template; ask only if it renders truthy
	Split    bool     `yaml:"split"`    // text: split the answer on spaces into a list
	// For branch/commit/files/todo pickers.
	Source  string   `yaml:"source"`  // branch: local, remote, all; files: changed (default), staged, unstaged, conflicted, untracked, tracked; commit/todo: revision range (template)
	Multi   bool     `yaml:"multi"`   // pickers: allow several picks; the answer is a []string
	Partial bool     `yaml:"partial"` // files: must leave at least one file unpicked
	Actions []string `yaml:"actions"` // todo: allowed actions besides reorder (drop, squash, fixup, reword)
}

// Option is one choice of a "choice" question.
type Option struct {
	Label  string `yaml:"label"`  // text shown to the user
	Value  string `yaml:"value"`  // the answer stored when picked
	Danger string `yaml:"danger"` // warning shown next to a risky option
}

// Step is one git command, or a question asked mid-recipe.
type Step struct {
	Git  []string `yaml:"git"`  // args after "git"; templates; "{{.x}}" alone splats a list answer
	Ask  string   `yaml:"ask"`  // question ID to ask at this point instead of upfront
	When string   `yaml:"when"` // template; run only if truthy
	Env  []string `yaml:"env"`  // extra environment, e.g. GIT_SEQUENCE_EDITOR=true
	Todo []Edit   `yaml:"todo"` // for "rebase -i": edits gitwhy applies to the todo list
	Note string   `yaml:"note"` // shown before this step runs
	// Fail and Warn are decided when the recipe is planned, before anything runs, so their `when`
	// can only use answers known up front, not an answer a later `ask` step collects: by then the
	// snapshot is taken and earlier steps have run, so the step is skipped instead.
	Fail string `yaml:"fail"` // with when: stop before anything runs, showing this
	Warn string `yaml:"warn"` // with when: add this warning to the preview
}

// Edit changes one line of a rebase todo list, or replaces it with a plan.
type Edit struct {
	Action  string `yaml:"action"`  // drop, reword, edit, plan
	Commit  string `yaml:"commit"`  // template: commit to change
	Message string `yaml:"message"` // template: new message for reword
	Plan    string `yaml:"plan"`    // template: answer holding a full plan (from a todo question)
}

// Answers maps question IDs to string or []string values.
type Answers map[string]any

// AskedLater reports whether a question is asked by a step instead of upfront.
func (r *Recipe) AskedLater(id string) bool {
	for _, s := range r.Steps {
		if s.Ask == id {
			return true
		}
	}
	return false
}

// Question returns the question with the given ID or nil.
func (r *Recipe) Question(id string) *Question {
	for i := range r.Questions {
		if r.Questions[i].ID == id {
			return &r.Questions[i]
		}
	}
	return nil
}

// funcs is the set of extra functions recipe templates may call; "join" joins a list answer.
// It also accepts a text answer (returned as is) and an answer that is not known yet (empty), so
// a recipe's predict text can be rendered before every question has been asked.
var funcs = template.FuncMap{"join": func(list any, sep string) string {
	switch v := list.(type) {
	case []string:
		return strings.Join(v, sep)
	case string:
		return v
	}
	return ""
}}

// Render executes a template string with the answers. A missing answer renders as an empty
// string, not "<no value>".
func Render(tmpl string, a Answers) (string, error) {
	t, err := template.New("").Funcs(funcs).Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, map[string]any(a)); err != nil {
		return "", err
	}
	return strings.ReplaceAll(b.String(), "<no value>", ""), nil
}

// Truthy renders a "when" template; empty templates are true. Output that is empty, "false" or
// "0" after trimming spaces counts as false.
func Truthy(tmpl string, a Answers) (bool, error) {
	if strings.TrimSpace(tmpl) == "" {
		return true, nil
	}
	out, err := Render(tmpl, a)
	if err != nil {
		return false, err
	}
	out = strings.TrimSpace(out)
	return out != "" && out != "false" && out != "0", nil
}

// soloVar matches a step argument that is only "{{.name}}", so a list answer can be splatted into
// several arguments instead of being rendered as one string.
var soloVar = regexp.MustCompile(`^\{\{\s*\.(\w+)\s*\}\}$`)

// Args renders a step's git args. A list answer used alone splats. An element that starts with
// "{{if" is optional and dropped when it renders empty; any other empty value stays as "" so a
// missing answer can't silently change the command.
func Args(s Step, a Answers) ([]string, error) {
	var out []string
	for _, el := range s.Git {
		if m := soloVar.FindStringSubmatch(el); m != nil {
			if list, ok := a[m[1]].([]string); ok {
				out = append(out, list...)
				continue
			}
		}
		v, err := Render(el, a)
		if err != nil {
			return nil, fmt.Errorf("step %v: %w", s.Git, err)
		}
		if v == "" && strings.HasPrefix(strings.TrimSpace(el), "{{if") {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

// Load reads and validates every recipes/*.yaml file and returns the recipes sorted by title.
// A recipe without an ID gets its file name; a duplicate ID is an error.
func Load(fsys fs.FS) ([]*Recipe, error) {
	files, err := fs.Glob(fsys, "recipes/*.yaml")
	if err != nil {
		return nil, err
	}
	var rs []*Recipe
	seen := map[string]bool{}
	for _, f := range files {
		data, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, err
		}
		var r Recipe
		if err := yaml.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if r.ID == "" {
			r.ID = strings.TrimSuffix(path.Base(f), ".yaml")
		}
		if err := r.validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if seen[r.ID] {
			return nil, fmt.Errorf("%s: duplicate recipe id %q", f, r.ID)
		}
		seen[r.ID] = true
		rs = append(rs, &r)
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i].Title < rs[j].Title })
	return rs, nil
}

// questionTypes is the set of question types the wizard knows how to ask.
var questionTypes = map[string]bool{"choice": true, "confirm": true, "text": true, "branch": true, "commit": true, "files": true, "remote": true, "todo": true}

// validate reports the first structural mistake in a recipe: missing title, category or steps,
// an unknown question type, a choice without options, a step that does nothing or asks an
// unknown question, an unknown todo action, or an unknown check name. It runs at load time so a
// broken recipe fails the tests instead of failing for a user.
func (r *Recipe) validate() error {
	if r.Title == "" || r.Category == "" || len(r.Steps) == 0 {
		return fmt.Errorf("recipe %q needs title, category and steps", r.ID)
	}
	ids := map[string]bool{}
	for _, q := range r.Questions {
		if !questionTypes[q.Type] {
			return fmt.Errorf("question %q: unknown type %q", q.ID, q.Type)
		}
		if q.Type == "choice" && len(q.Options) == 0 {
			return fmt.Errorf("question %q: choice needs options", q.ID)
		}
		ids[q.ID] = true
	}
	for _, s := range r.Steps {
		if s.Ask != "" && !ids[s.Ask] {
			return fmt.Errorf("step asks unknown question %q", s.Ask)
		}
		if s.Ask == "" && len(s.Git) == 0 && s.Fail == "" && s.Warn == "" {
			return fmt.Errorf("step needs git, ask, fail or warn")
		}
		for _, e := range s.Todo {
			switch e.Action {
			case "drop", "edit", "reword", "plan":
			default:
				return fmt.Errorf("todo: unknown action %q", e.Action)
			}
		}
	}
	for _, c := range r.Checks {
		if _, ok := checks[strings.TrimPrefix(c.If, "!")]; !ok {
			return fmt.Errorf("unknown check %q", c.If)
		}
	}
	return nil
}
