package ui

import (
	"fmt"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/danger"
	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/preview"
	"github.com/dpologdvinity/gitwhy/internal/recipe"
	"github.com/dpologdvinity/gitwhy/internal/snapshot"
	"github.com/dpologdvinity/gitwhy/internal/todo"
)

// job is what will run: recipe steps with answers, or one literal command.
type job struct {
	title   string         // names the job; the snapshot label when there are several steps
	r       *recipe.Recipe // nil for a literal command
	steps   []recipe.Step
	ans     recipe.Answers // answers so far; mid-recipe questions add to it while running
	literal []string       // args of a literal command (not templated)
	predict string         // recipe text shown when no exact preview exists
}

// planned is a step ready to show: its args with later answers as placeholders.
type planned struct {
	step recipe.Step
	args []string
}

// stopError is a recipe "fail" step: shown plainly, nothing ran.
type stopError struct{ msg string }

// Error returns the rendered "fail" message.
func (e stopError) Error() string { return e.msg }

// plan turns a job into the commands to show and the recipe warnings to add. A literal
// command is one step as typed; recipe steps go through planSteps.
func (a *App) plan(j job) ([]planned, []string, error) {
	if j.literal != nil {
		return []planned{{args: j.literal}}, nil, nil
	}
	return planSteps(j.steps, j.ans)
}

// planSteps resolves recipe steps with the answers so far: questions asked later show as
// <name>, "when" is honored, a true "fail" stops with stopError and "warn" adds a warning.
func planSteps(steps []recipe.Step, ans recipe.Answers) ([]planned, []string, error) {
	// Work on a copy so the placeholders never leak into the real answers.
	shown := recipe.Answers{}
	for k, v := range ans {
		shown[k] = v
	}
	for _, s := range steps {
		if s.Ask != "" {
			if _, ok := shown[s.Ask]; !ok {
				shown[s.Ask] = "<" + s.Ask + ">"
			}
		}
	}
	var out []planned
	var warns []string
	for _, s := range steps {
		if s.Ask != "" {
			continue
		}
		ok, err := recipe.Truthy(s.When, shown)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			continue
		}
		if s.Fail != "" {
			msg, _ := recipe.Render(s.Fail, shown)
			return nil, nil, stopError{msg}
		}
		if s.Warn != "" {
			msg, _ := recipe.Render(s.Warn, shown)
			warns = append(warns, msg)
			continue
		}
		args, err := recipe.Args(s, shown)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, planned{step: s, args: args})
	}
	return out, warns, nil
}

// runner returns where a job's commands run: recipes at the repo root (their file paths are
// root-relative), literal commands in the folder the user ran gitwhy from.
func (a *App) runner(j job) *git.Repo {
	r := *a.Repo
	if j.literal != nil {
		if a.Cwd != "" {
			r.Dir = a.Cwd
		}
		r.Runner = typed(r.Runner)
	}
	return &r
}

// execute previews, explains, confirms, snapshots and runs a job. Every recipe and every
// "gitwhy run" ends here. In order:
//
//  1. plan resolves the steps; a true "fail" step stops here and nothing runs.
//  2. "What will happen": for one step, the preview's lines. When the preview is not exact,
//     the recipe's predict text follows, or else (one step, no predict) the command's summary
//     from the docs.
//  3. Warnings from the recipe's "warn" steps and from the danger rules, without repeats.
//  4. Each command, explained part by part.
//  5. "Run it?" defaults to No when any warning was shown, and to Yes otherwise.
//  6. snapshot.Take saves an undo point, with the ignored files the preview says git will
//     delete. If it fails, nothing runs.
//  7. The steps run in order; mid-recipe questions are asked when their step is reached.
//  8. snapshot.Finish runs in a defer, so the "after" refs are recorded even when a step
//     fails. Undo uses them to leave alone refs that moved later.
//
// When a step fails or a mid-recipe question is cancelled, failed explains what to do next.
func (a *App) execute(j job) error {
	if err := a.needRepo(); err != nil {
		return err
	}
	steps, extraWarn, err := a.plan(j)
	if se, ok := err.(stopError); ok {
		say("%s", badStyle.Render("✗ "+se.msg))
		say("Nothing ran.")
		return ErrCancelled
	}
	if err != nil {
		return err
	}
	where := a.runner(j)
	if len(steps) == 0 {
		say("Nothing to do.")
		return nil
	}

	// What will happen.
	heading("What will happen")
	var save []string // ignored files git is about to delete; the snapshot keeps them
	exact := false
	// Exact previews only exist for one step: later steps depend on what earlier ones do.
	if len(steps) == 1 {
		p := preview.Preview(where, steps[0].args)
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
	warnings := extraWarn
	seen := map[string]bool{}
	for _, s := range steps {
		for _, w := range danger.WarningsIn(where, s.args) {
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

	// Default to No when there is a warning, so pressing Enter never runs something risky.
	ok, err := Confirm("Run it?", len(warnings) == 0)
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
	// The snapshot is taken of the whole repository (a.Repo), not the folder the command runs in.
	snap, err := snapshot.Take(a.Repo, label, save)
	if err != nil {
		return fmt.Errorf("couldn't save an undo point, so nothing ran: %w", err)
	}
	// Record the refs after the command, even if it failed, so undo knows what it changed.
	defer func() {
		if err := snapshot.Finish(a.Repo, snap); err != nil {
			warn("Couldn't record the result for undo (" + err.Error() + "); undo will restore everything, including later changes.")
		}
	}()

	if j.literal != nil {
		if err := a.runOne(where, recipe.Step{}, j.literal, j.ans); err != nil {
			return a.failed(err, nil)
		}
		a.done()
		return nil
	}
	for i, s := range j.steps {
		if s.Ask != "" {
			q := j.r.Question(s.Ask)
			if err := a.ask(q, j.ans); err != nil {
				delete(j.ans, s.Ask) // not answered: show it as <name>
				return a.failed(err, notRun(j.steps[i:], j.ans))
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
		// "fail" and "warn" were handled by plan before the prompt; they run nothing.
		if s.Fail != "" || s.Warn != "" {
			continue
		}
		if err := a.runOne(where, s, args, j.ans); err != nil {
			return a.failed(err, notRun(j.steps[i+1:], j.ans))
		}
	}
	a.done()
	return nil
}

// notRun lists the commands of the steps that would still have run, for the advice after a
// recipe stops part way. Answers not given yet show as <name>.
func notRun(steps []recipe.Step, ans recipe.Answers) [][]string {
	ps, _, _ := planSteps(steps, ans)
	var out [][]string
	for _, p := range ps {
		out = append(out, p.args)
	}
	return out
}

// runOne prints and runs one git command in where, attached to the terminal so git can ask
// for input or open an editor. The step's env is added, plus GIT_SEQUENCE_EDITOR for a todo step.
func (a *App) runOne(where *git.Repo, s recipe.Step, args []string, ans recipe.Answers) error {
	env := append([]string{}, s.Env...)
	if len(s.Todo) > 0 {
		e, err := a.todoEnv(s, ans)
		if err != nil {
			return err
		}
		env = append(env, e)
	}
	say("%s", dimStyle.Render("$ git "+shellJoin(args)))
	return where.With(env...).Interactive(args...)
}

// todoEnv returns the environment entry that makes git call "gitwhy __todo" as its rebase
// sequence editor. A "plan" entry points at a plan file the recipe already has; otherwise the
// step's todo edits are rendered and written to a new plan file.
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

// done reports success and how to undo it.
func (a *App) done() {
	say("")
	say("%s Undo with: %s", okStyle.Render("✓ Done."), cmdStyle.Render("gitwhy undo"))
}

// failed explains a step that stopped and what to do next. left lists the commands of the
// steps that did not run: when some are left, "gitwhy continue" would finish only git's part,
// so the advice is to go back with "gitwhy undo" instead. The advice, first match wins:
//
//   - a bisect is running: end it with "gitwhy abort" (bisect has no --continue);
//   - steps are left: list them and advise "gitwhy undo" (and "gitwhy resolve" for conflicts);
//   - conflicts inside a merge, rebase, cherry-pick or revert: resolve, then continue, or
//     abort or undo;
//   - conflicts with no operation (such as "stash pop"): resolve, then commit, or undo;
//   - an operation in progress without conflicts: continue, abort or undo;
//   - otherwise: undo puts everything back.
//
// It always returns ErrCancelled, because the error has already been shown.
func (a *App) failed(err error, left [][]string) error {
	say("")
	say("%s", badStyle.Render("✗ It stopped: "+firstLine(err.Error())))
	op := recipe.InProgress(a.Repo)
	conflicts := recipe.Test(a.Repo, "has-conflicts")
	undo, resolve := cmdStyle.Render("gitwhy undo"), cmdStyle.Render("gitwhy resolve")
	switch {
	case op == "bisect": // "git bisect" has no --continue; HEAD stays on a commit under test
		say("A bisect is still running, so you are not back on your branch. End it with %s.", cmdStyle.Render("gitwhy abort"))
	case len(left) > 0:
		if conflicts {
			say("There are conflicts (%s can fix them).", resolve)
		}
		say("These steps did not run:")
		for _, args := range left {
			say("  %s", cmdStyle.Render("git "+shellJoin(args)))
		}
		say("Go back to before this started with %s, then try again.", undo)
	case conflicts && op != "":
		say("There are conflicts. Fix them with %s, then %s.", resolve, cmdStyle.Render("gitwhy continue"))
		say("Or cancel it with %s, or go back to before you started with %s.", cmdStyle.Render("gitwhy abort"), undo)
	case conflicts:
		say("There are conflicts. Fix them with %s, then commit. Or go back to before this started with %s.", resolve, undo)
	case op != "":
		say("A %s is in progress. Continue with %s.", op, cmdStyle.Render("gitwhy continue"))
		say("Or cancel it with %s, or go back to before you started with %s.", cmdStyle.Render("gitwhy abort"), undo)
	default:
		say("Nothing is lost: %s puts everything back to before this started.", undo)
	}
	return ErrCancelled
}

// firstLine picks the line of git's output worth showing: the first "error:" or "fatal:"
// line, or else the first line.
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
