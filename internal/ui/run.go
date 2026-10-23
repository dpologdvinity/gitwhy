package ui

import (
	"fmt"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/danger"
	"github.com/dpologdvinity/gitwhy/internal/preview"
	"github.com/dpologdvinity/gitwhy/internal/recipe"
	"github.com/dpologdvinity/gitwhy/internal/snapshot"
	"github.com/dpologdvinity/gitwhy/internal/todo"
)

// job is what will run: recipe steps with answers, or one literal command.
type job struct {
	title   string
	r       *recipe.Recipe // nil for a literal command
	steps   []recipe.Step
	ans     recipe.Answers
	literal []string // args of a literal command (not templated)
	predict string
}

// planned is a step ready to show: its args with later answers as placeholders.
type planned struct {
	step recipe.Step
	args []string
}

func (a *App) plan(j job) ([]planned, error) {
	if j.literal != nil {
		return []planned{{args: j.literal}}, nil
	}
	shown := recipe.Answers{}
	for k, v := range j.ans {
		shown[k] = v
	}
	for _, s := range j.steps {
		if s.Ask != "" {
			if _, ok := shown[s.Ask]; !ok {
				shown[s.Ask] = "<" + s.Ask + ">"
			}
		}
	}
	var out []planned
	for _, s := range j.steps {
		if s.Ask != "" {
			continue
		}
		ok, err := recipe.Truthy(s.When, shown)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		args, err := recipe.Args(s, shown)
		if err != nil {
			return nil, err
		}
		out = append(out, planned{step: s, args: args})
	}
	return out, nil
}

// execute previews, explains, confirms, snapshots and runs a job.
func (a *App) execute(j job) error {
	if err := a.needRepo(); err != nil {
		return err
	}
	steps, err := a.plan(j)
	if err != nil {
		return err
	}
	if len(steps) == 0 {
		say("Nothing to do.")
		return nil
	}

	// What will happen.
	heading("What will happen")
	var save []string
	exact := false
	if len(steps) == 1 {
		p := preview.Preview(a.Repo, steps[0].args)
		if p.Exact {
			exact = true
			for _, l := range p.Lines {
				say("  %s", l)
			}
		} else if len(p.Lines) > 0 {
			for _, l := range p.Lines {
				say("  %s", l)
			}
		}
		save = p.Save
	}
	if !exact && j.predict != "" {
		say("  %s", strings.TrimSpace(j.predict))
	}
	if !exact && j.predict == "" && len(steps) == 1 {
		if c := a.KB.Commands[steps[0].args[0]]; c != nil {
			say("  %s", c.Summary)
		}
	}

	// Dangers.
	var warnings []string
	seen := map[string]bool{}
	for _, s := range steps {
		for _, w := range danger.Warnings(s.args) {
			if !seen[w] {
				seen[w] = true
				warnings = append(warnings, w)
			}
		}
	}
	if len(warnings) > 0 {
		say("")
		for _, w := range warnings {
			warn(w)
		}
	}

	// The commands, explained.
	if len(steps) == 1 {
		heading("Command")
	} else {
		heading(fmt.Sprintf("Commands (%d steps)", len(steps)))
	}
	for _, s := range steps {
		if s.step.Note != "" {
			say("  %s", dimStyle.Render(s.step.Note))
		}
		a.showCommand(s.args)
	}
	say("")

	ok, err := confirm("Run it?", len(warnings) == 0)
	if err != nil || !ok {
		if err == nil {
			say("Nothing ran.")
		}
		return err
	}

	label := "git " + shellJoin(steps[0].args)
	if len(steps) > 1 {
		label = j.title
	}
	snap, err := snapshot.Take(a.Repo, label, save)
	if err != nil {
		return fmt.Errorf("couldn't save an undo point, so nothing ran: %w", err)
	}
	_ = snap

	if j.literal != nil {
		return a.runOne(recipe.Step{}, j.literal, j.ans)
	}
	for _, s := range j.steps {
		if s.Ask != "" {
			q := j.r.Question(s.Ask)
			if err := a.ask(q, j.ans); err != nil {
				return a.failed(err)
			}
			continue
		}
		ok, err := recipe.Truthy(s.When, j.ans)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		args, err := recipe.Args(s, j.ans)
		if err != nil {
			return err
		}
		if err := a.runOne(s, args, j.ans); err != nil {
			return err
		}
	}
	a.done()
	return nil
}

func (a *App) runOne(s recipe.Step, args []string, ans recipe.Answers) error {
	env := append([]string{}, s.Env...)
	if len(s.Todo) > 0 {
		e, err := a.todoEnv(s, ans)
		if err != nil {
			return a.failed(err)
		}
		env = append(env, e)
	}
	say("%s", dimStyle.Render("$ git "+shellJoin(args)))
	if err := a.Repo.With(env...).Interactive(args...); err != nil {
		return a.failed(err)
	}
	if s.Git == nil { // literal command
		a.done()
	}
	return nil
}

func (a *App) todoEnv(s recipe.Step, ans recipe.Answers) (string, error) {
	var p todo.Plan
	for _, e := range s.Todo {
		if e.Action == "plan" {
			path, err := recipe.Render(e.Plan, ans)
			if err != nil {
				return "", err
			}
			return todo.Editor(a.Self, path), nil
		}
		commit, err := recipe.Render(e.Commit, ans)
		if err != nil {
			return "", err
		}
		msg, err := recipe.Render(e.Message, ans)
		if err != nil {
			return "", err
		}
		p.Edits = append(p.Edits, todo.Item{Action: e.Action, Commit: commit, Message: msg})
	}
	path, err := todo.Write(p)
	if err != nil {
		return "", err
	}
	return todo.Editor(a.Self, path), nil
}

func (a *App) done() {
	say("")
	say("%s Undo with: %s", okStyle.Render("✓ Done."), cmdStyle.Render("gitwhy undo"))
}

func (a *App) failed(err error) error {
	say("")
	say("%s", badStyle.Render("✗ It stopped: "+firstLine(err.Error())))
	if op := recipe.InProgress(a.Repo); op != "" {
		if recipe.Test(a.Repo, "has-conflicts") {
			say("There are conflicts. Fix them with %s, then %s.", cmdStyle.Render("gitwhy resolve"), cmdStyle.Render("gitwhy continue"))
		} else {
			say("A %s is in progress. Continue with %s.", op, cmdStyle.Render("gitwhy continue"))
		}
		say("Or cancel it with %s, or go back to before you started with %s.", cmdStyle.Render("gitwhy abort"), cmdStyle.Render("gitwhy undo"))
	} else {
		say("Nothing is lost: %s puts everything back to before this started.", cmdStyle.Render("gitwhy undo"))
	}
	return ErrCancelled
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "error:") || strings.HasPrefix(l, "fatal:") {
			return l
		}
	}
	if i := strings.IndexByte(s, '\n'); i > 0 {
		return s[:i]
	}
	return s
}
