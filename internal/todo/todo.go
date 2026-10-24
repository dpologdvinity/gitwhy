// Package todo rewrites a "git rebase -i" todo list from a plan, so rebases run without an editor.
// git calls "gitwhy __todo <plan> <todo-file>" as its sequence editor.
package todo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// Write saves a plan in a new temp directory and returns its path. Apply deletes the directory
// once git has read the plan; Write first deletes ones older than a day that were never run.
func Write(p Plan) (string, error) {
	removeStale(os.TempDir(), 24*time.Hour)
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", tempPrefix)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		_ = os.RemoveAll(dir) // best effort: the write error is the one to report
		return "", err
	}
	return path, nil
}

// tempPrefix starts the name of every directory Write makes.
const tempPrefix = "gitwhy-todo-"

// removeStale deletes Write's directories in tmp older than age: plans written for a rebase that
// never ran, for example because the user said no at the prompt.
func removeStale(tmp string, age time.Duration) {
	dirs, _ := filepath.Glob(filepath.Join(tmp, tempPrefix+"*"))
	for _, d := range dirs {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() && time.Since(fi.ModTime()) > age {
			_ = os.RemoveAll(d) // best effort: another user's directory may not be ours to delete
		}
	}
}

// Editor returns the GIT_SEQUENCE_EDITOR value that applies plan with the gitwhy binary self.
func Editor(self, plan string) string {
	return "GIT_SEQUENCE_EDITOR=" + quote(self) + " __todo " + quote(plan)
}

// quote wraps s in single quotes for the POSIX shell that git runs editors and exec steps in.
func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// line is one line of the todo file.
type line struct {
	text   string // the line as written
	action string // its todo command ("pick", "p", "exec", ...); "" for comments and blank lines
	commit string // abbreviated ID as written
	rest   string // subject
}

// commands are the todo commands git writes or accepts, long and short forms.
var commands = map[string]bool{
	"pick": true, "p": true, "reword": true, "r": true, "edit": true, "e": true,
	"squash": true, "s": true, "fixup": true, "f": true, "exec": true, "x": true,
	"break": true, "b": true, "drop": true, "d": true, "label": true, "l": true,
	"reset": true, "t": true, "merge": true, "m": true, "update-ref": true, "u": true,
	"noop": true,
}

// isPick reports whether todo command a is pick, long or short.
func isPick(a string) bool { return a == "pick" || a == "p" }

// Apply rewrites the todo file at todoPath according to the plan at planPath. With Full, the
// pick lines are replaced in the plan's order and any other step makes it refuse; with Edits,
// only the matching pick lines change. It deletes the plan once read.
func Apply(planPath, todoPath string) error {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return err
	}
	// git reads the plan once; nothing in it is needed after this, so delete it and its directory.
	_ = os.Remove(planPath) // best effort: a leftover file is only clutter
	if dir := filepath.Dir(planPath); strings.HasPrefix(filepath.Base(dir), tempPrefix) {
		_ = os.Remove(dir) // fails, harmlessly, if something else is in it
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
		// Anything that does not start with a todo command is a comment, whatever core.commentChar is.
		if len(f) >= 1 && commands[f[0]] {
			l.action = f[0]
		}
		if l.action != "" && len(f) >= 2 {
			l.commit = f[1]
			l.rest = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t), f[0]+" "+f[1]))
		}
		lines = append(lines, l)
	}
	render := func(it Item, l line) ([]string, error) {
		switch it.Action {
		case "pick", "drop", "edit", "squash", "fixup":
			return []string{it.Action + " " + l.commit + " " + l.rest}, nil
		case "reword":
			// The message travels in the exec line itself (one quoted argument per line, since a
			// todo step is one line), so no file has to outlive this editor run.
			var args []string
			for _, ml := range strings.Split(it.Message, "\n") {
				args = append(args, quote(ml))
			}
			return []string{"pick " + l.commit + " " + l.rest, `exec printf '%s\n' ` + strings.Join(args, " ") + " | git commit --amend --only --allow-empty --no-verify -F -"}, nil
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
