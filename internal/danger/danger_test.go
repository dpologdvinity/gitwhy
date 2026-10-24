package danger

import (
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestWarnings checks how many warnings each common command line gets: risky forms warn, their
// safe twins (dry runs, --soft, --cached, --staged, plain push) stay silent, and abbreviated long
// flags ("--ha", "--forc") count as the full flag. Too many warnings teach users to ignore them;
// too few let work be lost.
func TestWarnings(t *testing.T) {
	cases := map[string]int{
		"reset --hard HEAD~1":          1,
		"reset --soft HEAD~1":          0,
		"clean -fdx":                   2,
		"clean -n":                     0,
		"clean -dn":                    0,
		"push origin main":             0,
		"push --force-with-lease":      1,
		"push origin +main":            1,
		"push origin --delete feature": 1,
		"push origin :feature":         1,
		"stash drop":                   1,
		"stash list":                   0,
		"restore --staged a.txt":       0,
		"restore a.txt":                1,
		"rm --cached a.txt":            0,
		"rm a.txt":                     1,
		"branch -d x":                  0,
		"branch -D x":                  1,
		"rebase --continue":            0,
		"status":                       0,
		"worktree remove --force ../x": 2,
		"reset --ha":                   1,
		"push --forc":                  1,
		"checkout .":                   1,
		"checkout main a.txt":          1,
		"checkout main":                0,
		"checkout -b new":              0,
		"switch -C main":               1,
		"mv -f a b":                    1,
		"clean -d":                     1,
	}
	for cmd, want := range cases {
		if got := Warnings(strings.Fields(cmd)); len(got) != want {
			t.Errorf("%q: got %d warnings %q, want %d", cmd, len(got), got, want)
		}
	}
}

// TestCommandsListsRuleCommands checks that Commands includes commands from the rules table. The
// shell hook skips built-in commands missing from this list, so they would never get a snapshot.
func TestCommandsListsRuleCommands(t *testing.T) {
	got := Commands()
	has := map[string]bool{}
	for _, c := range got {
		has[c] = true
	}
	for _, want := range []string{"reset", "clean", "rebase", "checkout"} {
		if !has[want] {
			t.Errorf("Commands() missing %q: %v", want, got)
		}
	}
}

// TestShortFlagsWithAttachedValues checks that the letters after a value-taking short flag are
// that flag's value, not more flags. Reading them as flags made `clean -fd -ekeep.env` look like
// a dry run (the pattern contains an "n") and `checkout -bfix` look like `--force`.
func TestShortFlagsWithAttachedValues(t *testing.T) {
	cases := map[string]int{
		"clean -fd -ekeep.env": 1, // deletes untracked files; the "n" in the pattern is not -n
		"clean -fdekeep.env":   1,
		"clean -fdn":           0, // a real dry run
		"checkout -bfeature":   0, // the "f" in the branch name is not --force
		"checkout -Bfix":       1, // -B resets a branch: still warned
		"switch -cfix":         0,
		"switch -Cfix":         1,
		"commit -amend":        0, // -a -m "end": not --amend
		"commit -mfix --amend": 1,
		"tag -mdelete v1":      0, // the "d" in the message is not -d
		"tag -d v1":            1,
	}
	for cmd, want := range cases {
		if got := Warnings(strings.Fields(cmd)); len(got) != want {
			t.Errorf("%q: got %d warnings %q, want %d", cmd, len(got), got, want)
		}
	}
}

// TestEmptyDashIsNotAPathList checks that a trailing "--" with no paths after it does not make
// the command look like it overwrites files.
func TestEmptyDashIsNotAPathList(t *testing.T) {
	cases := map[string]int{
		"checkout main --":       0,
		"checkout -- f.txt":      1,
		"checkout main -- f.txt": 1,
	}
	for cmd, want := range cases {
		if got := Warnings(strings.Fields(cmd)); len(got) != want {
			t.Errorf("%q: got %d warnings %q, want %d", cmd, len(got), got, want)
		}
	}
}

// TestWarningsInTreatsAPlainCheckoutArgumentAsAPath checks that `git checkout f.txt` (the most
// common way to throw away edits) is warned about when f.txt is not a branch, tag or commit,
// while `git checkout main` and a remote-only branch name are not.
func TestWarningsInTreatsAPlainCheckoutArgumentAsAPath(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Git("branch", "feature")
	r.Git("update-ref", "refs/remotes/origin/remote-only", "HEAD")
	cases := map[string]int{
		"checkout f.txt":       1, // a path: edits are overwritten
		"checkout feature":     0, // a local branch
		"checkout remote-only": 0, // git creates a tracking branch
		"checkout HEAD":        0,
		"checkout -b new":      0,
		"checkout -f feature":  1, // already covered by the force rule
	}
	for cmd, want := range cases {
		if got := WarningsIn(r.Repo, strings.Fields(cmd)); len(got) != want {
			t.Errorf("%q: got %d warnings %q, want %d", cmd, len(got), got, want)
		}
	}
	// Without a repository it behaves exactly like Warnings.
	if got := WarningsIn(nil, strings.Fields("checkout f.txt")); len(got) != 0 {
		t.Errorf("nil repo: got %q, want none", got)
	}
}

// TestFileNamedLikeASubcommandIsStillAFile checks that a word such as "set" or "add" is only read
// as a subcommand for commands that have subcommands. `git checkout set` (a file called set) must
// still be warned about as overwriting that file.
func TestFileNamedLikeASubcommandIsStillAFile(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "set", "1\n", "add", "1\n")
	for _, name := range []string{"set", "add"} {
		if got := WarningsIn(r.Repo, []string{"checkout", name}); len(got) != 1 {
			t.Errorf("checkout %s: got %q, want the overwrite warning", name, got)
		}
	}
	if got := Warnings(strings.Fields("stash drop")); len(got) != 1 {
		t.Errorf("stash drop: got %q, want 1 warning (stash still has subcommands)", got)
	}
}
