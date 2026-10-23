package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"charm.land/huh/v2"

	"github.com/dpologdvinity/gitwhy/internal/recipe"
	"github.com/dpologdvinity/gitwhy/internal/todo"
)

// builtins are answers every recipe can use.
func (a *App) builtins() recipe.Answers {
	r := a.Repo
	ans := recipe.Answers{}
	ans["head"], _ = r.Output("rev-parse", "-q", "--verify", "HEAD")
	if b, err := r.Output("symbolic-ref", "--short", "-q", "HEAD"); err == nil {
		ans["branch"] = b
	} else {
		ans["branch"] = "HEAD"
	}
	ans["upstream"], _ = r.Output("rev-parse", "--abbrev-ref", "@{upstream}")
	remote, _ := r.Output("config", "branch."+ans["branch"].(string)+".remote")
	if remote == "" {
		remotes, _ := r.Output("remote")
		if strings.Contains("\n"+remotes+"\n", "\norigin\n") {
			remote = "origin"
		} else if remotes != "" {
			remote = strings.Split(remotes, "\n")[0]
		} else {
			remote = "origin"
		}
	}
	ans["remote"] = remote
	ans["op"] = recipe.InProgress(r)
	return ans
}

// RunRecipe walks the user through a recipe and runs it.
func (a *App) RunRecipe(r *recipe.Recipe) error {
	if err := a.needRepo(); err != nil {
		return err
	}
	ans := a.builtins()
	heading(r.Title)
	for _, c := range r.Checks {
		if !recipe.Test(a.Repo, c.If) {
			continue
		}
		if c.Block != "" {
			say("%s", badStyle.Render("✗ "+c.Block))
			return a.offer(c.Suggest)
		}
		warn(c.Warn)
		if c.Suggest != "" {
			if s := a.recipe(c.Suggest); s != nil {
				use, err := confirm("Do \""+s.Title+"\" instead?", false)
				if err != nil {
					return err
				}
				if use {
					return a.RunRecipe(s)
				}
			}
		}
	}
	if r.Explain != "" {
		if text, err := recipe.Render(r.Explain, ans); err == nil {
			say("%s", dimStyle.Render(strings.TrimSpace(text)))
		}
	}
	for i := range r.Questions {
		q := &r.Questions[i]
		if r.AskedLater(q.ID) {
			continue
		}
		if err := a.ask(q, ans); err != nil {
			return err
		}
	}
	predict, _ := recipe.Render(r.Predict, ans)
	return a.execute(job{title: r.Title, r: r, steps: r.Steps, ans: ans, predict: predict})
}

func (a *App) offer(id string) error {
	if id == "" {
		return ErrCancelled
	}
	s := a.recipe(id)
	if s == nil {
		return ErrCancelled
	}
	use, err := confirm("Do \""+s.Title+"\" instead?", true)
	if err != nil || !use {
		return err
	}
	return a.RunRecipe(s)
}

// ask asks one question and stores the answer (plus derived answers) in ans.
func (a *App) ask(q *recipe.Question, ans recipe.Answers) error {
	if ok, err := recipe.Truthy(q.When, ans); err != nil || !ok {
		if err == nil {
			ans[q.ID] = ""
		}
		return err
	}
	title, _ := recipe.Render(q.Ask, ans)
	def, _ := recipe.Render(q.Default, ans)
	var err error
	switch q.Type {
	case "choice":
		v := def
		var opts []huh.Option[string]
		for _, o := range q.Options {
			label := o.Label
			if o.Danger != "" {
				label += "  ⚠ " + o.Danger
			}
			opts = append(opts, huh.NewOption(label, o.Value))
		}
		err = prompt(huh.NewSelect[string]().Title(title).Description(q.Help).Options(opts...).Value(&v))
		ans[q.ID] = v
	case "confirm":
		v := def == "true"
		err = prompt(huh.NewConfirm().Title(title).Description(q.Help).Affirmative("Yes").Negative("No").Value(&v))
		ans[q.ID] = strconv.FormatBool(v)
	case "text":
		v := def
		in := huh.NewInput().Title(title).Description(q.Help).Value(&v)
		if !q.Optional {
			in = in.Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("required")
				}
				return nil
			})
		}
		err = prompt(in)
		if q.Split {
			ans[q.ID] = strings.Fields(v)
		} else {
			ans[q.ID] = strings.TrimSpace(v)
		}
	case "branch":
		err = a.askBranch(q, title, def, ans)
	case "commit":
		err = a.askCommit(q, title, ans)
	case "files":
		err = a.askFiles(q, title, ans)
	case "remote":
		v := def
		remotes, _ := a.Repo.Output("remote")
		var opts []huh.Option[string]
		for _, r := range strings.Split(remotes, "\n") {
			if r != "" {
				opts = append(opts, huh.NewOption(r, r))
			}
		}
		if len(opts) == 0 {
			return fmt.Errorf("this repository has no remotes")
		}
		err = prompt(huh.NewSelect[string]().Title(title).Options(opts...).Value(&v))
		ans[q.ID] = v
	case "todo":
		err = a.askTodo(q, title, ans)
	default:
		err = fmt.Errorf("unknown question type %q", q.Type)
	}
	return cancelled(err)
}

func (a *App) askBranch(q *recipe.Question, title, def string, ans recipe.Answers) error {
	var refs []string
	switch q.Source {
	case "local":
		refs = []string{"refs/heads"}
	case "remote":
		refs = []string{"refs/remotes"}
	default:
		refs = []string{"refs/heads", "refs/remotes"}
	}
	out, err := a.Repo.Output(append([]string{"for-each-ref", "--sort=-committerdate", "--format=%(refname)%00%(committerdate:relative)%00%(subject)"}, refs...)...)
	if err != nil {
		return err
	}
	var opts []huh.Option[string]
	found := false
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\x00")
		if len(f) != 3 || strings.HasSuffix(f[0], "/HEAD") {
			continue
		}
		f[0] = strings.TrimPrefix(strings.TrimPrefix(f[0], "refs/heads/"), "refs/remotes/")
		if f[0] == def {
			found = true
		}
		opts = append(opts, huh.NewOption(fmt.Sprintf("%-30s %s", f[0], dimStyle.Render(f[1]+" · "+trunc(f[2], 50))), f[0]))
	}
	if len(opts) == 0 {
		return fmt.Errorf("no branches to choose from")
	}
	v := def
	if !found {
		v = ""
	}
	err = prompt(huh.NewSelect[string]().Title(title).Description("Type / to filter").Options(opts...).Value(&v).Filtering(false).Height(min(len(opts)+2, 15)))
	ans[q.ID] = v
	return err
}

func (a *App) askCommit(q *recipe.Question, title string, ans recipe.Answers) error {
	src := "HEAD"
	if q.Source != "" {
		src, _ = recipe.Render(q.Source, ans)
	}
	out, err := a.Repo.Output("log", "-n", "60", "--format=%H%x00%h %s%x00%cr, %an", src, "--")
	if err != nil {
		return err
	}
	if out == "" {
		return fmt.Errorf("no commits to choose from in %s", src)
	}
	var opts []huh.Option[string]
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\x00")
		if len(f) == 3 {
			opts = append(opts, huh.NewOption(trunc(f[1], 70)+"  "+dimStyle.Render(f[2]), f[0]))
		}
	}
	var v string
	if err := prompt(huh.NewSelect[string]().Title(title).Description(q.Help).Options(opts...).Value(&v).Height(min(len(opts)+2, 15))); err != nil {
		return err
	}
	ans[q.ID] = v
	if parent, err := a.Repo.Output("rev-parse", "-q", "--verify", v+"^"); err == nil {
		ans[q.ID+"_base"] = parent
	} else {
		ans[q.ID+"_base"] = "--root"
	}
	files, _ := a.Repo.Output("diff-tree", "--no-commit-id", "--name-only", "-r", "--root", "-z", v)
	ans[q.ID+"_files"] = splitZ(files)
	return nil
}

func (a *App) askFiles(q *recipe.Question, title string, ans recipe.Answers) error {
	var out string
	var err error
	switch q.Source {
	case "staged":
		out, err = a.Repo.Output("diff", "--cached", "--name-only", "-z")
	case "unstaged":
		out, err = a.Repo.Output("diff", "--name-only", "-z")
	case "conflicted":
		out, err = a.Repo.Output("diff", "--name-only", "--diff-filter=U", "-z")
	case "untracked":
		out, err = a.Repo.Output("ls-files", "--others", "--exclude-standard", "-z")
	case "tracked":
		out, err = a.Repo.Output("ls-files", "-z")
	default: // changed: staged, unstaged and untracked
		out, err = a.Repo.Output("status", "--porcelain=v1", "-z", "--untracked-files=all")
		if err == nil {
			out = statusPaths(out)
		}
	}
	if err != nil {
		return err
	}
	files := splitZ(out)
	if len(files) == 0 {
		return fmt.Errorf("no matching files")
	}
	if len(files) > 2000 { // too many to list (monorepo): type the path
		var v string
		err := prompt(huh.NewInput().Title(title).Description("Path relative to the repo root").Value(&v))
		if q.Multi {
			ans[q.ID] = strings.Fields(v)
		} else {
			ans[q.ID] = v
		}
		return err
	}
	var opts []huh.Option[string]
	for _, f := range files {
		opts = append(opts, huh.NewOption(f, f))
	}
	if !q.Multi {
		var v string
		err := prompt(huh.NewSelect[string]().Title(title).Options(opts...).Value(&v).Height(min(len(opts)+2, 15)))
		ans[q.ID] = v
		return err
	}
	var v []string
	err = prompt(huh.NewMultiSelect[string]().Title(title).Description("Space to select, enter when done").
		Options(opts...).Value(&v).Height(min(len(opts)+2, 15)).
		Validate(func(s []string) error {
			if len(s) == 0 {
				return fmt.Errorf("pick at least one")
			}
			return nil
		}))
	ans[q.ID] = v
	return err
}

// statusPaths extracts paths from "git status --porcelain=v1 -z" (renames have two paths).
func statusPaths(out string) string {
	var paths []string
	f := strings.Split(out, "\x00")
	for i := 0; i < len(f); i++ {
		e := f[i]
		if len(e) < 4 {
			continue
		}
		paths = append(paths, e[3:])
		if e[0] == 'R' || e[0] == 'C' {
			i++ // skip the original path
		}
	}
	return strings.Join(paths, "\x00")
}

func splitZ(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "\x00") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func trunc(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

type todoItem struct {
	todo.Item
	short, subject string
}

// askTodo lets the user reorder, drop, squash, fixup or reword recent commits.
func (a *App) askTodo(q *recipe.Question, title string, ans recipe.Answers) error {
	n := "10"
	if c, err := a.Repo.Output("rev-list", "--count", "@{upstream}..HEAD"); err == nil && c != "0" {
		if v, _ := strconv.Atoi(c); v <= 30 {
			n = c
		}
	}
	if err := prompt(huh.NewInput().Title("How many recent commits?").Value(&n).Validate(func(s string) error {
		if v, err := strconv.Atoi(s); err != nil || v < 1 || v > 200 {
			return fmt.Errorf("a number from 1 to 200")
		}
		return nil
	})); err != nil {
		return err
	}
	out, err := a.Repo.Output("log", "--first-parent", "--reverse", "--format=%H%x00%h%x00%s%x00%P", "-n", n, "HEAD")
	if err != nil {
		return err
	}
	var items []todoItem
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\x00")
		if len(f) != 4 {
			continue
		}
		if strings.Contains(f[3], " ") {
			return fmt.Errorf("those commits include a merge commit; this editor can't rewrite merges safely")
		}
		items = append(items, todoItem{Item: todo.Item{Action: "pick", Commit: f[0]}, short: f[1], subject: f[2]})
	}
	if len(items) == 0 {
		return fmt.Errorf("no commits")
	}
	allowed := map[string]bool{}
	for _, act := range q.Actions {
		allowed[act] = true
	}
	for {
		var pick string
		var opts []huh.Option[string]
		for i, it := range items {
			label := fmt.Sprintf("%2d. %-7s %s %s", i+1, it.Action, it.short, trunc(it.subject, 55))
			if it.Action == "reword" {
				label += dimStyle.Render("  → " + trunc(it.Message, 30))
			}
			opts = append(opts, huh.NewOption(label, strconv.Itoa(i)))
		}
		opts = append(opts, huh.NewOption(okStyle.Render("✓ Done"), "done"), huh.NewOption("✗ Cancel", "cancel"))
		if err := prompt(huh.NewSelect[string]().Title(title).Description("Top = oldest. " + q.Help).Options(opts...).Value(&pick).Height(min(len(opts)+2, 20))); err != nil {
			return err
		}
		if pick == "cancel" {
			return huh.ErrUserAborted
		}
		if pick == "done" {
			if a := items[0].Action; a == "squash" || a == "fixup" {
				warn("The first (oldest) commit can't be squashed: there's nothing above it.")
				continue
			}
			break
		}
		i, _ := strconv.Atoi(pick)
		var act string
		acts := []huh.Option[string]{huh.NewOption("Move up (earlier)", "up"), huh.NewOption("Move down (later)", "down"), huh.NewOption("Keep as is", "pick")}
		for _, x := range []struct{ id, label string }{
			{"drop", "Drop (delete this commit and its changes)"},
			{"squash", "Squash into the commit above (combine messages)"},
			{"fixup", "Fixup into the commit above (keep its message)"},
			{"reword", "Change its message"},
		} {
			if allowed[x.id] {
				acts = append(acts, huh.NewOption(x.label, x.id))
			}
		}
		if err := prompt(huh.NewSelect[string]().Title(items[i].short + " " + items[i].subject).Options(acts...).Value(&act)); err != nil {
			return err
		}
		switch act {
		case "up":
			if i > 0 {
				items[i-1], items[i] = items[i], items[i-1]
			}
		case "down":
			if i+1 < len(items) {
				items[i+1], items[i] = items[i], items[i+1]
			}
		case "reword":
			msg := items[i].subject
			if err := prompt(huh.NewInput().Title("New message").Value(&msg)); err != nil {
				return err
			}
			items[i].Action, items[i].Message = "reword", msg
		default:
			items[i].Action = act
		}
	}
	var p todo.Plan
	for _, it := range items {
		p.Full = append(p.Full, it.Item)
	}
	path, err := todo.Write(p)
	if err != nil {
		return err
	}
	ans[q.ID] = path
	oldest := ""
	for _, line := range strings.Split(out, "\n") { // original order: first line is the oldest
		oldest = strings.Split(line, "\x00")[0]
		break
	}
	if parent, err := a.Repo.Output("rev-parse", "-q", "--verify", oldest+"^"); err == nil {
		ans[q.ID+"_base"] = parent
	} else {
		ans[q.ID+"_base"] = "--root"
	}
	return nil
}

// RunRecipeID runs a recipe by its ID.
func (a *App) RunRecipeID(id string) error {
	r := a.recipe(id)
	if r == nil {
		return fmt.Errorf("no recipe %q", id)
	}
	return a.RunRecipe(r)
}

// ListRecipes prints every recipe ID and title, by category.
func (a *App) ListRecipes() error {
	byCat := map[string][]string{}
	var cats []string
	for _, r := range a.Recipes {
		if _, ok := byCat[r.Category]; !ok {
			cats = append(cats, r.Category)
		}
		byCat[r.Category] = append(byCat[r.Category], fmt.Sprintf("  %-28s %s", r.ID, r.Title))
	}
	sort.Strings(cats)
	for _, c := range cats {
		heading(c)
		for _, l := range byCat[c] {
			say("%s", l)
		}
	}
	say("\nRun one with: gitwhy do <name>")
	return nil
}
