// Package ui is the interactive CLI: search, menus, wizards, builder, preview, run, undo, recover.
package ui

import (
	"errors"
	"fmt"
	"io"
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
	KB      *kb.KB           // the embedded command reference
	Recipes []*recipe.Recipe // every embedded recipe
	Index   *search.Index    // search over recipes, commands and flags, built once by New
	Repo    *git.Repo        // nil outside a repository
	Self    string           // path of the gitwhy binary, for the rebase sequence editor
	Quiet   bool             // hide the teaching breakdown
	Cwd     string           // where gitwhy was started; literal commands run here
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
		a.Cwd = wd
		a.Repo, _ = git.Open(wd) // outside a repository Repo stays nil; needRepo reports it
	}
	a.Self, _ = os.Executable()
	a.Quiet = os.Getenv("GITWHY_EXPLAIN") == "0"
	return a, nil
}

// ErrCancelled means the user backed out; callers exit quietly.
var ErrCancelled = errors.New("cancelled")

// GitFailed means git ran attached to the terminal and failed, and has already printed why.
// The program exits with Code (git's exit status) and prints nothing more.
type GitFailed struct{ Code int }

// Error describes the exit status; main does not print it.
func (e GitFailed) Error() string { return fmt.Sprintf("git exited with status %d", e.Code) }

// passthrough turns the error of a git command run straight through (stderr shown to the user)
// into GitFailed, so its message is not printed twice and its exit status is kept.
func passthrough(err error) error {
	if code := git.ExitCode(err); code > 0 {
		return GitFailed{code}
	}
	return err
}

// cancelled turns huh's "user aborted" (Ctrl+C, Esc, end of input) into ErrCancelled.
func cancelled(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrCancelled
	}
	return err
}

// needRepo returns a plain error when gitwhy was not started inside a git repository.
func (a *App) needRepo() error {
	if a.Repo == nil {
		return errors.New("not inside a git repository (run this from a project folder)")
	}
	return nil
}

// recipe returns the recipe with the given ID, or nil.
func (a *App) recipe(id string) *recipe.Recipe {
	for _, r := range a.Recipes {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// Styles for titles, commands, tokens, dim notes, warnings, failures and success. init
// replaces them all with a plain style when output is not a color terminal.
var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	cmdStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	tokStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	dimStyle   = lipgloss.NewStyle().Faint(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	badStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
)

// say prints one formatted line to stdout.
func say(format string, args ...any) { lipgloss.Println(fmt.Sprintf(format, args...)) }

// heading prints a bold section title after a blank line.
func heading(s string) { lipgloss.Println("\n" + titleStyle.Render(s)) }

// warn prints a warning line with a warning sign.
func warn(s string) { lipgloss.Println(warnStyle.Render("⚠ " + s)) }

// showCommand prints a git command and, unless quiet, what each part means.
func (a *App) showCommand(args []string) {
	lipgloss.Println("  " + cmdStyle.Render("git "+shellJoin(args)))
	if a.Quiet {
		return
	}
	parts := explain.Explain(a.KB, a.Repo, args)
	// Align meanings after the longest token, but let tokens over 24 characters overflow.
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

// shellJoin quotes args that need it, for display and copy-paste. It uses POSIX single
// quotes, so the output is meant for sh-like shells.
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
	in := false // inside a word; true even for an empty quoted word such as ''
	var q rune  // the open quote character, or 0
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

// Confirm asks a yes/no question; def is the answer when the user just presses Enter.
func Confirm(title string, def bool) (bool, error) {
	v := def
	err := prompt(huh.NewConfirm().Title(title).Affirmative("Yes").Negative("No").Value(&v))
	return v, cancelled(err)
}

// accessible switches every prompt to plain line-based text (screen readers, scripts):
// set ACCESSIBLE=1 or GITWHY_ACCESSIBLE=1.
var accessible = os.Getenv("ACCESSIBLE") != "" || os.Getenv("GITWHY_ACCESSIBLE") != ""

// prompt runs one form field: as a full-screen huh form, or as plain text prompts when
// accessible is set. A user abort or end of input returns huh.ErrUserAborted.
func prompt(f huh.Field) (err error) {
	if accessible {
		defer func() {
			if v := recover(); v != nil {
				if v != errEOF {
					panic(v)
				}
				err = huh.ErrUserAborted
			}
		}()
		return f.RunAccessible(os.Stdout, lineReader{})
	}
	return huh.Run(f)
}

// errEOF is what lineReader panics with when stdin is used up. huh's plain-text prompts treat
// end of input as an empty answer (take the default) and re-ask forever when the default is not
// valid, so prompt recovers it and cancels instead.
var errEOF = errors.New("end of input")

// lineReader reads stdin at most one line per Read, so a prompt's buffered scanner never
// swallows the answers meant for the next prompts. At end of input it panics with errEOF.
type lineReader struct{}

// Read reads one byte at a time up to and including the next newline.
func (lineReader) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) {
		m, err := os.Stdin.Read(p[n : n+1])
		n += m
		if err == io.EOF && n == 0 {
			panic(errEOF)
		}
		if err != nil {
			return n, err
		}
		if m == 1 && p[n-1] == '\n' {
			break
		}
	}
	return n, nil
}

// init turns off all styles in accessible mode, with NO_COLOR, or when stdout is not a terminal.
func init() {
	// Plain text for screen readers, pipes and logs: no colors baked into prompt labels.
	if fi, err := os.Stdout.Stat(); accessible || os.Getenv("NO_COLOR") != "" || err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		plain := lipgloss.NewStyle()
		titleStyle, cmdStyle, tokStyle, dimStyle, warnStyle, badStyle, okStyle = plain, plain, plain, plain, plain, plain, plain
	}
}
