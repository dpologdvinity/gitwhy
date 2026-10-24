// Package danger warns about git commands that can lose work or rewrite shared history.
package danger

import (
	"sort"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Args is a parsed command line (without the leading "git").
type Args struct {
	Cmd   string
	Sub   string
	Words []string // everything after Cmd
	short map[byte]bool
	long  map[string]bool
}

// valueShort lists, per command, the short flags that take a value. In "-ekeep.env" the letters
// after "e" are its value, not more flags, so Parse stops reading letters there. Without this,
// the "n" in "keep.env" would look like -n (dry run) and "-bfix" would look like -f (force).
var valueShort = map[string]string{
	"checkout": "bB",
	"switch":   "cC",
	"clean":    "e",
	"commit":   "mFCcStu",
	"tag":      "mFu",
	"push":     "o",
	"rebase":   "xsXSC",
	"merge":    "msXFS",
	"stash":    "m",
	"restore":  "s",
	"branch":   "u",
}

// Parse splits args into command, optional subcommand and flags. Combined short flags ("-fdx")
// count as each letter, up to a flag that takes a value (see valueShort).
func Parse(args []string, subs map[string]bool) Args {
	a := Args{short: map[byte]bool{}, long: map[string]bool{}}
	if len(args) == 0 {
		return a
	}
	a.Cmd, a.Words = args[0], args[1:]
	for i, w := range a.Words {
		if w == "--" {
			break
		}
		switch {
		case strings.HasPrefix(w, "--"):
			name, _, _ := strings.Cut(w, "=")
			a.long[name] = true
		case strings.HasPrefix(w, "-") && len(w) > 1:
			for j := 1; j < len(w); j++ {
				a.short[w[j]] = true
				if strings.IndexByte(valueShort[a.Cmd], w[j]) >= 0 {
					break // the rest of the word is this flag's value
				}
			}
		case i == 0 && subs[w]:
			a.Sub = w
		}
	}
	return a
}

// Has reports whether any of the flags is present. Short flags are one letter after "-". Long
// flags also match git's accepted abbreviations ("--ha" for "--hard").
func (a Args) Has(flags ...string) bool {
	for _, f := range flags {
		if strings.HasPrefix(f, "--") {
			if a.long[f] {
				return true
			}
			for l := range a.long {
				if len(l) >= 4 && strings.HasPrefix(f, l) {
					return true
				}
			}
		} else if len(f) == 2 && a.short[f[1]] {
			return true
		}
	}
	return false
}

// Positional returns non-flag words (excluding the subcommand), stopping at "--".
func (a Args) Positional() []string {
	var out []string
	for i, w := range a.Words {
		if w == "--" {
			break
		}
		if i == 0 && w == a.Sub {
			continue
		}
		if !strings.HasPrefix(w, "-") {
			out = append(out, w)
		}
	}
	return out
}

// AfterDash returns the words after "--".
func (a Args) AfterDash() []string {
	for i, w := range a.Words {
		if w == "--" {
			return a.Words[i+1:]
		}
	}
	return nil
}

type rule struct {
	cmd  string
	when func(Args) bool
	msg  string
}

func has(flags ...string) func(Args) bool { return func(a Args) bool { return a.Has(flags...) } }
func sub(names ...string) func(Args) bool {
	return func(a Args) bool {
		for _, n := range names {
			if a.Sub == n {
				return true
			}
		}
		return false
	}
}
func always(Args) bool { return true }

const unrecoverable = " gitwhy can't undo this."

// overwritesFiles is the warning for commands that replace files with another version.
const overwritesFiles = "Overwrites changes in those files."

var rules = []rule{
	{"reset", has("--hard"), "Deletes every uncommitted change to tracked files."},
	{"reset", has("--merge"), "Staged changes are lost."},
	{"checkout", has("-f", "--force"), "Discards local changes."},
	{"checkout", func(a Args) bool {
		// "checkout -- f", "checkout .", "checkout <rev> <path>": overwrite files. A lone path
		// ("checkout f") needs the repository to tell it from a branch: see WarningsIn.
		return len(a.AfterDash()) > 0 || a.Has("--ours", "--theirs") || contains(a.Positional(), ".") || (len(a.Positional()) >= 2 && !a.Has("-b", "-B", "--orphan"))
	}, overwritesFiles},
	{"checkout", has("-B"), "Resets an existing branch to a new start point."},
	{"switch", has("-C", "--force-create"), "Resets an existing branch to a new start point."},
	{"mv", has("-f", "--force"), "Overwrites the destination file."},
	{"submodule", func(a Args) bool { return a.Sub == "update" && a.Has("-f", "--force") }, "Discards local changes in submodules."},
	{"restore", func(a Args) bool { return !a.Has("-S", "--staged") || a.Has("-W", "--worktree") }, "Discards unstaged changes in those files."},
	{"switch", has("-f", "--force", "--discard-changes"), "Discards local changes and overwrites untracked files in the way."},
	{"clean", func(a Args) bool { return !a.Has("-n", "--dry-run", "-i", "--interactive") }, "Deletes untracked files from disk."},
	{"clean", has("-x", "-X"), "Also deletes ignored files (.env, build output, node_modules)."},
	{"rm", func(a Args) bool { return !a.Has("--cached") }, "Deletes the files from disk too."},
	{"branch", has("-D"), "Deletes the branch even if its commits aren't merged anywhere."},
	{"branch", func(a Args) bool { return a.Has("-d", "--delete") && a.Has("-f", "--force") }, "Deletes the branch even if its commits aren't merged anywhere."},
	{"branch", has("-f", "--force", "-C", "-M"), "Moves or overwrites an existing branch."},
	{"stash", sub("drop", "clear"), "Deletes stashed changes."},
	{"stash", sub("pop"), "Removes the stash after applying (kept if it conflicts)."},
	{"commit", has("--amend"), "Replaces the last commit (new ID); force-push needed if already pushed."},
	{"rebase", func(a Args) bool {
		return !a.Has("--abort", "--continue", "--skip", "--quit", "--edit-todo", "--show-current-patch")
	}, "Rewrites commits (new IDs); force-push needed if already pushed."},
	{"rebase", has("--abort"), "Discards the rebase, including conflict fixes."},
	{"merge", has("--abort"), "Discards the merge, including conflict fixes."},
	{"cherry-pick", has("--abort"), "Discards the cherry-pick, including conflict fixes."},
	{"am", has("--abort"), "Discards the patch series in progress."},
	{"push", func(a Args) bool {
		if a.Has("-f", "--force", "--force-with-lease", "--force-if-includes") {
			return true
		}
		for _, w := range a.Positional() {
			if strings.HasPrefix(w, "+") {
				return true
			}
		}
		return false
	}, "Overwrites the remote branch; others' new commits there can be lost." + unrecoverable},
	{"push", func(a Args) bool {
		if a.Has("-d", "--delete") {
			return true
		}
		for _, w := range a.Positional() {
			if strings.HasPrefix(w, ":") {
				return true
			}
		}
		return false
	}, "Deletes branches or tags on the remote." + unrecoverable},
	{"push", has("--mirror", "--prune"), "Deletes remote branches you don't have locally." + unrecoverable},
	{"push", has("--mirror"), "Also pushes gitwhy's undo refs (refs/gitwhy/*)."},
	{"filter-repo", always, "Rewrites history; commit IDs change. Without --refs it also expires old history and removes the 'origin' remote."},
	{"filter-branch", always, "Rewrites history; every affected commit ID changes."},
	{"lfs", sub("migrate"), "Can rewrite history (import/export); commit IDs change."},
	{"lfs", sub("prune"), "Deletes local LFS files not needed by recent commits."},
	{"gc", func(a Args) bool {
		for _, w := range a.Words {
			if strings.HasPrefix(w, "--prune=now") || w == "--prune=all" {
				return true
			}
		}
		return false
	}, "Permanently deletes unreachable commits." + unrecoverable},
	{"prune", always, "Permanently deletes unreachable objects." + unrecoverable},
	{"reflog", sub("expire", "delete"), "Deletes history git uses for recovery." + unrecoverable},
	{"worktree", sub("remove"), "Deletes the worktree folder, including ignored files in it."},
	{"worktree", func(a Args) bool { return a.Sub == "remove" && a.Has("-f", "--force") }, "Even with uncommitted changes, which are lost." + unrecoverable},
	{"update-ref", has("-d"), "Deletes a ref."},
	{"tag", has("-d", "--delete"), "Deletes the tag locally."},
	{"tag", has("-f", "--force"), "Moves an existing tag."},
	{"remote", sub("remove", "rm"), "Removes the remote and all its remote-tracking branches."},
	{"sparse-checkout", sub("set", "add", "reapply"), "Files outside the chosen folders vanish from disk; ignored files in removed folders are deleted."},
	{"submodule", func(a Args) bool { return a.Sub == "deinit" && a.Has("-f", "--force") }, "Deletes the submodule's local changes and untracked files." + unrecoverable},
	{"read-tree", has("--reset"), "Overwrites the index; with -u, files too."},
	{"replace", always, "Changes how git sees existing commits everywhere in this repo."},
}

// Subcommands git commands take, so "stash drop" is read as a subcommand.
var Subs = map[string]bool{
	"drop": true, "clear": true, "pop": true, "apply": true, "push": true, "list": true, "show": true,
	"migrate": true, "prune": true, "expire": true, "delete": true, "remove": true, "rm": true, "add": true,
	"set": true, "reapply": true, "deinit": true, "update": true, "init": true, "track": true,
}

// Commands lists every git command that has a warning rule, sorted.
func Commands() []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rules {
		if !seen[r.cmd] {
			seen[r.cmd] = true
			out = append(out, r.cmd)
		}
	}
	sort.Strings(out)
	return out
}

// Warnings returns plain-English warnings for a command line (without the leading "git").
func Warnings(args []string) []string {
	a := Parse(args, Subs)
	var out []string
	seen := map[string]bool{}
	for _, r := range rules {
		if r.cmd == a.Cmd && r.when(a) && !seen[r.msg] {
			seen[r.msg] = true
			out = append(out, r.msg)
		}
	}
	return out
}

// WarningsIn is Warnings plus the cases that need the repository to decide. A lone argument to
// "git checkout" is a path (and overwrites its edits) unless it names a commit, branch, tag or
// remote branch; Warnings cannot tell, so it stays silent. With a nil repo this is Warnings.
func WarningsIn(repo *git.Repo, args []string) []string {
	out := Warnings(args)
	if repo == nil {
		return out
	}
	a := Parse(args, Subs)
	if a.Cmd != "checkout" || a.Has("-b", "-B", "--orphan", "--detach", "-t", "--track") || len(a.AfterDash()) > 0 {
		return out
	}
	pos := a.Positional()
	if len(pos) != 1 || namesARef(repo, pos[0]) || contains(out, overwritesFiles) {
		return out
	}
	return append(out, overwritesFiles)
}

// namesARef reports whether x is something "git checkout x" would switch to: a commit, branch or
// tag, or a remote-only branch that git would create locally.
func namesARef(repo *git.Repo, x string) bool {
	if repo.OK("rev-parse", "--verify", "-q", x+"^{commit}") {
		return true
	}
	refs, err := repo.Output("for-each-ref", "--format=%(refname)", "refs/remotes")
	if err != nil {
		return false
	}
	for _, ref := range strings.Split(refs, "\n") {
		// refs/remotes/<remote>/<branch>: git's DWIM matches <branch> in any remote.
		if parts := strings.SplitN(strings.TrimPrefix(ref, "refs/remotes/"), "/", 2); len(parts) == 2 && parts[1] == x {
			return true
		}
	}
	return false
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
