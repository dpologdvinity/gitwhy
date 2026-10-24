// Package hook gives plain `git` commands undo points: a shell function calls Pre before
// each git command and Post after it.
package hook

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/dpologdvinity/gitwhy"
	"github.com/dpologdvinity/gitwhy/internal/danger"
	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/preview"
	"github.com/dpologdvinity/gitwhy/internal/snapshot"
)

// Risky reports whether a git command line (without "git") can lose local work.
// Remote-only dangers don't count: a local snapshot can't undo them. The repository is needed to
// tell "git checkout <file>" (overwrites edits) from "git checkout <branch>".
func Risky(repo *git.Repo, args []string) bool {
	if len(args) == 0 || args[0] == "push" {
		return false
	}
	for _, w := range danger.WarningsIn(repo, args) {
		if !strings.Contains(w, "gitwhy can't undo this") {
			return true
		}
	}
	return false
}

// valueGlobal are git's global options that take their value as the next word. In the
// "--name=value" form the value is attached, so those need no special handling.
var valueGlobal = map[string]bool{
	"-C": true, "-c": true, "--git-dir": true, "--work-tree": true,
	"--namespace": true, "--super-prefix": true, "--config-env": true,
}

// splitGlobalOptions separates git's global options (such as "-C dir", "-c k=v" and
// "--no-pager") from the command that follows them. It returns the folder git would run in once
// every "-C" is applied, and the rest of the command line. ok is false when an option can point
// git at a different repository than the current folder (--git-dir, --work-tree): the hook then
// cannot be sure which repository the command changes and does nothing.
func splitGlobalOptions(cwd string, args []string) (dir string, rest []string, ok bool) {
	dir = cwd
	i := 0
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		opt, _, _ := strings.Cut(args[i], "=")
		switch {
		case opt == "--git-dir" || opt == "--work-tree" || opt == "--bare":
			return "", nil, false
		case valueGlobal[args[i]]:
			if i+1 >= len(args) {
				return "", nil, false
			}
			if args[i] == "-C" {
				if dir = args[i+1]; !filepath.IsAbs(dir) {
					dir = filepath.Join(cwd, dir)
				}
			}
			i += 2
		default:
			i++
		}
	}
	return dir, args[i:], true
}

// expandAlias returns args with a leading git alias replaced by its definition. Only simple
// aliases that expand to a git command are followed; shell aliases (starting with "!") are not,
// since gitwhy cannot know what they run. A definition is split on spaces, so quoted arguments
// inside an alias are not understood. Aliases of aliases are followed a few levels deep.
func expandAlias(repo *git.Repo, args []string) []string {
	known := danger.Commands()
	for range 5 {
		if len(args) == 0 || slices.Contains(known, args[0]) {
			return args
		}
		def, err := repo.Output("config", "--get", "alias."+args[0])
		if err != nil || def == "" || strings.HasPrefix(def, "!") {
			return args
		}
		args = append(strings.Fields(def), args[1:]...)
	}
	return args
}

// Pre takes a snapshot before a risky command and returns its ID, or "" when it took none.
// It never fails: the user's git command must run regardless. Global options in front of the
// command (-C, -c, --no-pager) and simple git aliases are seen through.
func Pre(cwd string, args []string) string {
	if os.Getenv("GITWHY_HOOK_OFF") == "1" {
		return ""
	}
	dir, args, ok := splitGlobalOptions(cwd, args)
	if !ok || len(args) == 0 {
		return ""
	}
	r, err := git.Open(dir)
	if err != nil {
		return ""
	}
	args = expandAlias(r, args)
	if !Risky(r, args) {
		return ""
	}
	// preview runs git in the user's folder and returns paths from the repository root, which is
	// where snapshot.Take reads them.
	here := *r
	here.Dir = dir
	s, err := snapshot.Take(r, "git "+strings.Join(args, " "), preview.Preview(&here, args).Save)
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

// scriptFmt is the shell function that wraps git, for bash and zsh. It finds the command word,
// skipping git's global options, and only starts gitwhy for a command that has a danger rule
// (%[1]s) or one it does not know (%[2]s lists the ones it does and can skip): an unknown word may
// be a git alias for something risky, which gitwhy then looks up. "function git {" (not "git()")
// still parses when the user has aliased git. The variables are initialised so "set -u" is safe.
const scriptFmt = `function git {
  local __gw_id= __gw_rc= __gw_sub= __gw_skip=0 __gw_a=
  for __gw_a in "$@"; do
    if [ "$__gw_skip" = 1 ]; then __gw_skip=0; continue; fi
    case "$__gw_a" in
      -C|-c|--git-dir|--work-tree|--namespace|--super-prefix|--config-env) __gw_skip=1 ;;
      -*) ;;
      *) __gw_sub=$__gw_a; break ;;
    esac
  done
  case "$__gw_sub" in
    %[2]s|'') ;;
    *) __gw_id=$(command gitwhy __pre "$@" 2>/dev/null) ;;
  esac
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
	// push has rules too, but Pre never snapshots it: a local snapshot can't undo a push.
	var risky []string
	for _, c := range danger.Commands() {
		if c != "push" {
			risky = append(risky, c)
		}
	}
	// Every documented command that is not risky can be skipped without asking gitwhy.
	var safe []string
	for _, c := range builtinCommands() {
		if !slices.Contains(risky, c) {
			safe = append(safe, c)
		}
	}
	return fmt.Sprintf(scriptFmt, strings.Join(risky, "|"), strings.Join(safe, "|")), nil
}

// commandName matches a safe shell word for a git command name. Names are put into a shell
// "case" pattern, so anything else is left out.
var commandName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// builtinCommands lists git's commands from the embedded command docs (docs/commands/<name>.md),
// sorted. "git" itself is the page for global options, not a command.
func builtinCommands() []string {
	entries, err := fs.ReadDir(gitwhy.Content, "docs/commands")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		if name != "git" && name != e.Name() && commandName.MatchString(name) {
			names = append(names, name)
		}
	}
	return names
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
