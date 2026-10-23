// Package ui is the interactive CLI: search, menus, wizards, builder, preview, run, undo, recover.
package ui

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/dpologdvinity/gitwhy/internal/explain"
	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/kb"
	"github.com/dpologdvinity/gitwhy/internal/recipe"
	"github.com/dpologdvinity/gitwhy/internal/search"
)

// App holds everything the UI needs.
type App struct {
	KB      *kb.KB
	Recipes []*recipe.Recipe
	Index   *search.Index
	Repo    *git.Repo // nil outside a repository
	Self    string    // path of the gitwhy binary, for the rebase sequence editor
	Quiet   bool      // hide the teaching breakdown
}

// New loads the knowledge base and recipes from content and opens the repo around the cwd.
func New(content fs.FS) (*App, error) {
	k, err := kb.Load(content)
	if err != nil {
		return nil, fmt.Errorf("loading docs: %w", err)
	}
	rs, err := recipe.Load(content)
	if err != nil {
		return nil, fmt.Errorf("loading recipes: %w", err)
	}
	a := &App{KB: k, Recipes: rs, Index: search.New(k, rs)}
	if wd, err := os.Getwd(); err == nil {
		a.Repo, _ = git.Open(wd)
	}
	a.Self, _ = os.Executable()
	a.Quiet = os.Getenv("GITWHY_EXPLAIN") == "0"
	return a, nil
}

// ErrCancelled means the user backed out; callers exit quietly.
var ErrCancelled = errors.New("cancelled")

func cancelled(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrCancelled
	}
	return err
}

func (a *App) needRepo() error {
	if a.Repo == nil {
		return errors.New("not inside a git repository (run this from a project folder)")
	}
	return nil
}

func (a *App) recipe(id string) *recipe.Recipe {
	for _, r := range a.Recipes {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// Styles.
var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	cmdStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	tokStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	dimStyle   = lipgloss.NewStyle().Faint(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	badStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
)

func say(format string, args ...any) { lipgloss.Println(fmt.Sprintf(format, args...)) }

func heading(s string) { lipgloss.Println("\n" + titleStyle.Render(s)) }

func warn(s string) { lipgloss.Println(warnStyle.Render("⚠ " + s)) }

// showCommand prints a git command and, unless quiet, what each part means.
func (a *App) showCommand(args []string) {
	lipgloss.Println("  " + cmdStyle.Render("git "+shellJoin(args)))
	if a.Quiet {
		return
	}
	parts := explain.Explain(a.KB, a.Repo, args)
	w := 0
	for _, p := range parts {
		if n := len(p.Token); n > w && n <= 24 {
			w = n
		}
	}
	for _, p := range parts {
		tok := p.Token
		pad := max(w-len(tok), 0)
		lipgloss.Println("    " + tokStyle.Render(tok) + strings.Repeat(" ", pad+2) + dimStyle.Render(brief(p.Meaning)))
	}
}

// brief keeps the first sentence of a description, for the one-line breakdown.
func brief(s string) string {
	if i := strings.Index(s, ". "); i > 0 {
		return s[:i+1]
	}
	return s
}

// shellJoin quotes args that need it, for display and copy-paste.
func shellJoin(args []string) string {
	out := make([]string, len(args))
	for i, s := range args {
		if s == "" || strings.ContainsAny(s, " \t\n'\"\\$`*?![]{}()<>|&;#") || strings.HasPrefix(s, "~") {
			out[i] = "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		} else {
			out[i] = s
		}
	}
	return strings.Join(out, " ")
}

// shellSplit splits a line into words, honoring single and double quotes and backslashes.
func shellSplit(s string) ([]string, error) {
	var out []string
	var cur strings.Builder
	in := false
	var q rune
	esc := false
	for _, r := range s {
		switch {
		case esc:
			cur.WriteRune(r)
			esc = false
			in = true
		case r == '\\' && q != '\'':
			esc = true
		case q != 0:
			if r == q {
				q = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			q = r
			in = true
		case r == ' ' || r == '\t':
			if in {
				out = append(out, cur.String())
				cur.Reset()
				in = false
			}
		default:
			cur.WriteRune(r)
			in = true
		}
	}
	if q != 0 {
		return nil, errors.New("unclosed quote")
	}
	if in {
		out = append(out, cur.String())
	}
	return out, nil
}

func confirm(title string, def bool) (bool, error) {
	v := def
	err := prompt(huh.NewConfirm().Title(title).Affirmative("Yes").Negative("No").Value(&v))
	return v, cancelled(err)
}

// accessible switches every prompt to plain line-based text (screen readers, scripts):
// set ACCESSIBLE=1 or GITWHY_ACCESSIBLE=1.
var accessible = os.Getenv("ACCESSIBLE") != "" || os.Getenv("GITWHY_ACCESSIBLE") != ""

func prompt(f huh.Field) error {
	if accessible {
		return f.RunAccessible(os.Stdout, lineReader{})
	}
	return huh.Run(f)
}

// lineReader reads stdin at most one line per Read, so a prompt's buffered scanner never
// swallows the answers meant for the next prompts.
type lineReader struct{}

func (lineReader) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) {
		m, err := os.Stdin.Read(p[n : n+1])
		n += m
		if err != nil {
			return n, err
		}
		if m == 1 && p[n-1] == '\n' {
			break
		}
	}
	return n, nil
}

func init() {
	// Plain text for screen readers, pipes and logs: no colors baked into prompt labels.
	if fi, err := os.Stdout.Stat(); accessible || os.Getenv("NO_COLOR") != "" || err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		plain := lipgloss.NewStyle()
		titleStyle, cmdStyle, tokStyle, dimStyle, warnStyle, badStyle, okStyle = plain, plain, plain, plain, plain, plain, plain
	}
}
