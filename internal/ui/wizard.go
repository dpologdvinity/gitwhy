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
	locals, _ := r.Output("for-each-ref", "--format=%(refname)", "refs/heads", "refs/tags")
	ans["local_refs"] = strings.Fields(locals)
	return ans
}

// RunRecipe walks the user through a recipe and runs it.
func (a *App) RunRecipe(r *recipe.Recipe) error { return a.RunRecipeWith(r, nil) }

// RunRecipeWith runs a recipe with some answers already given (e.g. the commit picked in the TUI);
// those questions are skipped.
func (a *App) RunRecipeWith(r *recipe.Recipe, preset recipe.Answers) error {
	if err := a.needRepo(); err != nil {
		return err
	}
	ans := a.builtins()
	for k, v := range preset {
		ans[k] = v
		if q := r.Question(k); q != nil && q.Type == "commit" {
			a.deriveCommit(q.ID, v.(string), ans)
		} else if q != nil && q.Type == "branch" {
			a.deriveBranch(q.ID, v.(string), ans)
		} else if q != nil && q.Type == "files" {
			a.deriveFiles(q.ID, v, ans)
		}
	}
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
				use, err := Confirm("Do \""+s.Title+"\" instead?", false)
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
		if _, given := preset[q.ID]; given || r.AskedLater(q.ID) {
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
	use, err := Confirm("Do \""+s.Title+"\" instead?", true)
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
	a.deriveBranch(q.ID, v, ans)
	return err
}

// deriveBranch sets a branch answer plus where it pushes/pulls: _remote and _merge (remote branch).
// It also sets _local, the local branch to switch to, and _track, the remote-tracking branch to
// make _local from when no local branch of that name exists yet ("" otherwise). git refuses to
// switch to a remote branch such as origin/feat, and checking one out would detach HEAD.
func (a *App) deriveBranch(id, v string, ans recipe.Answers) {
	ans[id] = v
	ans[id+"_remote"], _ = a.Repo.Output("config", "branch."+v+".remote")
	merge, _ := a.Repo.Output("config", "branch."+v+".merge")
	ans[id+"_merge"] = strings.TrimPrefix(merge, "refs/heads/")
	local, track := v, ""
	isRef := func(ref string) bool { return a.Repo.OK("show-ref", "--verify", "-q", ref) }
	if v != "" && !isRef("refs/heads/"+v) {
		remotes, _ := a.Repo.Output("remote")
		for _, r := range strings.Fields(remotes) {
			if name, ok := strings.CutPrefix(v, r+"/"); ok && isRef("refs/remotes/"+v) {
				local, track = name, v // picked as origin/feat
			} else if isRef("refs/remotes/" + r + "/" + v) {
				if track != "" { // on several remotes: let git say which to pick
					local, track = v, ""
					break
				}
				track = r + "/" + v // given as feat, as the TUI does
			}
		}
		if track != "" && isRef("refs/heads/"+local) {
			track = "" // a local branch of that name already exists: use it
		}
	}
	ans[id+"_local"], ans[id+"_track"] = local, track
}

func (a *App) askCommit(q *recipe.Question, title string, ans recipe.Answers) error {
	src := "HEAD"
	if q.Source != "" {
		src, _ = recipe.Render(q.Source, ans)
	}
	out, err := a.Repo.Output("log", "--first-parent", "-n", "60", "--format=%H%x00%h %s%x00%cr, %an", src, "--")
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
	a.deriveCommit(q.ID, v, ans)
	return nil
}

// deriveCommit sets a commit answer and what recipes derive from it: _base (parent or --root),
// _files (files it changed) and _pushed.
func (a *App) deriveCommit(id, v string, ans recipe.Answers) {
	ans[id] = v
	if parent, err := a.Repo.Output("rev-parse", "-q", "--verify", v+"^"); err == nil {
		ans[id+"_base"] = parent
	} else {
		ans[id+"_base"] = "--root"
	}
	files, _ := a.Repo.Output("diff-tree", "--no-commit-id", "--name-only", "-r", "--root", "-z", v)
	ans[id+"_files"] = splitZ(files)
	ans[id+"_pushed"] = a.pushed(v)
}

// pushed reports ("true"/"false") whether a commit is on any remote-tracking branch.
func (a *App) pushed(commit string) string {
	out, _ := a.Repo.Output("branch", "-r", "--contains", commit)
	return fmt.Sprint(strings.TrimSpace(out) != "")
}

func (a *App) askFiles(q *recipe.Question, title string, ans recipe.Answers) error {
	var out string
	var err error
	var files []string
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
		files, _ = statusPaths(out)
	}
	if err != nil {
		return err
	}
	if files == nil {
		files = splitZ(out)
	}
	if len(files) == 0 {
		return fmt.Errorf("no matching files")
	}
	if len(files) > 2000 { // too many to list (monorepo): type the path
		var v string
		err := prompt(huh.NewInput().Title(title).Description("Path relative to the repo root").Value(&v))
		if q.Multi { // one path, as typed: names can contain spaces
			a.deriveFiles(q.ID, []string{v}, ans)
		} else {
			a.deriveFiles(q.ID, v, ans)
		}
		return err
	}
	from := a.renames()
	var opts []huh.Option[string]
	for _, f := range files {
		label := f
		if old, ok := from[f]; ok {
			label += "  (renamed from " + old + ")"
		}
		opts = append(opts, huh.NewOption(label, f))
	}
	if !q.Multi {
		var v string
		err := prompt(huh.NewSelect[string]().Title(title).Options(opts...).Value(&v).Height(min(len(opts)+2, 15)))
		a.deriveFiles(q.ID, v, ans)
		return err
	}
	var v []string
	err = prompt(huh.NewMultiSelect[string]().Title(title).Description("Space to select, enter when done").
		Options(opts...).Value(&v).Height(min(len(opts)+2, 15)).
		Validate(func(s []string) error {
			if len(s) == 0 {
				return fmt.Errorf("pick at least one")
			}
			if q.Partial && len(s) == len(files) {
				return fmt.Errorf("leave at least one file for the other part")
			}
			return nil
		}))
	a.deriveFiles(q.ID, v, ans)
	return err
}

// deriveFiles sets a files answer (a path or a list of paths) and <id>_from: the original paths
// of the picked files that are staged renames. A commit limited to the new path alone would add
// a copy and leave the deletion of the original staged, so recipes that commit paths pass both.
func (a *App) deriveFiles(id string, v any, ans recipe.Answers) {
	ans[id] = v
	paths, _ := v.([]string)
	if s, ok := v.(string); ok {
		paths = []string{s}
	}
	from := a.renames()
	var olds []string // a typed nil list still splats to no arguments
	for _, p := range paths {
		if old, ok := from[p]; ok {
			olds = append(olds, old)
		}
	}
	ans[id+"_from"] = olds
}

// renames maps the new path of each staged rename to its original path.
func (a *App) renames() map[string]string {
	out, _ := a.Repo.Output("status", "--porcelain=v1", "-z", "--untracked-files=no")
	_, from := statusPaths(out)
	return from
}

// statusPaths parses "git status --porcelain=v1 -z": the path of each changed file, and for
// renames a map from the new path to the original one (copies keep their original unchanged).
func statusPaths(out string) ([]string, map[string]string) {
	var paths []string
	from := map[string]string{}
	f := strings.Split(out, "\x00")
	for i := 0; i < len(f); i++ {
		e := f[i]
		if len(e) < 4 {
			continue
		}
		paths = append(paths, e[3:])
		if (e[0] == 'R' || e[0] == 'C') && i+1 < len(f) {
			i++ // the original path is the next field
			if e[0] == 'R' {
				from[e[3:]] = f[i]
			}
		}
	}
	return paths, from
}

func splitZ(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "\x00") {
		if p != "" { // never trim: file names can start or end with spaces
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
			if i := orphanSquash(items); i >= 0 {
				warn(fmt.Sprintf("Commit %d (%s) can't be squashed: no kept commit is above it.", i+1, items[i].short))
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
	ans[q.ID+"_pushed"] = a.pushed(oldest)
	return nil
}

// orphanSquash returns the index of the first squash or fixup with no kept (not dropped) commit
// above it, or -1. git stops on such a plan with "cannot 'squash' without a previous commit".
func orphanSquash(items []todoItem) int {
	kept := false
	for i, it := range items {
		switch it.Action {
		case "squash", "fixup":
			if !kept {
				return i
			}
		case "drop":
		default:
			kept = true
		}
	}
	return -1
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
