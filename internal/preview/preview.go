// Package preview says what a git command will do before it runs.
//
// Exact previews use git's own dry runs and plumbing (merge-tree, rev-list, diff) without touching
// the repository. Merge, rebase, cherry-pick and revert are simulated with "git merge-tree", so
// conflicts are reported before they happen. Commands the package does not handle get an empty
// Result; callers fall back to recipe text. A clean preview also returns the files the command
// will delete, ignored ones included, so the caller can add them to the snapshot.
package preview

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/danger"
	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Result is the outcome of previewing one command. A Result with no Lines means "no preview".
type Result struct {
	Exact bool     // computed, not guessed
	Lines []string // what will happen
	Save  []string // paths clean will delete (ignored or untracked), relative to the repo root; include them in the snapshot
}

// add appends one formatted line to the preview.
func (r *Result) add(format string, a ...any) { r.Lines = append(r.Lines, fmt.Sprintf(format, a...)) }

// withGlobs returns a copy of repo that lets git expand pathspec globs. gitwhy's own git calls may
// treat file names literally, but the command being previewed was typed by the user and the real
// git will expand its globs, so the preview must too.
func withGlobs(repo *git.Repo) *git.Repo {
	c := *repo
	c.Runner = c.With("GIT_LITERAL_PATHSPECS=0")
	return &c
}

// Preview computes what args (without the leading "git") will do in repo. It never changes the
// repository. It returns a Result with no Lines when it cannot say anything useful.
func Preview(repo *git.Repo, args []string) Result {
	repo = withGlobs(repo)
	a := danger.Parse(args, danger.Subs)
	switch a.Cmd {
	case "reset":
		return reset(repo, a)
	case "clean":
		return dryRun(repo, args, "-n", "Would remove ", "will delete", true)
	case "rm":
		return dryRun(repo, args, "-n", "rm '", "will remove", false)
	case "add":
		return dryRun(repo, args, "--dry-run", "add '", "will stage", false)
	case "mv":
		return dryRun(repo, args, "-n", "Renaming ", "will move", false)
	case "merge":
		// Only a fresh merge of one commit can be simulated; these act on a merge in progress.
		if a.Has("--abort", "--continue", "--quit") {
			return Result{}
		}
		if p := a.Positional(); len(p) == 1 {
			return merge(repo, p[0], mergeOptionsOf(repo, a))
		}
	case "rebase":
		// --onto and --root change which commits are replayed, so the simple
		// upstream..HEAD simulation below would be wrong for them.
		if a.Has("--abort", "--continue", "--skip", "--quit", "--onto", "--root") {
			return Result{}
		}
		if p := a.Positional(); len(p) == 1 {
			return rebase(repo, p[0], a.Has("-i", "--interactive"))
		}
	case "cherry-pick":
		if a.Has("--abort", "--continue", "--skip", "--quit") {
			return Result{}
		}
		return cherryPick(repo, a.Positional())
	case "revert":
		if p := a.Positional(); len(p) == 1 && !a.Has("--abort", "--continue", "--skip", "--quit") {
			return revert(repo, p[0])
		}
	case "branch":
		return branch(repo, a)
	case "stash":
		return stash(repo, a)
	case "switch", "checkout":
		return switchBranch(repo, a)
	case "push":
		return push(repo, a)
	case "commit":
		if a.Has("--amend") {
			r := Result{Exact: true}
			r.add("The last commit %s is replaced by a new one.", subject(repo, "HEAD"))
			return r
		}
	}
	return Result{}
}

// short returns the abbreviated commit ID of rev, or rev itself if git cannot resolve it.
func short(repo *git.Repo, rev string) string {
	out, err := repo.Output("rev-parse", "--short", rev)
	if err != nil {
		return rev
	}
	return out
}

// subject returns rev as a short ID and quoted subject line, such as `abc1234 "Fix typo"`, or rev
// itself if git cannot resolve it.
func subject(repo *git.Repo, rev string) string {
	out, err := repo.Output("log", "-1", "--format=%h \"%s\"", rev, "--")
	if err != nil {
		return rev
	}
	return out
}

// commits lists up to 5 one-line commits ("<short id> <subject>") in a range, plus the total
// count. rng is passed to git as is, so it may hold several revisions and options such as --not.
// It returns nil and 0 if the range is empty or git fails.
func commits(repo *git.Repo, rng ...string) ([]string, int) {
	args := append([]string{"rev-list", "--count"}, rng...)
	out, err := repo.Output(append(args, "--")...)
	if err != nil {
		return nil, 0
	}
	n, _ := strconv.Atoi(out)
	if n == 0 {
		return nil, 0
	}
	args = append([]string{"log", "--format=%h %s", "-n", "5"}, rng...)
	list, _ := repo.Output(append(args, "--")...)
	return strings.Split(list, "\n"), n
}

// listCommits adds a "title (n):" line followed by the commits in list, and a "... and N more"
// line when n is larger than the list shown.
func listCommits(r *Result, title string, list []string, n int) {
	r.add("%s (%d):", title, n)
	for _, c := range list {
		r.add("    %s", c)
	}
	if n > len(list) {
		r.add("    ... and %d more", n-len(list))
	}
}

// changedFiles returns the paths "git diff --name-only <args>" prints: with one commit, tracked
// files whose working-tree version differs from it; with two, files that differ between them.
// Untracked files are not included. It returns nil on error.
func changedFiles(repo *git.Repo, args ...string) []string {
	out, err := repo.Output(append([]string{"diff", "--name-only"}, args...)...)
	if err != nil || out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// listFiles adds a "title (n):" line followed by at most 8 files. It adds nothing when files is
// empty.
func listFiles(r *Result, title string, files []string) {
	if len(files) == 0 {
		return
	}
	r.add("%s (%d):", title, len(files))
	for i, f := range files {
		if i == 8 {
			r.add("    ... and %d more", len(files)-8)
			break
		}
		r.add("    %s", f)
	}
}

// reset previews "git reset". With paths (or --patch) it only unstages. Otherwise it reports where
// the branch moves, the commits it leaves and gains, and what happens to files for each mode.
func reset(repo *git.Repo, a danger.Args) Result {
	r := Result{Exact: true}
	pos := a.Positional()
	paths := a.AfterDash()
	target := "HEAD"
	// The first positional is the commit; any more are paths given without "--".
	if len(pos) > 0 {
		target = pos[0]
		paths = append(paths, pos[1:]...)
	}
	if len(paths) > 0 || a.Has("-p", "--patch") {
		r.add("Unstages changes in the given files (sets their staged version to %s). Your files on disk don't change.", target)
		return r
	}
	// Not a commit (for example a path given without "--"): no exact preview.
	if _, err := repo.Output("rev-parse", "--verify", "-q", target+"^{commit}"); err != nil {
		return Result{}
	}
	branch, _ := repo.Output("symbolic-ref", "--short", "-q", "HEAD")
	if branch == "" {
		branch = "HEAD (detached)"
	}
	r.add("%s moves from %s to %s.", branch, short(repo, "HEAD"), subject(repo, target))
	if list, n := commits(repo, target+"..HEAD"); n > 0 {
		listCommits(&r, "Commits leaving "+branch, list, n)
	}
	if list, n := commits(repo, "HEAD.."+target); n > 0 {
		listCommits(&r, "Commits added to "+branch, list, n)
	}
	moved := changedFiles(repo, "HEAD", target)
	switch {
	case a.Has("--soft"):
		listFiles(&r, "Their changes stay staged", moved)
	case a.Has("--hard"):
		// --hard drops uncommitted changes to tracked files; untracked files are left alone.
		listFiles(&r, "Files with uncommitted changes that will be RESET", changedFiles(repo, "HEAD"))
		listFiles(&r, "Files that change to match "+target, moved)
	case a.Has("--keep"):
		r.add("Uncommitted changes are kept; git stops if a file with changes would have to change.")
		listFiles(&r, "Files that change to match "+target, moved)
	case a.Has("--merge"):
		listFiles(&r, "Files that change to match "+target, moved)
	default: // --mixed
		listFiles(&r, "Their changes stay in your files, unstaged", moved)
	}
	return r
}

// dryRun runs the command itself with a dry-run flag and reports its output. The user's flags are
// passed through untouched (git lets -n win over -f, and a second -f changes what clean deletes).
// Only lines that start with prefix are file lines; git's progress lines are skipped.
func dryRun(repo *git.Repo, args []string, flag, prefix, verb string, isClean bool) Result {
	// core.quotePath=false keeps non-ASCII names readable; names with a double quote or newline
	// are still quoted, which unquote undoes for clean.
	dry := append([]string{"-c", "core.quotePath=false", args[0], flag}, args[1:]...)
	out, err := repo.Output(dry...)
	if err != nil {
		return Result{}
	}
	r := Result{Exact: true}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, prefix) || line == "" {
			continue
		}
		f := strings.TrimPrefix(line, prefix)
		if isClean {
			f = unquote(f)
		} else {
			// rm and add print "rm 'path'" and "add 'path'"; drop the closing quote.
			f = strings.TrimSuffix(f, "'")
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		r.add("Nothing to do.")
		return r
	}
	listFiles(&r, "Files git "+verb, files)
	// Only clean can delete ignored files, which a snapshot does not hold unless told to. Every
	// path it lists is handed over; the untracked ones are in the snapshot anyway.
	if isClean {
		r.Save = fromRepoRoot(repo, files)
	}
	return r
}

// unquote undoes git's C-style quoting of a path ("q\"uote.txt"). Git quotes names with special
// characters such as a double quote or a newline; other names are printed as they are.
func unquote(s string) string {
	if strings.HasPrefix(s, `"`) {
		if u, err := strconv.Unquote(s); err == nil {
			return u
		}
	}
	return s
}

// fromRepoRoot turns paths that git printed relative to the folder it ran in into paths relative
// to the repository root, which is where snapshot.Take reads them from.
func fromRepoRoot(repo *git.Repo, files []string) []string {
	prefix, err := repo.Output("rev-parse", "--show-prefix")
	if err != nil || prefix == "" {
		return files
	}
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = prefix + f
	}
	return out
}

// mergeTree runs "git merge-tree --write-tree" with args and returns the resulting tree ID and
// the conflicting paths. It writes only objects, never refs, the index or files. A path can be
// listed more than once and the list can hold empty lines, so callers pass conflicts through unique.
func mergeTree(repo *git.Repo, args ...string) (tree string, conflicts []string, err error) {
	out, err := repo.Output(append([]string{"merge-tree", "--write-tree", "--name-only", "--no-messages"}, args...)...)
	lines := strings.Split(out, "\n")
	// merge-tree exits 1 when the merge has conflicts; its output is still the tree ID followed by
	// the conflicting paths. Any other failure is a real error.
	if err != nil {
		if git.ExitCode(err) != 1 || len(lines) == 0 {
			return "", nil, err
		}
		return lines[0], lines[1:], nil
	}
	return lines[0], nil, nil
}

// mergeOptions are the merge flags that change what a merge does.
type mergeOptions struct {
	squash   bool // --squash: stage the changes, never commit or fast-forward
	noFF     bool // --no-ff: always create a merge commit
	ffOnly   bool // --ff-only: refuse unless the branch can fast-forward
	noCommit bool // --no-commit: stage the merge but do not commit it
}

// mergeOptionsOf reads the merge flags from a command line. Without an explicit --ff, --no-ff or
// --ff-only, git follows the merge.ff setting, so that is read too.
func mergeOptionsOf(repo *git.Repo, a danger.Args) mergeOptions {
	o := mergeOptions{squash: a.Has("--squash"), noFF: a.Has("--no-ff"), ffOnly: a.Has("--ff-only"), noCommit: a.Has("--no-commit")}
	if !o.noFF && !o.ffOnly && !a.Has("--ff") {
		switch cfg, _ := repo.Output("config", "--get", "merge.ff"); cfg {
		case "false":
			o.noFF = true
		case "only":
			o.ffOnly = true
		}
	}
	return o
}

// merge previews "git merge <from>": already up to date, a fast-forward, a refused --ff-only, or
// a real merge simulated with merge-tree to find conflicts.
func merge(repo *git.Repo, from string, o mergeOptions) Result {
	if !repo.OK("rev-parse", "--verify", "-q", from+"^{commit}") {
		return Result{}
	}
	r := Result{Exact: true}
	if repo.OK("merge-base", "--is-ancestor", from, "HEAD") {
		r.add("Already up to date: everything on %s is already here. Nothing changes.", from)
		return r
	}
	list, n := commits(repo, "HEAD.."+from)
	canFastForward := repo.OK("merge-base", "--is-ancestor", "HEAD", from)
	// --squash never fast-forwards and --no-ff always makes a merge commit, so both fall through
	// to the merge-tree simulation.
	if canFastForward && !o.squash && !o.noFF {
		r.add("Fast-forward: your branch just moves ahead to %s. No merge commit.", short(repo, from))
		listCommits(&r, "Commits coming in", list, n)
		return r
	}
	if o.ffOnly && !canFastForward && !o.squash {
		r.add("git refuses: this is not a fast-forward (your branch has commits that %s doesn't). Nothing changes.", from)
		return r
	}
	listCommits(&r, "Commits coming in", list, n)
	_, conflicts, err := mergeTree(repo, "HEAD", from)
	if err != nil {
		return Result{}
	}
	if len(conflicts) > 0 {
		listFiles(&r, "CONFLICTS expected in", unique(conflicts))
		r.add("git will stop for you to fix them (gitwhy resolve), or you can abort.")
	} else if o.squash {
		r.add("No conflicts. The changes are staged; you then make one commit.")
	} else if o.noCommit {
		r.add("No conflicts. The merge is staged but not committed; you commit it yourself.")
	} else {
		r.add("No conflicts. A merge commit is created.")
	}
	return r
}

// simulate replays commits (oldest first) onto base with merge-tree, without touching the repo.
// When revert is set, each commit is undone instead of applied. It returns the first commit that
// would conflict and its conflicting files, or "" when none would. A root commit (no parent) is
// an error, since there is no parent to diff it against. Each replayed step is written as a
// dangling commit with commit-tree so the next step can build on it; no ref points at it.
func simulate(repo *git.Repo, base string, list []string, revert bool) (string, []string, error) {
	cur := base
	for _, c := range list {
		var tree string
		var conflicts []string
		var err error
		// A cherry-pick is a three-way merge of c into cur with c's parent as the base; a revert
		// swaps the sides, using c as the base and its parent as the side merged in.
		if revert {
			tree, conflicts, err = mergeTree(repo, "--merge-base", c, cur, c+"^")
		} else if repo.OK("rev-parse", "--verify", "-q", c+"^") {
			tree, conflicts, err = mergeTree(repo, "--merge-base", c+"^", cur, c)
		} else {
			return "", nil, fmt.Errorf("root commit")
		}
		if err != nil {
			return "", nil, err
		}
		if len(conflicts) > 0 {
			return c, unique(conflicts), nil
		}
		cur, err = repo.With(identity...).Output("commit-tree", tree, "-p", cur, "-m", "gitwhy preview")
		if err != nil {
			return "", nil, err
		}
	}
	return "", nil, nil
}

// identity lets commit-tree work for previews even with no user.name configured.
var identity = []string{"GIT_AUTHOR_NAME=gitwhy", "GIT_AUTHOR_EMAIL=gitwhy@localhost", "GIT_COMMITTER_NAME=gitwhy", "GIT_COMMITTER_EMAIL=gitwhy@localhost"}

// revList returns the full IDs of the non-merge commits in a range, oldest first, which is the
// order rebase and cherry-pick apply them in. It returns nil if the range is empty or git fails.
func revList(repo *git.Repo, rng ...string) []string {
	out, err := repo.Output(append([]string{"rev-list", "--reverse", "--no-merges"}, append(rng, "--")...)...)
	if err != nil || out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// reportSim adds the result of simulate to r: no conflicts when at is empty, otherwise the
// commit where git would stop and its conflicting files.
func reportSim(r *Result, repo *git.Repo, at string, files []string) {
	if at == "" {
		r.add("No conflicts expected.")
		return
	}
	r.add("CONFLICT expected at %s.", subject(repo, at))
	listFiles(r, "Conflicting files", files)
	r.add("git will stop there for you to fix it (gitwhy resolve), or you can abort.")
}

// rebase previews "git rebase <upstream>": the commits replayed and, unless the rebase is
// interactive, whether replaying them onto upstream would conflict.
func rebase(repo *git.Repo, upstream string, interactive bool) Result {
	if !repo.OK("rev-parse", "--verify", "-q", upstream+"^{commit}") {
		return Result{}
	}
	r := Result{Exact: true}
	mine := revList(repo, upstream+"..HEAD")
	list, n := commits(repo, upstream+"..HEAD")
	if n == 0 {
		r.add("Your branch has no commits that %s doesn't; it just moves to %s.", upstream, short(repo, upstream))
		return r
	}
	listCommits(&r, "Your commits replayed on top of "+upstream+" (they get new IDs)", list, n)
	if interactive {
		return r // the todo list can change anything; conflicts can't be predicted
	}
	at, files, err := simulate(repo, upstream, mine, false)
	if err != nil {
		r.add("(Couldn't predict conflicts.)")
		return r
	}
	reportSim(&r, repo, at, files)
	return r
}

// cherryPick previews "git cherry-pick <revs>". Each rev may be one commit or an "a..b" range. It
// lists the commits copied and simulates applying them onto HEAD.
func cherryPick(repo *git.Repo, revs []string) Result {
	var list []string
	for _, rev := range revs {
		if strings.Contains(rev, "..") {
			list = append(list, revList(repo, rev)...)
		} else if id, err := repo.Output("rev-parse", "--verify", "-q", rev+"^{commit}"); err == nil {
			list = append(list, id)
		} else {
			return Result{}
		}
	}
	if len(list) == 0 {
		return Result{}
	}
	r := Result{Exact: true}
	r.add("Copies %d commit(s) onto your branch:", len(list))
	for i, c := range list {
		if i == 5 {
			r.add("    ... and %d more", len(list)-5)
			break
		}
		r.add("    %s", subject(repo, c))
	}
	at, files, err := simulate(repo, "HEAD", list, false)
	if err != nil {
		r.add("(Couldn't predict conflicts.)")
		return r
	}
	reportSim(&r, repo, at, files)
	return r
}

// revert previews "git revert <rev>" for one commit and simulates whether undoing it on top of
// HEAD would conflict.
func revert(repo *git.Repo, rev string) Result {
	id, err := repo.Output("rev-parse", "--verify", "-q", rev+"^{commit}")
	// A root commit has no parent to revert to, so it gets no preview.
	if err != nil || !repo.OK("rev-parse", "--verify", "-q", id+"^") {
		return Result{}
	}
	r := Result{Exact: true}
	r.add("A new commit reverses %s.", subject(repo, id))
	at, files, err := simulate(repo, "HEAD", []string{id}, true)
	if err != nil {
		return r
	}
	reportSim(&r, repo, at, files)
	return r
}

// branch previews deleting branches (-d, -D, --delete) and force-moving one (-f, --force). For a
// delete it lists the commits that no other branch, remote branch or tag still reaches.
func branch(repo *git.Repo, a danger.Args) Result {
	pos := a.Positional()
	r := Result{Exact: true}
	switch {
	case a.Has("-d", "-D", "--delete") && len(pos) > 0:
		for _, b := range pos {
			ref := "refs/heads/" + b
			// Names that are not local branches are skipped.
			if !repo.OK("show-ref", "--verify", "-q", ref) {
				continue
			}
			list, n := commits(repo, ref, "--not", "--exclude="+b, "--branches", "--remotes", "--tags") // --exclude is relative to refs/heads/
			if n == 0 {
				r.add("Branch %s is deleted. All its commits are also on other branches.", b)
			} else {
				r.add("Branch %s is deleted.", b)
				listCommits(&r, "Commits that will be on NO branch (recoverable for ~2 weeks with gitwhy recover)", list, n)
			}
		}
		return r
	case a.Has("-f", "--force") && len(pos) >= 1:
		// "branch -f <name> [<start>]": the start point defaults to HEAD.
		target := "HEAD"
		if len(pos) > 1 {
			target = pos[1]
		}
		b := "refs/heads/" + pos[0]
		if !repo.OK("show-ref", "--verify", "-q", b) {
			r.add("New branch %s at %s.", pos[0], subject(repo, target))
			return r
		}
		r.add("%s moves from %s to %s.", pos[0], short(repo, b), subject(repo, target))
		if list, n := commits(repo, target+".."+b); n > 0 {
			listCommits(&r, "Commits leaving "+pos[0], list, n)
		}
		return r
	}
	return Result{}
}

// stash previews "git stash" clear, drop, pop and push (also plain "git stash").
func stash(repo *git.Repo, a danger.Args) Result {
	r := Result{Exact: true}
	switch a.Sub {
	case "clear":
		out, _ := repo.Output("stash", "list")
		if out == "" {
			r.add("There are no stashes.")
			return r
		}
		listFiles(&r, "Stashes deleted", strings.Split(out, "\n"))
		return r
	case "drop", "pop":
		which := "stash@{0}"
		if p := a.Positional(); len(p) > 0 {
			which = p[0]
		}
		// The stash's reflog subject is its message, such as "WIP on main: abc1234 Fix typo".
		out, err := repo.Output("log", "-1", "-g", "--format=%gs", which)
		if err != nil {
			return Result{}
		}
		if a.Sub == "drop" {
			r.add("Deletes %s: %s", which, out)
		} else {
			r.add("Applies and removes %s: %s", which, out)
			files, _ := repo.Output("stash", "show", "--name-only", which)
			if files != "" {
				listFiles(&r, "Files it changes", strings.Split(files, "\n"))
			}
		}
		return r
	case "", "push":
		files := changedFiles(repo, "HEAD")
		if len(files) == 0 && !a.Has("-u", "--include-untracked", "-a", "--all") {
			r.add("No changes to tracked files to stash.")
			return r
		}
		listFiles(&r, "Files set aside (back to the last commit)", files)
		if a.Has("-u", "--include-untracked") {
			r.add("Untracked files are stashed too.")
		}
		return r
	}
	return Result{}
}

// switchBranch previews "git switch" and "git checkout". "checkout -- <files>" overwrites those
// files; otherwise it previews moving to one existing commit or branch. Creating a branch (-c, -b,
// --orphan and friends) gets no preview.
func switchBranch(repo *git.Repo, a danger.Args) Result {
	pos := a.Positional()
	// An empty "--" is not a file list, so "checkout <branch> --" is still a branch switch.
	if a.Cmd == "checkout" && len(a.AfterDash()) > 0 {
		r := Result{Exact: true}
		listFiles(&r, "Files overwritten", a.AfterDash())
		return r
	}
	if len(pos) != 1 || a.Has("-c", "-C", "-b", "-B", "--orphan") {
		return Result{}
	}
	target := pos[0]
	if !repo.OK("rev-parse", "--verify", "-q", target+"^{commit}") {
		// No local branch by that name: git creates one from a remote branch of the same name.
		// Only origin is checked here.
		if repo.OK("show-ref", "--verify", "-q", "refs/remotes/origin/"+target) {
			r := Result{Exact: true}
			r.add("Creates local branch %s tracking origin/%s, and switches to it.", target, target)
			return r
		}
		return Result{}
	}
	r := Result{Exact: true}
	r.add("Your files change to %s.", subject(repo, target))
	listFiles(&r, "Files that change", changedFiles(repo, "HEAD", target))
	if dirty := changedFiles(repo, "HEAD"); len(dirty) > 0 {
		if a.Has("-f", "--force", "--discard-changes") {
			listFiles(&r, "Uncommitted changes that will be DISCARDED", dirty)
		} else {
			r.add("Your %d uncommitted change(s) come along; git stops if one would be overwritten.", len(dirty))
		}
	}
	return r
}

// push previews "git push <remote> <refspec>" for a single refspec. It compares against the
// remote-tracking branch, so it is only as current as the last fetch and is never marked Exact.
func push(repo *git.Repo, a danger.Args) Result {
	pos := a.Positional()
	if len(pos) != 2 || a.Has("--delete", "-d", "--mirror", "--all", "--tags") {
		return Result{}
	}
	// A refspec is "[+]<src>[:<dst>]"; a leading "+" forces this one ref.
	remote, src := pos[0], strings.TrimPrefix(pos[1], "+")
	src, dst, ok := strings.Cut(src, ":")
	if !ok {
		dst = src
	}
	rref := "refs/remotes/" + remote + "/" + dst
	r := Result{}  // not Exact: the remote may have moved since the last fetch
	if src == "" { // "<remote> :<branch>" deletes the remote branch
		r.add("Deletes %s on %s.", dst, remote)
		return r
	}
	if !repo.OK("show-ref", "--verify", "-q", rref) {
		r.add("Creates %s on %s.", dst, remote)
		return r
	}
	if list, n := commits(repo, rref+".."+src); n > 0 {
		listCommits(&r, "Commits uploaded", list, n)
	} else {
		r.add("Nothing new to upload (as of your last fetch).")
	}
	if list, n := commits(repo, src+".."+rref); n > 0 {
		if a.Has("-f", "--force", "--force-with-lease") || strings.HasPrefix(pos[1], "+") {
			listCommits(&r, "Commits REMOVED from "+remote+"/"+dst+" (as of your last fetch)", list, n)
		} else {
			r.add("The remote has %d commit(s) you don't; the push will be rejected. Pull first.", n)
		}
	}
	r.add("(Based on your last fetch; the remote may have changed since.)")
	return r
}

// unique returns xs without empty strings and duplicates, keeping the first-seen order.
func unique(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if x != "" && !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
