package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/huh/v2"

	"github.com/dpologdvinity/gitwhy/internal/danger"
	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/kb"
	"github.com/dpologdvinity/gitwhy/internal/recipe"
	"github.com/dpologdvinity/gitwhy/internal/recover"
	"github.com/dpologdvinity/gitwhy/internal/search"
	"github.com/dpologdvinity/gitwhy/internal/snapshot"
)

// Home is the main menu.
func (a *App) Home() error {
	var v string
	err := prompt(huh.NewSelect[string]().Title("gitwhy: what do you want to do?").Options(
		huh.NewOption("Search: describe what you want (\"undo last commit\")", "search"),
		huh.NewOption("Browse by category", "browse"),
		huh.NewOption("What just happened? See recent changes and go back", "recover"),
		huh.NewOption("Undo the last thing gitwhy did", "undo"),
		huh.NewOption("Build any git command (every command and flag)", "build"),
	).Value(&v))
	if err != nil {
		return cancelled(err)
	}
	switch v {
	case "search":
		var q string
		if err := prompt(huh.NewInput().Title("What do you want to do?").Placeholder("e.g. move my last commit to a new branch").Value(&q)); err != nil {
			return cancelled(err)
		}
		return a.Search(q)
	case "browse":
		return a.Browse()
	case "recover":
		return a.Recover()
	case "undo":
		return a.Undo(false)
	case "build":
		return a.Build("", "")
	}
	return nil
}

// Search shows results for a query and runs the chosen one.
func (a *App) Search(q string) error {
	results := a.Index.Search(q, 15)
	if len(results) == 0 {
		say("Nothing matched %q. Try other words, or browse.", q)
		return a.Browse()
	}
	var opts []huh.Option[int]
	for i, r := range results {
		var label string
		switch r.Kind {
		case search.Recipe:
			label = "★ " + r.Title
		case search.Command:
			label = cmdStyle.Render(r.Title) + "  " + dimStyle.Render(trunc(r.Detail, 60))
		case search.Flag:
			label = tokStyle.Render(r.Title) + "  " + dimStyle.Render(trunc(r.Detail, 50))
		}
		opts = append(opts, huh.NewOption(label, i))
	}
	opts = append(opts, huh.NewOption(dimStyle.Render("Not here? Browse by category"), -1))
	var pick int
	if err := prompt(huh.NewSelect[int]().Title("Results for \"" + q + "\"").Description("★ = guided steps").Options(opts...).Value(&pick).Height(min(len(opts)+2, 18))); err != nil {
		return cancelled(err)
	}
	if pick < 0 {
		return a.Browse()
	}
	r := results[pick]
	switch r.Kind {
	case search.Recipe:
		return a.RunRecipe(r.Recipe)
	case search.Flag:
		return a.Build(r.Cmd, longest(r.Flag))
	default:
		return a.Build(r.Cmd, "")
	}
}

// Browse picks a recipe by category.
func (a *App) Browse() error {
	cats := map[string][]*recipe.Recipe{}
	for _, r := range a.Recipes {
		cats[r.Category] = append(cats[r.Category], r)
	}
	names := make([]string, 0, len(cats))
	for c := range cats {
		names = append(names, c)
	}
	sort.Strings(names)
	var opts []huh.Option[string]
	for _, c := range names {
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s (%d)", c, len(cats[c])), c))
	}
	opts = append(opts, huh.NewOption("Any git command… (build it yourself)", ""))
	var cat string
	if err := prompt(huh.NewSelect[string]().Title("Pick a category").Options(opts...).Value(&cat).Height(min(len(opts)+2, 18))); err != nil {
		return cancelled(err)
	}
	if cat == "" {
		return a.Build("", "")
	}
	var ropts []huh.Option[int]
	for i, r := range cats[cat] {
		ropts = append(ropts, huh.NewOption(r.Title, i))
	}
	var i int
	if err := prompt(huh.NewSelect[int]().Title(cat).Options(ropts...).Value(&i).Height(min(len(ropts)+2, 18))); err != nil {
		return cancelled(err)
	}
	return a.RunRecipe(cats[cat][i])
}

// longest picks the spelling to type for a flag: the positive long form if there is one.
func longest(f kb.Flag) string {
	for _, s := range f.Spellings {
		if strings.HasPrefix(s, "--") && !strings.HasPrefix(s, "--no-") {
			return s
		}
	}
	for _, s := range f.Spellings {
		if strings.HasPrefix(s, "--") {
			return s
		}
	}
	return f.Spellings[0]
}

// Build lets the user assemble any git command from the docs: command, subcommand, flags, args.
func (a *App) Build(cmd, preselect string) error {
	if cmd == "" {
		var opts []huh.Option[string]
		for _, n := range a.KB.Names() {
			if n == "git" {
				continue
			}
			opts = append(opts, huh.NewOption(fmt.Sprintf("%-18s %s", n, dimStyle.Render(trunc(a.KB.Commands[n].Summary, 70))), n))
		}
		if err := prompt(huh.NewSelect[string]().Title("Which git command?").Description("Type / to filter").Options(opts...).Value(&cmd).Height(18)); err != nil {
			return cancelled(err)
		}
	}
	c := a.KB.Commands[cmd]
	if c == nil {
		return fmt.Errorf("unknown command %q", cmd)
	}
	heading("git " + cmd)
	say("%s", dimStyle.Render(c.Summary))
	args := []string{cmd}
	sub := ""
	if len(c.Subs) > 0 {
		opts := []huh.Option[string]{huh.NewOption("(no subcommand)", "")}
		for _, s := range c.Subs {
			opts = append(opts, huh.NewOption(fmt.Sprintf("%-16s %s", s.Name, dimStyle.Render(trunc(s.Summary, 70))), s.Name))
		}
		if err := prompt(huh.NewSelect[string]().Title("Subcommand").Options(opts...).Value(&sub).Height(min(len(opts)+2, 18))); err != nil {
			return cancelled(err)
		}
		if sub != "" {
			args = append(args, sub)
		}
	}
	flags := a.KB.FlagsFor(cmd, sub)
	if len(flags) > 0 {
		var opts []huh.Option[int]
		for i, f := range flags {
			o := huh.NewOption(fmt.Sprintf("%-28s %s", trunc(f.Display, 28), dimStyle.Render(trunc(f.Desc, 80))), i)
			if preselect != "" && contains(f.Spellings, preselect) {
				o = o.Selected(true)
			}
			opts = append(opts, o)
		}
		var picked []int
		if err := prompt(huh.NewMultiSelect[int]().Title("Flags (optional)").Description("Space to select, / to filter, enter when done").Options(opts...).Value(&picked).Height(18)); err != nil {
			return cancelled(err)
		}
		for _, i := range picked {
			f := flags[i]
			name := longest(f)
			if f.Value == "" {
				args = append(args, name)
				continue
			}
			var v string
			in := huh.NewInput().Title(name + " " + f.Value).Description(f.Desc).Value(&v)
			if err := prompt(in); err != nil {
				return cancelled(err)
			}
			switch {
			case v == "" && f.Optional:
				args = append(args, name)
			case strings.HasPrefix(name, "--"):
				args = append(args, name+"="+v)
			default:
				args = append(args, name, v)
			}
		}
	}
	usage := strings.Join(c.Usage, "\n")
	if sub != "" && c.Sub(sub) != nil && len(c.Sub(sub).Usage) > 0 {
		usage = strings.Join(c.Sub(sub).Usage, "\n")
	}
	var rest string
	if err := prompt(huh.NewInput().Title("Anything else? (branches, commits, files)").Description(trunc(usage, 300)).Value(&rest)); err != nil {
		return cancelled(err)
	}
	more, err := shellSplit(rest)
	if err != nil {
		return err
	}
	return a.Run(append(args, more...))
}

// Run previews, explains, snapshots and runs a literal git command.
func (a *App) Run(args []string) error {
	if len(args) > 0 && args[0] == "git" {
		args = args[1:]
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: gitwhy run <git command>")
	}
	if a.Repo == nil || readOnly[args[0]] {
		// Nothing to undo: just explain and run.
		a.showCommand(args)
		if a.Repo == nil && args[0] != "clone" && args[0] != "init" {
			return a.needRepo()
		}
		r := a.Repo
		if r == nil {
			wd, _ := os.Getwd()
			return gitIn(wd, args)
		}
		return r.Interactive(args...)
	}
	return a.execute(job{title: "git " + shellJoin(args), literal: args})
}

// readOnly commands don't change the repo: no preview, confirm or snapshot.
var readOnly = map[string]bool{
	"status": true, "log": true, "diff": true, "show": true, "blame": true, "shortlog": true, "grep": true,
	"ls-files": true, "ls-tree": true, "ls-remote": true, "rev-parse": true, "rev-list": true, "describe": true,
	"cat-file": true, "help": true, "version": true, "whatchanged": true, "range-diff": true,
	"show-ref": true, "for-each-ref": true, "name-rev": true, "merge-base": true, "count-objects": true, "fsck": true,
	"check-ignore": true, "check-attr": true, "var": true, "cherry": true, "show-branch": true, "difftool": true,
}

// Explain prints the breakdown of a command without running it.
func (a *App) Explain(args []string) error {
	if len(args) > 0 && args[0] == "git" {
		args = args[1:]
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: gitwhy explain <git command>")
	}
	quiet := a.Quiet
	a.Quiet = false
	a.showCommand(args)
	a.Quiet = quiet
	for _, w := range dangerOf(args) {
		warn(w)
	}
	return nil
}

// Undo restores the latest snapshot, or one the user picks.
func (a *App) Undo(pick bool) error {
	if err := a.needRepo(); err != nil {
		return err
	}
	all, err := snapshot.List(a.Repo)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		say("gitwhy hasn't run anything here yet, so there's nothing to undo.")
		say("For things you did in plain git, try %s.", cmdStyle.Render("gitwhy recover"))
		return nil
	}
	s := all[0]
	if pick {
		var opts []huh.Option[int]
		for i, s := range all {
			opts = append(opts, huh.NewOption(fmt.Sprintf("%-10s %s", recover.Ago(s.Time), s.Label), i))
		}
		var i int
		if err := prompt(huh.NewSelect[int]().Title("Go back to before which action?").Description("Newest first").Options(opts...).Value(&i).Height(min(len(opts)+2, 18))); err != nil {
			return cancelled(err)
		}
		s = all[i]
	}
	heading("Undo")
	say("Go back to just before %s (%s).", cmdStyle.Render(s.Label), recover.Ago(s.Time))
	say("%s", dimStyle.Render("Branches, HEAD, staged and unstaged changes, untracked files and stashes are put back."))
	say("%s", dimStyle.Render("Anything sent to a remote (push) is not undone."))
	ok, err := confirm("Undo it?", true)
	if err != nil || !ok {
		return err
	}
	if _, err := snapshot.Restore(a.Repo, s); err != nil {
		return fmt.Errorf("undo failed: %w", err)
	}
	snapshot.Delete(a.Repo, s)
	say("%s Changed your mind? %s again redoes it.", okStyle.Render("✓ Undone."), cmdStyle.Render("gitwhy undo"))
	return nil
}

// Recover shows recent events (from plain git too) and lost work, and undoes one.
func (a *App) Recover() error {
	if err := a.needRepo(); err != nil {
		return err
	}
	evs, err := recover.Events(a.Repo, 25)
	if err != nil {
		return err
	}
	var opts []huh.Option[int]
	for i, e := range evs {
		opts = append(opts, huh.NewOption(fmt.Sprintf("%-9s %s", recover.Ago(e.Time), trunc(e.Text, 90)), i))
	}
	opts = append(opts, huh.NewOption(dimStyle.Render("Find lost commits and staged files (deleted branches, hard resets)"), -1))
	var i int
	if err := prompt(huh.NewSelect[int]().Title("What happened recently (newest first)").Description("Pick one to go back to before it").Options(opts...).Value(&i).Height(min(len(opts)+2, 20))); err != nil {
		return cancelled(err)
	}
	if i < 0 {
		return a.lost()
	}
	e := evs[i]
	undo := e.Undo(a.Repo)
	if undo == nil {
		say("There's no earlier state recorded for this.")
		return nil
	}
	return a.execute(job{title: "Undo: " + e.Text, literal: undo,
		predict: "Goes back to before: " + e.Text + ". Uncommitted changes are kept; git stops if one would be overwritten."})
}

func (a *App) lost() error {
	say("%s", dimStyle.Render("Searching (can take a while in big repos)..."))
	ls, err := recover.LostObjects(a.Repo, 40)
	if err != nil {
		return err
	}
	if len(ls) == 0 {
		say("No lost commits or staged files found.")
		return nil
	}
	var opts []huh.Option[int]
	for i, l := range ls {
		if l.Kind == "commit" {
			opts = append(opts, huh.NewOption(fmt.Sprintf("commit %s  %s  %s", l.ID[:7], trunc(l.Title, 60), dimStyle.Render(recover.Ago(l.Time))), i))
		} else {
			opts = append(opts, huh.NewOption(fmt.Sprintf("file   %s  %s", l.ID[:7], dimStyle.Render(trunc(l.Title, 70))), i))
		}
	}
	var i int
	if err := prompt(huh.NewSelect[int]().Title("Lost work git still has").Options(opts...).Value(&i).Height(min(len(opts)+2, 20))); err != nil {
		return cancelled(err)
	}
	l := ls[i]
	if l.Kind == "commit" {
		name := "recovered-" + l.ID[:7]
		if err := prompt(huh.NewInput().Title("Name for a branch at this commit").Value(&name)); err != nil {
			return cancelled(err)
		}
		return a.execute(job{title: "Recover commit", literal: []string{"branch", name, l.ID},
			predict: "Creates branch " + name + " at the lost commit, so it's safe again."})
	}
	path := "recovered-" + l.ID[:7] + ".txt"
	if err := prompt(huh.NewInput().Title("Save the file as").Value(&path)); err != nil {
		return cancelled(err)
	}
	data, err := a.Repo.OutputIn(nil, "cat-file", "blob", l.ID)
	if err != nil {
		return err
	}
	full := filepath.Join(a.Repo.Top, path)
	if _, err := os.Stat(full); err == nil {
		return fmt.Errorf("%s already exists; pick another name", path)
	}
	if err := os.WriteFile(full, []byte(data), 0o644); err != nil {
		return err
	}
	say("%s Saved to %s", okStyle.Render("✓"), path)
	return nil
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

func dangerOf(args []string) []string { return danger.Warnings(args) }

func gitIn(dir string, args []string) error { return git.Runner{Dir: dir}.Interactive(args...) }
