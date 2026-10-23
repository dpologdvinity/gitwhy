// Package hook gives plain `git` commands undo points: a shell function calls Pre before
// each git command and Post after it.
package hook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/danger"
	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/preview"
	"github.com/dpologdvinity/gitwhy/internal/snapshot"
)

// Risky reports whether a git command line (without "git") can lose local work.
// Remote-only dangers don't count: a local snapshot can't undo them.
func Risky(args []string) bool {
	if len(args) == 0 || args[0] == "push" {
		return false
	}
	for _, w := range danger.Warnings(args) {
		if !strings.Contains(w, "gitwhy can't undo this") {
			return true
		}
	}
	return false
}

// Pre takes a snapshot before a risky command and returns its ID, or "" when it took none.
// It never fails: the user's git command must run regardless.
func Pre(cwd string, args []string) string {
	// Global options (-C, -c, --git-dir) would change which repo or command runs; skip them.
	if len(args) == 0 || strings.HasPrefix(args[0], "-") || os.Getenv("GITWHY_HOOK_OFF") == "1" || !Risky(args) {
		return ""
	}
	r, err := git.Open(cwd)
	if err != nil {
		return ""
	}
	// Paths in args are relative to the user's folder; Take wants them relative to the repo top.
	here := *r
	here.Dir = cwd
	save := preview.Preview(&here, args).Save
	if prefix, err := here.Output("rev-parse", "--show-prefix"); err == nil && prefix != "" {
		for i, p := range save {
			save[i] = prefix + p
		}
	}
	s, err := snapshot.Take(r, "git "+strings.Join(args, " "), save)
	if err != nil {
		return ""
	}
	return s.ID
}

// Post records the state after the command ran. Errors are ignored.
func Post(cwd, id string) {
	r, err := git.Open(cwd)
	if err != nil {
		return
	}
	if s, err := snapshot.Get(r, id); err == nil {
		_ = snapshot.Finish(r, s)
	}
}

const script = `git() {
  local __gw_id __gw_rc
  __gw_id=$(command gitwhy __pre "$@" 2>/dev/null)
  command git "$@"
  __gw_rc=$?
  if [ -n "$__gw_id" ]; then command gitwhy __post "$__gw_id" >/dev/null 2>&1; fi
  return $__gw_rc
}
`

const (
	begin = "# >>> gitwhy hook >>>"
	end   = "# <<< gitwhy hook <<<"
)

// Script returns the shell function that wraps git.
func Script(shell string) (string, error) {
	if shell != "bash" && shell != "zsh" {
		return "", fmt.Errorf("%q: only bash and zsh are supported", shell)
	}
	return script, nil
}

// Install adds the hook to a shell rc file. It does nothing if it's already there.
func Install(rcPath, shell string) (changed bool, err error) {
	if _, err := Script(shell); err != nil {
		return false, err
	}
	data, err := os.ReadFile(rcPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	s := string(data)
	if strings.Contains(s, begin) {
		return false, nil
	}
	if s != "" && !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	s += begin + "\neval \"$(gitwhy hook init " + shell + ")\"\n" + end + "\n"
	return true, os.WriteFile(rcPath, []byte(s), 0o644)
}

// Uninstall removes the hook block from a shell rc file, leaving the rest as it was.
func Uninstall(rcPath string) (changed bool, err error) {
	data, err := os.ReadFile(rcPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	s := string(data)
	i := strings.Index(s, begin)
	if i < 0 {
		return false, nil
	}
	j := strings.Index(s[i:], end)
	if j < 0 {
		return false, fmt.Errorf("%s has %q but no end marker %q: remove those lines by hand", rcPath, begin, end)
	}
	rest := strings.TrimPrefix(s[i+j+len(end):], "\n")
	return true, os.WriteFile(rcPath, []byte(s[:i]+rest), 0o644)
}

// RCPath is the startup file the hook goes in: ~/.bashrc, or $ZDOTDIR/.zshrc (default ~/.zshrc).
func RCPath(shell, home string) string {
	if z := os.Getenv("ZDOTDIR"); shell == "zsh" && z != "" {
		home = z
	}
	return filepath.Join(home, "."+shell+"rc")
}
