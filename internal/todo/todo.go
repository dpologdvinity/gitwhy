// Package todo rewrites a "git rebase -i" todo list from a plan, so rebases run without an editor.
// git calls "gitwhy __todo <plan> <todo-file>" as its sequence editor.
package todo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Item is one change to the todo list.
type Item struct {
	Action  string `json:"action"` // pick, drop, edit, reword, squash, fixup
	Commit  string `json:"commit"` // full commit ID
	Message string `json:"message,omitempty"`
}

// Plan says how to rewrite the todo list.
type Plan struct {
	Edits []Item `json:"edits,omitempty"` // change the matching lines, leave the rest
	Full  []Item `json:"full,omitempty"`  // the exact new order of every pick line
}

// Write saves a plan in a new temp directory and returns its path.
func Write(p Plan) (string, error) {
	dir, err := os.MkdirTemp("", "gitwhy-todo-")
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "plan.json")
	return path, os.WriteFile(path, data, 0o600)
}

// Editor returns the GIT_SEQUENCE_EDITOR value that applies plan with the gitwhy binary self.
func Editor(self, plan string) string {
	return "GIT_SEQUENCE_EDITOR=" + quote(self) + " __todo " + quote(plan)
}

func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

type line struct {
	text   string
	action string
	commit string // abbreviated ID as written
	rest   string // subject
}

func isPick(a string) bool { return a == "pick" || a == "p" }

// Apply rewrites the todo file at todoPath according to the plan at planPath.
func Apply(planPath, todoPath string) error {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return err
	}
	var p Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	raw, err := os.ReadFile(todoPath)
	if err != nil {
		return err
	}
	var lines []line
	for _, t := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		l := line{text: t}
		f := strings.Fields(t)
		if len(f) >= 2 && !strings.HasPrefix(t, "#") {
			l.action, l.commit = f[0], f[1]
			l.rest = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t), f[0]+" "+f[1]))
		}
		lines = append(lines, l)
	}
	msgDir := filepath.Dir(planPath)
	n := 0
	render := func(it Item, l line) ([]string, error) {
		switch it.Action {
		case "pick", "drop", "edit", "squash", "fixup":
			return []string{it.Action + " " + l.commit + " " + l.rest}, nil
		case "reword":
			n++
			mf := filepath.Join(msgDir, fmt.Sprintf("msg-%d.txt", n))
			if err := os.WriteFile(mf, []byte(it.Message+"\n"), 0o600); err != nil {
				return nil, err
			}
			return []string{"pick " + l.commit + " " + l.rest, "exec git commit --amend --only --allow-empty --no-verify -F " + quote(mf)}, nil
		}
		return nil, fmt.Errorf("unknown action %q", it.Action)
	}
	find := func(commit string) int {
		for i, l := range lines {
			if isPick(l.action) && l.commit != "" && strings.HasPrefix(commit, l.commit) {
				return i
			}
		}
		return -1
	}

	var out []string
	switch {
	case len(p.Full) > 0:
		var picks []int
		for i, l := range lines {
			if isPick(l.action) {
				picks = append(picks, i)
			}
		}
		for _, l := range lines {
			if l.action != "" && !isPick(l.action) {
				// update-ref, exec, label, merge...: moving picks around them would break them.
				return fmt.Errorf("this rebase has %q steps; gitwhy's editor only handles plain commits", l.action)
			}
		}
		if len(picks) != len(p.Full) {
			return fmt.Errorf("plan has %d commits but the rebase has %d", len(p.Full), len(picks))
		}
		for _, it := range p.Full {
			i := find(it.Commit)
			if i < 0 {
				return fmt.Errorf("commit %s isn't in this rebase", it.Commit)
			}
			r, err := render(it, lines[i])
			if err != nil {
				return err
			}
			out = append(out, r...)
		}
		for _, l := range lines {
			if !isPick(l.action) {
				out = append(out, l.text)
			}
		}
	default:
		repl := map[int][]string{}
		for _, it := range p.Edits {
			i := find(it.Commit)
			if i < 0 {
				return fmt.Errorf("commit %s isn't in this rebase", it.Commit)
			}
			r, err := render(it, lines[i])
			if err != nil {
				return err
			}
			repl[i] = r
		}
		for i, l := range lines {
			if r, ok := repl[i]; ok {
				out = append(out, r...)
			} else {
				out = append(out, l.text)
			}
		}
	}
	return os.WriteFile(todoPath, []byte(strings.Join(out, "\n")+"\n"), 0o644)
}
