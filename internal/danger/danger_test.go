package danger

import (
	"strings"
	"testing"
)

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
