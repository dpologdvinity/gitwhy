// Package explain breaks a git command line into parts with a plain-English meaning for each.
//
// Flag meanings come from the embedded command docs (package kb). Positional arguments are
// described by Revision, which recognises special names, ranges and revision suffixes, and asks
// the repository whether a word is a branch, tag, remote, commit or tracked file.
package explain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/git"
	"github.com/dpologdvinity/gitwhy/internal/kb"
)

// Part is one token of the command line and what it means.
type Part struct {
	Token   string // the word as typed; for "bisect run" the whole test command
	Meaning string // plain-English meaning
}

// Patterns Revision uses to recognise revision syntax.
var (
	shaRe    = regexp.MustCompile(`^[0-9a-f]{7,64}$`) // a full or abbreviated commit ID (SHA-1 or SHA-256)
	tildeRe  = regexp.MustCompile(`^(.*?)~(\d*)$`)    // "<rev>~<n>": n first-parent steps back (default 1)
	caretRe  = regexp.MustCompile(`^(.*?)\^(\d*)$`)   // "<rev>^<n>": the n-th parent (default 1)
	reflogRe = regexp.MustCompile(`^(.*?)@\{(.+)\}$`) // "<ref>@{...}": reflog entry, upstream, push or stash
)

// Explain describes each part of args (without the leading "git"). repo may be nil; then
// positional arguments are described from their syntax alone. An unknown command gives a single
// part, and flags missing from the docs are still listed, with a pointer to git's help.
func Explain(k *kb.KB, repo *git.Repo, args []string) []Part {
	if len(args) == 0 {
		return nil
	}
	cmd := args[0]
	c := k.Commands[cmd]
	var parts []Part
	if c == nil {
		return append(parts, Part{cmd, "not a git command gitwhy knows (maybe an alias or a separate tool)"})
	}
	parts = append(parts, Part{cmd, firstSentence(c.Summary)})
	rest := args[1:]
	sub := ""
	// A documented subcommand, such as "pop" in "stash pop", has its own flags in the docs.
	if len(rest) > 0 && c.Sub(rest[0]) != nil {
		sub = rest[0]
		parts = append(parts, Part{sub, firstSentence(c.Sub(sub).Summary)})
		rest = rest[1:]
	}
	afterDash := false
	for i := 0; i < len(rest); i++ {
		tok := rest[i]
		// Everything after "bisect run" is the user's own test command, not git flags.
		if cmd == "bisect" && sub == "run" {
			parts = append(parts, Part{strings.Join(rest[i:], " "), "the test command git runs on each commit (exit 0 = good)"})
			break
		}
		switch {
		case afterDash:
			parts = append(parts, Part{tok, "file or folder to act on"})
		case tok == "--":
			afterDash = true
			parts = append(parts, Part{tok, "end of options; what follows are file paths"})
		case strings.HasPrefix(tok, "-") && len(tok) > 1:
			f, ok := k.Lookup(cmd, sub, tok)
			if ok {
				meaning := f.Desc
				// Docs describe the positive form; "--no-x" turns it off.
				if name, _, _ := strings.Cut(tok, "="); strings.HasPrefix(name, "--no-") && strings.Contains(f.Display, "[no-]") {
					meaning = "Turns off " + "--" + strings.TrimPrefix(name, "--no-") + ": " + f.Desc
				}
				parts = append(parts, Part{tok, meaning})
				// "-m <msg>": the next word is this flag's value.
				if f.Value != "" && !f.Optional && !strings.Contains(tok, "=") && i+1 < len(rest) && takesSeparateValue(f, tok) {
					i++
					parts = append(parts, Part{rest[i], "value for " + tok})
				}
				continue
			}
			// Combined short flags: "-fdx". Only used when every letter is a documented flag;
			// otherwise the whole word is shown as one unknown option.
			if !strings.HasPrefix(tok, "--") && len(tok) > 2 {
				all := true
				var sub2 []Part
				for _, ch := range tok[1:] {
					f, ok := k.Lookup(cmd, sub, "-"+string(ch))
					if !ok {
						all = false
						break
					}
					sub2 = append(sub2, Part{"-" + string(ch), f.Desc})
				}
				if all {
					parts = append(parts, sub2...)
					continue
				}
			}
			parts = append(parts, Part{tok, "option (not in gitwhy's docs; see git " + cmd + " --help)"})
		default:
			parts = append(parts, Part{tok, Revision(repo, tok)})
		}
	}
	return parts
}

// takesSeparateValue reports whether a flag written as tok takes its value as the next word.
// Short flags written with a value ("-m <msg>") do; long flags documented as "--x=<v>" also
// accept "--x <v>" in git, so they do too.
func takesSeparateValue(f kb.Flag, tok string) bool {
	return !strings.HasPrefix(tok, "--") || strings.Contains(f.Display, tok+" ") || strings.Contains(f.Display, tok+"=")
}

// Revision explains a positional argument: a revision, range, remote or path. Syntax is checked
// first ("a...b" before "a..b", then "@{...}", "~" and "^", recursing on the base). If repo is not
// nil, git is then asked what the word names; otherwise a generic description is returned.
func Revision(repo *git.Repo, tok string) string {
	switch tok {
	case "HEAD", "@":
		return "the commit you're on now"
	case "-":
		return "the branch you were on before"
	case ".":
		return "the current folder (everything in it)"
	case "ORIG_HEAD":
		return "where HEAD was before the last reset, merge or rebase"
	case "FETCH_HEAD":
		return "what the last fetch downloaded"
	case "MERGE_HEAD":
		return "the commit being merged in"
	}
	if a, b, ok := strings.Cut(tok, "..."); ok {
		return "commits on " + orHEAD(a) + " or " + orHEAD(b) + " but not both (for diff: changes on " + orHEAD(b) + " since they split)"
	}
	if a, b, ok := strings.Cut(tok, ".."); ok {
		return "commits on " + orHEAD(b) + " that aren't on " + orHEAD(a)
	}
	if m := reflogRe.FindStringSubmatch(tok); m != nil {
		base := orBranch(m[1])
		switch m[2] {
		case "u", "upstream":
			return "the remote branch " + base + " tracks"
		case "push":
			return "where " + base + " pushes to"
		}
		if strings.HasPrefix(tok, "stash@") {
			return "stash number " + m[2] + " (0 = newest)"
		}
		if n, err := strconv.Atoi(m[2]); err == nil {
			if n < 0 && m[1] == "" { // "@{-N}": the branch checked out N switches ago
				return fmt.Sprintf("the branch you were on %d %s ago", -n, plural(-n, "switch", "switches"))
			}
			return base + " as it was " + m[2] + " moves ago (from the reflog)"
		}
		return base + " as it was " + m[2] + " (from the reflog)" // a date such as "yesterday"
	}
	if m := tildeRe.FindStringSubmatch(tok); m != nil && m[1] != "" {
		n := m[2]
		if n == "" {
			n = "1"
		}
		return n + " commit(s) before " + Revision(repo, m[1])
	}
	if m := caretRe.FindStringSubmatch(tok); m != nil && m[1] != "" {
		if m[2] == "" || m[2] == "1" {
			return "the commit before " + Revision(repo, m[1])
		}
		return "parent number " + m[2] + " of " + Revision(repo, m[1]) + " (a merge commit)"
	}
	// Order matters when one name is several things: a local branch wins over a remote branch,
	// tag, remote, commit ID or file.
	if repo != nil {
		switch {
		case repo.OK("show-ref", "--verify", "-q", "refs/heads/"+tok):
			return "branch " + tok
		case repo.OK("show-ref", "--verify", "-q", "refs/remotes/"+tok):
			return "remote branch " + tok + " (as of your last fetch)"
		case repo.OK("show-ref", "--verify", "-q", "refs/tags/"+tok):
			return "tag " + tok
		case repo.OK("remote", "get-url", tok):
			return "remote " + tok
		case shaRe.MatchString(tok) && repo.OK("cat-file", "-e", tok+"^{commit}"):
			return "commit " + tok
		}
		// --error-unmatch makes ls-files fail when tok matches no tracked file.
		if _, err := repo.Output("ls-files", "--error-unmatch", "--", tok); err == nil {
			return "file " + tok
		}
	}
	if shaRe.MatchString(tok) {
		return "a commit ID"
	}
	return "a branch, tag, commit, remote or path"
}

// orHEAD returns s, or "HEAD" when s is empty, since git reads a missing side of a range as HEAD.
// plural returns one when n is 1, else many.
func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func orHEAD(s string) string {
	if s == "" {
		return "HEAD"
	}
	return s
}

// orBranch returns s, or "the current branch" when s is empty, as in "@{u}".
func orBranch(s string) string {
	if s == "" {
		return "the current branch"
	}
	return s
}

// firstSentence returns s up to and including the first ". ", or all of s if there is none.
func firstSentence(s string) string {
	if i := strings.Index(s, ". "); i > 0 {
		return s[:i+1]
	}
	return s
}
