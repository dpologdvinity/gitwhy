package recipe

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/git"
)

// checks are situation tests recipes can reference by name.
var checks = map[string]func(*git.Repo) bool{
	// Uncommitted changes to tracked files, staged or not.
	"dirty-tree": func(r *git.Repo) bool {
		out, _ := r.Output("status", "--porcelain", "--untracked-files=no")
		return out != ""
	},
	// Any untracked (not ignored) files.
	"untracked-files": func(r *git.Repo) bool {
		out, _ := r.Output("ls-files", "--others", "--exclude-standard", "--directory", "--no-empty-directory")
		return out != ""
	},
	// HEAD is reachable from some remote-tracking branch.
	"head-is-pushed": func(r *git.Repo) bool {
		out, _ := r.Output("branch", "-r", "--contains", "HEAD")
		return out != ""
	},
	// A merge, rebase, cherry-pick, revert, am or bisect is in progress.
	"in-progress-op": func(r *git.Repo) bool { return InProgress(r) != "" },
	"detached-head": func(r *git.Repo) bool {
		return !r.OK("symbolic-ref", "-q", "HEAD")
	},
	// Nothing is staged for the next commit.
	"nothing-staged": func(r *git.Repo) bool {
		return r.OK("diff", "--cached", "--quiet")
	},
	"has-conflicts": func(r *git.Repo) bool {
		out, _ := r.Output("diff", "--name-only", "--diff-filter=U")
		return out != ""
	},
	"no-upstream": func(r *git.Repo) bool {
		return !r.OK("rev-parse", "--abbrev-ref", "@{upstream}")
	},
	// HEAD has no parent (first commit) or there are no commits.
	"no-parent": func(r *git.Repo) bool {
		return !r.OK("rev-parse", "--verify", "-q", "HEAD~1")
	},
	"no-commits": func(r *git.Repo) bool {
		return !r.OK("rev-parse", "--verify", "-q", "HEAD")
	},
	"filter-repo-missing": func(r *git.Repo) bool {
		return !r.OK("filter-repo", "--version")
	},
	"lfs-missing": func(r *git.Repo) bool {
		return !r.OK("lfs", "version")
	},
}

// Test runs a named check; a "!" prefix negates it.
func Test(r *git.Repo, name string) bool {
	neg := strings.HasPrefix(name, "!")
	f := checks[strings.TrimPrefix(name, "!")]
	return f != nil && f(r) != neg
}

// InProgress names the operation in progress ("merge", "rebase", ...) or "".
func InProgress(r *git.Repo) string {
	exists := func(p string) bool { _, err := os.Stat(filepath.Join(r.GitDir, p)); return err == nil }
	switch {
	case exists("rebase-merge"), exists("rebase-apply/rebasing"):
		return "rebase"
	case exists("rebase-apply"):
		return "am"
	case exists("MERGE_HEAD"):
		return "merge"
	case exists("CHERRY_PICK_HEAD"):
		return "cherry-pick"
	case exists("REVERT_HEAD"):
		return "revert"
	case exists("BISECT_LOG"):
		return "bisect"
	}
	return ""
}
