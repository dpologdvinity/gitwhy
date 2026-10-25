// Command gitwhy helps you do anything with git, see what it will do first, and undo it.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dpologdvinity/gitwhy"
	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/hook"
	"github.com/dpologdvinity/gitwhy/internal/todo"
	"github.com/dpologdvinity/gitwhy/internal/ui"
)

// version is the release version, set at build time with -ldflags "-X main.version=...".
var version = "dev"

// usage is the help text for -h, --help and "gitwhy help".
const usage = `gitwhy: do anything with git, see what it will do first, and undo it.

Usage:
  gitwhy                     full-screen home: changes, commits, branches, next steps
  gitwhy menu                simple menu
  gitwhy <words...>          search, e.g. gitwhy undo last commit
  gitwhy run <git command>   preview, explain, run, with undo (e.g. gitwhy run reset --hard HEAD~1)
  gitwhy explain <command>   explain each part of a git command without running it
  gitwhy build [command]     build any git command from every flag, explained
  gitwhy do [recipe]         run a guided recipe by name (no name: list them)
  gitwhy undo [--list]       undo the last gitwhy action (or pick one)
  gitwhy recover             see what happened recently (plain git too) and go back
  gitwhy resolve             fix merge conflicts
  gitwhy continue | abort    continue or cancel a merge/rebase/cherry-pick in progress
  gitwhy hook install        undo points for plain git commands too (bash, zsh; not fish)
  gitwhy hook uninstall      remove the hook from your shell startup file
  gitwhy hook init <shell>   print the hook (bash or zsh)
  gitwhy version

Flags:
  -q, --quiet   don't explain each part of commands (or set GITWHY_EXPLAIN=0)
`

// main runs gitwhy and turns its error into an exit status: git's own status when git failed
// in the terminal, 1 without a message when the user cancelled, and 1 with a message otherwise.
func main() {
	if err := run(os.Args[1:]); err != nil {
		var gf ui.GitFailed
		if errors.As(err, &gf) { // git already said why
			os.Exit(gf.Code)
		}
		if errors.Is(err, ui.ErrCancelled) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "gitwhy:", err)
		os.Exit(1)
	}
}

// run routes the command line to the matching App method. Words that are not a known command
// are a search query.
func run(args []string) error {
	// git runs us as the rebase sequence editor: gitwhy __todo <plan> <todo-file>.
	if len(args) == 3 && args[0] == "__todo" {
		return todo.Apply(args[1], args[2])
	}
	// The shell hook calls these around every git command, so they skip checkGit and docs loading.
	if len(args) >= 1 && args[0] == "__pre" {
		wd, _ := os.Getwd()
		fmt.Print(hook.Pre(wd, args[1:]))
		return nil
	}
	if len(args) == 2 && args[0] == "__post" {
		wd, _ := os.Getwd()
		hook.Post(wd, args[1])
		return nil
	}
	if len(args) >= 2 && args[0] == "hook" {
		return hookCmd(args[1:])
	}
	// gitwhy's own flags come before the first word; after it they belong to that command
	// (e.g. "gitwhy run git grep -q foo").
	quiet := false
	rest := args
	for len(rest) > 0 && (rest[0] == "-q" || rest[0] == "--quiet") {
		quiet = true
		rest = rest[1:]
	}
	if len(rest) > 0 {
		switch rest[0] {
		case "-h", "--help", "help":
			fmt.Print(usage)
			return nil
		case "version", "--version":
			fmt.Println("gitwhy", version)
			return nil
		}
	}
	if err := checkGit(); err != nil {
		return err
	}
	app, err := ui.New(gitwhy.Content)
	if err != nil {
		return err
	}
	app.Quiet = app.Quiet || quiet
	if len(rest) == 0 {
		if interactive() {
			return app.TUI()
		}
		return app.Home()
	}
	// "gitwhy undo" is the undo command, but "gitwhy undo last commit" is a search. The same goes
	// for recover, resolve, continue and abort: with more words they fall through to Search.
	words := len(rest)
	if words > 1 && rest[0] == "undo" && (rest[1] == "--list" || rest[1] == "-l") {
		words = 1
	}
	switch rest[0] {
	case "run":
		return app.Run(rest[1:])
	case "menu":
		return app.Home()
	case "explain":
		return app.Explain(rest[1:])
	case "do":
		if words == 2 {
			return app.RunRecipeID(rest[1])
		}
		return app.ListRecipes()
	case "build":
		cmd := ""
		if len(rest) > 1 {
			cmd = strings.TrimPrefix(rest[1], "git-")
		}
		return app.Build(cmd, "")
	case "undo":
		if words > 1 {
			break
		}
		return app.Undo(len(rest) > 1 && (rest[1] == "--list" || rest[1] == "-l"))
	case "recover":
		if words > 1 {
			break
		}
		return app.Recover()
	case "resolve":
		if words > 1 {
			break
		}
		return app.RunRecipeID("resolve-conflicts")
	case "continue":
		if words > 1 {
			break
		}
		return app.RunRecipeID("continue-operation")
	case "abort":
		if words > 1 {
			break
		}
		return app.RunRecipeID("abort-operation")
	}
	return app.Search(strings.Join(rest, " "))
}

// hookCmd handles "gitwhy hook init|install|uninstall". install edits the startup file of the
// shell in $SHELL; uninstall removes the hook from both the bash and zsh files, so it works
// even after the user changed shells.
func hookCmd(args []string) error {
	switch args[0] {
	case "init":
		if len(args) != 2 {
			return errors.New("usage: gitwhy hook init bash|zsh")
		}
		s, err := hook.Script(args[1])
		if err == nil {
			fmt.Print(s)
		}
		return err
	case "install", "uninstall":
		shell := filepath.Base(os.Getenv("SHELL"))
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		rc := hook.RCPath(shell, home)
		if args[0] == "uninstall" {
			changed := false
			paths := []string{hook.RCPath("bash", home), hook.RCPath("zsh", home)}
			if runtime.GOOS == "darwin" {
				paths = append(paths, filepath.Join(home, ".bash_profile"))
			}
			for _, rc := range paths {
				c, err := hook.Uninstall(rc)
				if err != nil {
					return err
				}
				changed = changed || c
			}
			if changed {
				fmt.Println("Removed. Open a new terminal for it to take effect.")
			} else {
				fmt.Println("The hook wasn't installed.")
			}
			return nil
		}
		if _, err := hook.Script(shell); err != nil {
			return fmt.Errorf("your shell is %s: %w", shell, err)
		}
		if ok, err := ui.Confirm("Add the gitwhy hook to "+rc+"?", true); err != nil || !ok {
			return err
		}
		changed, err := hook.Install(rc, shell)
		if err != nil {
			return err
		}
		if !changed {
			fmt.Println("Already in " + rc + ".")
			return nil
		}
		fmt.Println("Open a new terminal (or run: source " + rc + ")")
		return nil
	}
	return fmt.Errorf("unknown hook command %q: use init, install or uninstall", args[0])
}

// checkGit returns an error unless git 2.40 or newer is installed: previews use
// "merge-tree --write-tree --merge-base".
func checkGit() error {
	major, minor, err := git.Version()
	if err != nil {
		return fmt.Errorf("git not found: install git first (%v)", err)
	}
	if major < 2 || (major == 2 && minor < 40) {
		return fmt.Errorf("gitwhy needs git 2.40 or newer; you have %d.%d", major, minor)
	}
	return nil
}

// interactive reports whether the full-screen UI can run (a terminal, not accessible mode).
func interactive() bool {
	if os.Getenv("ACCESSIBLE") != "" || os.Getenv("GITWHY_ACCESSIBLE") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
