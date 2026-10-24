// Package git runs the real git CLI. It is the only package that starts git processes.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Runner runs git in Dir with extra environment variables.
type Runner struct {
	Dir string   // folder git runs in
	Env []string // extra "KEY=value" variables, added after the defaults so they win
}

// Error is a failed git command with its stderr.
type Error struct {
	Args   []string // the arguments after "git"
	Stderr string   // what git printed to stderr
	Err    error    // the underlying error, usually an *exec.ExitError
}

// Error returns "git <args>: <message>", where the message is stderr if git printed any.
func (e *Error) Error() string {
	msg := strings.TrimSpace(e.Stderr)
	if msg == "" {
		msg = e.Err.Error()
	}
	return fmt.Sprintf("git %s: %s", strings.Join(e.Args, " "), msg)
}

// Unwrap returns the underlying error, so errors.As can find an *exec.ExitError.
func (e *Error) Unwrap() error { return e.Err }

// cmd builds the git command for args, with the environment every git process gets.
func (r Runner) cmd(args ...string) *exec.Cmd {
	c := exec.Command("git", args...)
	c.Dir = r.Dir
	// Stable, parseable output regardless of the user's locale. File names are literal pathspecs,
	// so "[ab].txt" never matches a.txt; git refuses literal mixed with the other global settings.
	// git stash runs "git clean ... :/" itself, which literal pathspecs break, so stash keeps
	// git's normal pathspecs.
	literal := "1"
	if subcommand(args) == "stash" {
		literal = "0"
	}
	c.Env = append(append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0", "GIT_LITERAL_PATHSPECS="+literal,
		"GIT_GLOB_PATHSPECS=0", "GIT_NOGLOB_PATHSPECS=0", "GIT_ICASE_PATHSPECS=0"), r.Env...)
	return c
}

// subcommand returns the git subcommand in args, skipping global options such as "-c k=v".
func subcommand(args []string) string {
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "-c" || a == "-C":
			i++ // the option's value
		case !strings.HasPrefix(a, "-"):
			return a
		}
	}
	return ""
}

// With returns a copy of r with extra environment variables.
func (r Runner) With(env ...string) Runner {
	r.Env = append(append([]string{}, r.Env...), env...)
	return r
}

// Output runs git and returns stdout with the trailing newline removed.
func (r Runner) Output(args ...string) (string, error) {
	out, err := r.OutputIn(nil, args...)
	return strings.TrimRight(out, "\n"), err
}

// OutputIn runs git with stdin and returns raw stdout.
func (r Runner) OutputIn(stdin io.Reader, args ...string) (string, error) {
	c := r.cmd(args...)
	c.Stdin = stdin
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	if err := c.Run(); err != nil {
		return out.String(), &Error{Args: args, Stderr: errb.String(), Err: err}
	}
	return out.String(), nil
}

// Stream runs git with stdin and hands stdout to read as it arrives, so large output is never
// held in memory. Whatever read leaves unread is discarded.
func (r Runner) Stream(stdin io.Reader, read func(stdout io.Reader) error, args ...string) error {
	c := r.cmd(args...)
	c.Stdin = stdin
	var errb bytes.Buffer
	c.Stderr = &errb
	out, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return &Error{Args: args, Err: err}
	}
	rerr := read(out)
	_, _ = io.Copy(io.Discard, out) // let git finish writing
	if err := c.Wait(); err != nil {
		return &Error{Args: args, Stderr: errb.String(), Err: err}
	}
	return rerr
}

// OK reports whether git exits 0.
func (r Runner) OK(args ...string) bool {
	_, err := r.Output(args...)
	return err == nil
}

// Interactive runs git attached to the terminal, so editors and pagers work.
// Stderr is also captured and returned in the error on failure.
func (r Runner) Interactive(args ...string) error {
	c := r.cmd(args...)
	c.Stdin, c.Stdout = os.Stdin, os.Stdout
	var errb bytes.Buffer
	c.Stderr = io.MultiWriter(os.Stderr, &errb)
	if err := c.Run(); err != nil {
		return &Error{Args: args, Stderr: errb.String(), Err: err}
	}
	return nil
}

// ExitCode returns the exit code of a failed git command, or -1.
func ExitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// versionRe finds "major.minor" in the output of "git version".
var versionRe = regexp.MustCompile(`(\d+)\.(\d+)`)

// Version returns git's major and minor version.
func Version() (major, minor int, err error) {
	out, err := Runner{}.Output("version")
	if err != nil {
		return 0, 0, err
	}
	m := versionRe.FindStringSubmatch(out)
	if m == nil {
		return 0, 0, fmt.Errorf("can't read git version from %q", out)
	}
	major, _ = strconv.Atoi(m[1])
	minor, _ = strconv.Atoi(m[2])
	return major, minor, nil
}

// Repo describes the repository around a directory.
type Repo struct {
	Runner
	Top    string // working tree root
	GitDir string // absolute path of this worktree's git dir
	Common string // absolute path of the shared git dir
}

// Open finds the repository containing dir.
func Open(dir string) (*Repo, error) {
	r := Runner{Dir: dir}
	out, err := r.Output("rev-parse", "--show-toplevel", "--absolute-git-dir", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, err
	}
	parts := strings.Split(out, "\n")
	if len(parts) != 3 {
		return nil, fmt.Errorf("not a working tree (bare repo?)")
	}
	return &Repo{Runner: Runner{Dir: parts[0]}, Top: parts[0], GitDir: parts[1], Common: parts[2]}, nil
}
