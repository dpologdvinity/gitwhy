// Command gitwhy helps you do anything with git, see what it will do first, and undo it.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpologdvinity/gitwhy"
	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/hook"
	"github.com/dpologdvinity/gitwhy/internal/todo"
	"github.com/dpologdvinity/gitwhy/internal/ui"
)

var version = "dev"

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

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, ui.ErrCancelled) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "gitwhy:", err)
		os.Exit(1)
	}
}

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
	quiet := false
	var rest []string
	for i, a := range args {
		if a == "-q" || a == "--quiet" {
			quiet = true
			continue
		}
		if a == "-h" || a == "--help" || (i == 0 && a == "help") {
			fmt.Print(usage)
			return nil
		}
		if i == 0 && (a == "version" || a == "--version") {
			fmt.Println("gitwhy", version)
			return nil
		}
		rest = append(rest, a)
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
	// "gitwhy undo" is the undo command, but "gitwhy undo last commit" is a search.
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
		rc := filepath.Join(home, "."+shell+"rc")
		if args[0] == "uninstall" {
			changed := false
			for _, sh := range []string{"bash", "zsh"} {
				c, err := hook.Uninstall(filepath.Join(home, "."+sh+"rc"))
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

// checkGit needs git 2.38+ (merge-tree --write-tree, used for previews).
func checkGit() error {
	major, minor, err := git.Version()
	if err != nil {
		return fmt.Errorf("git not found: install git first (%v)", err)
	}
	if major < 2 || (major == 2 && minor < 38) {
		return fmt.Errorf("gitwhy needs git 2.38 or newer; you have %d.%d", major, minor)
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
