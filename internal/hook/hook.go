// Package hook gives plain `git` commands undo points: a shell function calls Pre before
// each git command and Post after it.
package hook

import (
	"os"
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
	s, err := snapshot.Take(r, "git "+strings.Join(args, " "), preview.Preview(r, args).Save)
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
