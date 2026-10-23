// Package tui is gitwhy's full-screen home: where you are, what changed, and what you can do next.
// It only reads the repo. Every change goes through a recipe (preview, explain, undo): the TUI
// returns the chosen Action, the caller runs it, then reopens the TUI.
package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/graph"
	"github.com/dpologdvinity/gitwhy/internal/recipe"
)

// Action is what the user picked.
type Action struct {
	Kind   string // "recipe", "search", "undo", "recover", "build", "quit"
	Recipe string
	Preset recipe.Answers
}

func run(id string, preset recipe.Answers) Action {
	return Action{Kind: "recipe", Recipe: id, Preset: preset}
}

const (
	tabChanges = iota
	tabCommits
	tabBranches
)

var tabNames = []string{"Changes", "Commits", "Branches"}

type file struct {
	x, y byte // porcelain status letters: index, work tree
	path string
}

func (f file) conflict() bool {
	return f.x == 'U' || f.y == 'U' || (f.x == 'A' && f.y == 'A') || (f.x == 'D' && f.y == 'D')
}
func (f file) untracked() bool { return f.x == '?' }
func (f file) staged() bool    { return f.x != ' ' && f.x != '?' && !f.conflict() }
func (f file) unstaged() bool  { return f.y != ' ' && f.y != '?' && !f.conflict() }

func (f file) label() string {
	switch {
	case f.conflict():
		return "conflict"
	case f.untracked():
		return "new"
	case f.staged() && f.unstaged():
		return "staged+edited"
	case f.staged():
		return "staged"
	}
	return "changed"
}

type branch struct{ name, track, when, subject string }

type model struct {
	repo    *git.Repo
	w, h    int
	tab     int
	cursor  [3]int
	files   []file
	rows    []graph.Row // Commits tab; the cursor counts commit rows only
	all     bool        // Commits tab shows every branch
	brs     []branch
	head    string // branch name or "HEAD (detached)"
	ab      string // "↑2 ↓1" vs upstream
	up      string
	op      string
	help    bool
	msg     string // one-line result of the last quick action
	detail  string
	action  Action
	hunks   *hunkView // non-nil while zoomed into a file's hunks
}

// Run shows the TUI and returns what the user chose.
func Run(repo *git.Repo) (Action, error) {
	m := &model{repo: repo, action: Action{Kind: "quit"}}
	m.load()
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return Action{}, err
	}
	return final.(*model).action, nil
}

func (m *model) load() {
	r := m.repo
	m.op = recipe.InProgress(r)
	out, _ := r.Output("status", "--porcelain=v2", "--branch", "-z", "--untracked-files=all")
	m.files, m.head, m.up, m.ab = nil, "", "", ""
	parts := strings.Split(out, "\x00")
	for i := 0; i < len(parts); i++ {
		e := parts[i]
		switch {
		case strings.HasPrefix(e, "# branch.head "):
			m.head = strings.TrimPrefix(e, "# branch.head ")
		case strings.HasPrefix(e, "# branch.upstream "):
			m.up = strings.TrimPrefix(e, "# branch.upstream ")
		case strings.HasPrefix(e, "# branch.ab "):
			var a, b int
			if _, err := fmt.Sscanf(strings.TrimPrefix(e, "# branch.ab "), "+%d -%d", &a, &b); err == nil {
				m.ab = fmt.Sprintf("↑%d ↓%d", a, b)
			}
		case strings.HasPrefix(e, "1 ") || strings.HasPrefix(e, "u "):
			f := strings.SplitN(e, " ", 9)
			if e[0] == 'u' {
				f = strings.SplitN(e, " ", 11)
			}
			m.files = append(m.files, file{x: dot(f[1][0]), y: dot(f[1][1]), path: f[len(f)-1]})
		case strings.HasPrefix(e, "2 "):
			f := strings.SplitN(e, " ", 10)
			m.files = append(m.files, file{x: dot(f[1][0]), y: dot(f[1][1]), path: f[9]})
			i++ // original path follows
		case strings.HasPrefix(e, "? "):
			m.files = append(m.files, file{x: '?', y: '?', path: e[2:]})
		}
	}
	if m.head == "(detached)" {
		m.head = "HEAD (detached)"
	}
	m.rows, _ = graph.Load(r, m.all, 200)
	m.brs = nil
	if refs, err := r.Output("for-each-ref", "--sort=-committerdate", "--format=%(refname)%00%(upstream:track)%00%(committerdate:relative)%00%(subject)", "refs/heads", "refs/remotes"); err == nil && refs != "" {
		for _, l := range strings.Split(refs, "\n") {
			f := strings.Split(l, "\x00")
			if len(f) != 4 || strings.HasSuffix(f[0], "/HEAD") {
				continue
			}
			name := strings.TrimPrefix(strings.TrimPrefix(f[0], "refs/heads/"), "refs/remotes/")
			m.brs = append(m.brs, branch{name, f[1], f[2], f[3]})
		}
	}
	for t := range m.cursor {
		if n := m.count(t); m.cursor[t] >= n {
			m.cursor[t] = max(n-1, 0)
		}
	}
	m.loadDetail()
}

func (m *model) count(tab int) int {
	switch tab {
	case tabChanges:
		return len(m.files)
	case tabCommits:
		n := 0
		for _, row := range m.rows {
			if row.Commit != nil {
				n++
			}
		}
		return n
	}
	return len(m.brs)
}

// commitAt returns the i-th commit in the graph, skipping connector rows.
func (m *model) commitAt(i int) *graph.Commit {
	for _, row := range m.rows {
		if row.Commit != nil {
			if i == 0 {
				return row.Commit
			}
			i--
		}
	}
	return nil
}

func (m *model) loadDetail() {
	r := m.repo
	m.detail = ""
	i := m.cursor[m.tab]
	switch m.tab {
	case tabChanges:
		if i >= len(m.files) {
			return
		}
		f := m.files[i]
		switch {
		case f.untracked():
			m.detail, _ = r.Output("diff", "--no-index", "--", "/dev/null", f.path)
		case f.staged() && !f.unstaged():
			m.detail, _ = r.Output("diff", "--cached", "--", f.path)
		default:
			m.detail, _ = r.Output("diff", "--", f.path)
		}
	case tabCommits:
		if c := m.commitAt(i); c != nil {
			m.detail, _ = r.Output("show", "--stat", "--format=%h %s%n%an, %cr%n%b", c.ID, "--")
		}
	case tabBranches:
		if i < len(m.brs) {
			b := m.brs[i].name
			in, _ := r.Output("log", "--oneline", "-n", "15", "HEAD.."+b, "--")
			out, _ := r.Output("log", "--oneline", "-n", "15", b+"..HEAD", "--")
			m.detail = "Only on " + b + ":\n" + orNone(in) + "\n\nOnly on " + m.head + ":\n" + orNone(out)
		}
	}
}

// dot maps porcelain v2's "." (unchanged) to a space, like porcelain v1.
func dot(b byte) byte {
	if b == '.' {
		return ' '
	}
	return b
}

func orNone(s string) string {
	if s == "" {
		return "  (nothing)"
	}
	return "  " + strings.ReplaceAll(s, "\n", "\n  ")
}

// suggest returns next steps for the current state, most urgent first.
func (m *model) suggest() []string {
	var s []string
	conflicts, staged, dirty := 0, 0, 0
	for _, f := range m.files {
		switch {
		case f.conflict():
			conflicts++
		case f.staged():
			staged++
		}
		if !f.conflict() {
			dirty++
		}
	}
	switch {
	case conflicts > 0:
		s = append(s, fmt.Sprintf("%d conflict(s): Enter on a file to resolve", conflicts))
	case m.op != "":
		s = append(s, "C continue the "+m.op, "A abort it")
	}
	if staged > 0 && m.op == "" {
		s = append(s, fmt.Sprintf("c commit %d staged file(s)", staged))
	} else if dirty > 0 && m.op == "" {
		s = append(s, "space stage a file", "a stage all", "s stash")
	}
	if strings.HasPrefix(m.ab, "↑") && !strings.HasPrefix(m.ab, "↑0") {
		s = append(s, "P push")
	}
	if strings.Contains(m.ab, "↓") && !strings.HasSuffix(m.ab, "↓0") {
		s = append(s, "p pull")
	}
	if m.up == "" && m.head != "" && !strings.HasPrefix(m.head, "HEAD") && m.count(tabCommits) > 0 {
		s = append(s, "P publish this branch")
	}
	return s
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil
	case tea.KeyPressMsg:
		return m.key(msg.String())
	}
	return m, nil
}

func (m *model) pick(a Action) (tea.Model, tea.Cmd) {
	m.action = a
	return m, tea.Quit
}

func (m *model) key(k string) (tea.Model, tea.Cmd) {
	if m.hunks != nil {
		return m.hunkKey(k)
	}
	m.msg = ""
	i := m.cursor[m.tab]
	n := m.count(m.tab)
	switch k {
	case "q", "ctrl+c", "esc":
		return m.pick(Action{Kind: "quit"})
	case "?":
		m.help = !m.help
		return m, nil
	case "tab", "right", "l":
		m.tab = (m.tab + 1) % 3
	case "shift+tab", "left", "h":
		m.tab = (m.tab + 2) % 3
	case "1", "2", "3":
		m.tab = int(k[0] - '1')
	case "down", "j":
		if i+1 < n {
			m.cursor[m.tab]++
		}
	case "up", "k":
		if i > 0 {
			m.cursor[m.tab]--
		}
	case "g", "home":
		m.cursor[m.tab] = 0
	case "G", "end":
		m.cursor[m.tab] = max(n-1, 0)
	case "ctrl+r":
		m.load()
		return m, nil
	case "/":
		return m.pick(Action{Kind: "search"})
	case "u":
		return m.pick(Action{Kind: "undo"})
	case "R":
		return m.pick(Action{Kind: "recover"})
	case "B":
		return m.pick(Action{Kind: "build"})
	case "P":
		return m.pick(run("push-branch", nil))
	case "p":
		if m.tab != tabCommits {
			return m.pick(run("pull-latest", nil))
		}
		if a, ok := m.tabKey(k, i); ok { // p copies a commit on the Commits tab
			return m.pick(a)
		}
	case "C":
		if m.op != "" {
			return m.pick(run("continue-operation", nil))
		}
	case "A":
		if m.op != "" {
			return m.pick(run("abort-operation", nil))
		}
	default:
		if a, ok := m.tabKey(k, i); ok {
			return m.pick(a)
		}
	}
	m.loadDetail()
	return m, nil
}

// tabKey handles keys that act on the selected item.
func (m *model) tabKey(k string, i int) (Action, bool) {
	switch m.tab {
	case tabChanges:
		if k == "a" {
			m.quick("add", "-A")
			return Action{}, false
		}
		if k == "s" {
			return run("stash-save", nil), true
		}
		if k == "c" {
			var staged []string
			for _, f := range m.files {
				if f.staged() {
					staged = append(staged, f.path)
				}
			}
			if len(staged) > 0 {
				return run("commit-changes", recipe.Answers{"files": staged}), true
			}
			return run("commit-changes", nil), true
		}
		if i >= len(m.files) {
			return Action{}, false
		}
		f := m.files[i]
		switch k {
		case "space":
			if f.conflict() {
				return run("resolve-conflicts", recipe.Answers{"files": []string{f.path}}), true
			}
			if f.staged() {
				m.quick("restore", "--staged", "--", f.path)
			} else {
				m.quick("add", "-A", "--", f.path)
			}
		case "enter":
			if f.conflict() {
				return run("resolve-conflicts", recipe.Answers{"files": []string{f.path}}), true
			}
			if f.untracked() {
				m.msg = "New file: press space to stage it whole"
			} else if err := m.openHunks(f.path, !f.unstaged()); err != nil {
				m.msg = err.Error()
			}
		case "d":
			if f.untracked() {
				return run("delete-untracked", nil), true
			}
			return run("discard-file-changes", recipe.Answers{"files": []string{f.path}}), true
		case "H":
			return run("file-history", recipe.Answers{"file": f.path}), true
		}
	case tabCommits:
		if k == "a" {
			m.all = !m.all
			m.load()
			return Action{}, false
		}
		gc := m.commitAt(i)
		if gc == nil {
			return Action{}, false
		}
		c := recipe.Answers{"commit": gc.ID}
		if len(k) == 1 && strings.Contains("rdfsxmo", k) && !m.repo.OK("merge-base", "--is-ancestor", gc.ID, "HEAD") {
			m.msg = "That commit isn't on your branch. Switch to its branch first, or p to copy it here."
			return Action{}, false
		}
		switch k {
		case "p":
			return run("copy-commit", recipe.Answers{"from": gc.ID, "commit": gc.ID}), true
		case "r":
			return run("reword-commit", c), true
		case "d":
			return run("drop-commit", c), true
		case "f":
			return run("fix-past-commit", c), true
		case "s":
			return run("split-commit", c), true
		case "x":
			return run("revert-commit", c), true
		case "e":
			return run("edit-recent-commits", nil), true
		case "b":
			return run("new-branch", recipe.Answers{"base": gc.ID}), true
		case "o":
			return run("reset-branch-to", recipe.Answers{"target": gc.ID}), true
		case "t":
			return run("create-tag", c), true
		case "m":
			return run("move-commits-new-branch", c), true
		}
	case tabBranches:
		if i >= len(m.brs) {
			return Action{}, false
		}
		b := m.brs[i].name
		switch k {
		case "enter":
			return run("switch-branch", recipe.Answers{"target": b}), true
		case "m":
			return run("merge-branch", recipe.Answers{"from": b}), true
		case "b":
			return run("update-from-main", recipe.Answers{"base": b}), true
		case "D":
			return run("delete-branch", recipe.Answers{"target": b}), true
		case "n":
			return run("new-branch", recipe.Answers{"base": b}), true
		case "c":
			return run("compare-branches", recipe.Answers{"a": m.head, "b": b}), true
		}
	}
	return Action{}, false
}

// quick runs a safe, always-reversible command (stage/unstage) without leaving the TUI.
func (m *model) quick(args ...string) {
	if _, err := m.repo.Output(args...); err != nil {
		m.msg = "✗ " + strings.TrimSpace(err.Error())
	} else {
		m.msg = "✓ git " + strings.Join(args, " ")
	}
	m.load()
}

var (
	bold    = lipgloss.NewStyle().Bold(true)
	faint   = lipgloss.NewStyle().Faint(true)
	sel     = lipgloss.NewStyle().Reverse(true)
	tabOn   = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("6"))
	banner  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("3")).Padding(0, 1)
	addLine = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	delLine = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	hunkCol = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	tagCol  = map[string]lipgloss.Style{
		"conflict":      lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		"staged":        lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		"staged+edited": lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		"changed":       lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		"new":           lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
	}
)

var keys = [][]string{
	{"space stage/unstage", "a stage all", "c commit", "d discard", "s stash", "H history", "enter hunks / resolve conflict"},
	{"r reword", "d drop", "f add changes to it", "s split", "x revert", "e reorder/squash", "m move to new branch", "b branch from", "o reset branch here", "t tag", "p copy here", "a all branches"},
	{"enter switch", "m merge into current", "b rebase onto", "n new branch from", "c compare", "D delete"},
}

func (m *model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = "gitwhy"
	return v
}

func (m *model) render() string {
	w, h := max(m.w, 60), max(m.h, 15)
	var top strings.Builder
	head := m.head
	if head == "" {
		head = "(no branch)"
	}
	top.WriteString(bold.Render("gitwhy") + "  on " + bold.Render(head))
	if m.up != "" {
		top.WriteString(faint.Render("  tracking " + m.up + " " + m.ab))
	} else if !strings.HasPrefix(head, "HEAD") {
		top.WriteString(faint.Render("  not on any remote yet"))
	}
	if m.op != "" {
		top.WriteString("  " + banner.Render(strings.ToUpper(m.op)+" IN PROGRESS"))
	}
	var tabs []string
	for t, name := range tabNames {
		label := fmt.Sprintf("%d %s (%d)", t+1, name, m.count(t))
		if t == m.tab {
			tabs = append(tabs, tabOn.Render(label))
		} else {
			tabs = append(tabs, faint.Render(label))
		}
	}
	title := strings.Join(tabs, "   ")
	if m.hunks != nil {
		title = m.hunkTitle()
	}
	lines := []string{top.String(), title, ""}

	sugg := m.suggest()
	footer := []string{""}
	if m.msg != "" {
		footer = append(footer, m.msg)
	}
	if len(sugg) > 0 {
		footer = append(footer, bold.Render("Next: ")+strings.Join(sugg, " · "))
	}
	if m.hunks != nil {
		footer = append(footer, faint.Render("space "+m.hunkKeys()))
	} else if m.help {
		footer = append(footer, faint.Render(strings.Join(keys[m.tab], " · ")))
		footer = append(footer, faint.Render("tab switch · / search · u undo · R what happened? · B any command · P push · p pull · ctrl+r refresh · q quit"))
	} else {
		footer = append(footer, faint.Render("? keys · / search · u undo · q quit"))
	}

	bodyH := max(h-len(lines)-len(footer), 3)
	leftW := min(max(w*2/5, 30), 60)
	rightW := w - leftW - 3
	left, right := m.list(leftW, bodyH), colorDiff(m.detail, rightW, bodyH)
	if m.hunks != nil {
		left, right = m.hunkList(leftW, bodyH), m.hunkDetail(rightW, bodyH)
	}
	for r := 0; r < bodyH; r++ {
		l, rr := "", ""
		if r < len(left) {
			l = left[r]
		}
		if r < len(right) {
			rr = right[r]
		}
		lines = append(lines, pad(l, leftW)+faint.Render(" │ ")+rr)
	}
	lines = append(lines, footer...)
	return strings.Join(lines, "\n")
}

func (m *model) list(w, h int) []string {
	var rows []string
	switch m.tab {
	case tabChanges:
		if len(m.files) == 0 {
			return []string{faint.Render("Nothing changed. Working tree is clean.")}
		}
		for _, f := range m.files {
			tag := f.label()
			rows = append(rows, tagCol[tag].Render(fmt.Sprintf("%-13s", tag))+" "+f.path)
		}
	case tabCommits:
		if len(m.rows) == 0 {
			return []string{faint.Render("No commits yet.")}
		}
		cur, k := -1, 0
		for i, row := range m.rows {
			line := graphArt(row.Graph)
			if c := row.Commit; c != nil {
				if k == m.cursor[tabCommits] {
					cur = i
				}
				k++
				refs := ""
				if c.Refs != "" {
					refs = " " + tabOn.UnsetUnderline().Render("("+c.Refs+")")
				}
				line += " " + faint.Render(c.Short) + " " + c.Subject + refs
			}
			rows = append(rows, line)
		}
		return scroll(rows, cur, w, h)
	case tabBranches:
		for _, b := range m.brs {
			mark := "  "
			if b.name == m.head {
				mark = "* "
			}
			rows = append(rows, mark+b.name+" "+faint.Render(b.track+" "+b.when))
		}
	}
	return scroll(rows, m.cursor[m.tab], w, h)
}

// scroll shows the h rows around cur, truncated to w, with cur highlighted.
func scroll(rows []string, cur, w, h int) []string {
	start := 0
	if cur >= h {
		start = cur - h + 1
	}
	var out []string
	for i := start; i < len(rows) && len(out) < h; i++ {
		row := trunc(rows[i], w)
		if i == cur {
			row = sel.Render(pad(ansi.Strip(row), w)) // strip inner colors so the highlight is unbroken
		}
		out = append(out, row)
	}
	return out
}

func colorDiff(text string, w, h int) []string {
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if len(out) == h {
			break
		}
		l = trunc(strings.ReplaceAll(l, "\t", "    "), w)
		switch {
		case strings.HasPrefix(l, "+++"), strings.HasPrefix(l, "---"), strings.HasPrefix(l, "diff "), strings.HasPrefix(l, "index "):
			l = faint.Render(l)
		case strings.HasPrefix(l, "+"):
			l = addLine.Render(l)
		case strings.HasPrefix(l, "-"):
			l = delLine.Render(l)
		case strings.HasPrefix(l, "@@"):
			l = hunkCol.Render(l)
		}
		out = append(out, l)
	}
	return out
}

// graphArt colors each column of git's graph drawing so lanes are easy to follow.
func graphArt(g string) string {
	var b strings.Builder
	for col, r := range g {
		if r == ' ' {
			b.WriteRune(r)
			continue
		}
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprint(2 + col%5))).Render(string(r)))
	}
	return b.String()
}

func trunc(s string, w int) string { return ansi.Truncate(s, w, "…") }

func pad(s string, w int) string {
	if n := lipgloss.Width(s); n < w {
		return s + strings.Repeat(" ", w-n)
	}
	return s
}
