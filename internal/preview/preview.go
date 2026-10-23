// Package preview says what a git command will do before it runs.
//
// Exact previews use git's own dry runs and plumbing (merge-tree, rev-list, diff) without touching
// the repository. Other commands get no exact preview; callers fall back to recipe text.
package preview

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dpologdvinity/gitwhy/internal/danger"
	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Result of previewing one command.
type Result struct {
	Exact bool     // computed, not guessed
	Lines []string // what will happen
	Save  []string // ignored paths the command will delete; include them in the snapshot
}

func (r *Result) add(format string, a ...any) { r.Lines = append(r.Lines, fmt.Sprintf(format, a...)) }

// Preview computes what args (without the leading "git") will do in repo.
func Preview(repo *git.Repo, args []string) Result {
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
		return dryRun(repo, args, "-n", "", "will move", false)
	case "merge":
		if a.Has("--abort", "--continue", "--quit") {
			return Result{}
		}
		if p := a.Positional(); len(p) == 1 {
			return merge(repo, p[0], a.Has("--squash"))
		}
	case "rebase":
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

func short(repo *git.Repo, rev string) string {
	out, err := repo.Output("rev-parse", "--short", rev)
	if err != nil {
		return rev
	}
	return out
}

func subject(repo *git.Repo, rev string) string {
	out, err := repo.Output("log", "-1", "--format=%h \"%s\"", rev, "--")
	if err != nil {
		return rev
	}
	return out
}

// commits lists up to max one-line commits in a range, plus the total count.
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

func listCommits(r *Result, title string, list []string, n int) {
	r.add("%s (%d):", title, n)
	for _, c := range list {
		r.add("    %s", c)
	}
	if n > len(list) {
		r.add("    ... and %d more", n-len(list))
	}
}

func changedFiles(repo *git.Repo, args ...string) []string {
	out, err := repo.Output(append([]string{"diff", "--name-only"}, args...)...)
	if err != nil || out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

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

func reset(repo *git.Repo, a danger.Args) Result {
	r := Result{Exact: true}
	pos := a.Positional()
	paths := a.AfterDash()
	target := "HEAD"
	if len(pos) > 0 {
		target = pos[0]
		paths = append(paths, pos[1:]...)
	}
	if len(paths) > 0 || a.Has("-p", "--patch") {
		r.add("Unstages changes in the given files (sets their staged version to %s). Your files on disk don't change.", target)
		return r
	}
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

// dryRun runs the command itself with a dry-run flag and reports its output.
func dryRun(repo *git.Repo, args []string, flag, prefix, verb string, isClean bool) Result {
	var dry []string
	for i, w := range args {
		if i > 0 && isClean && strings.HasPrefix(w, "-") && !strings.HasPrefix(w, "--") {
			w = strings.ReplaceAll(w, "f", "") // drop -f: -n must win
			if w == "-" {
				continue
			}
		}
		if isClean && (w == "--force") {
			continue
		}
		dry = append(dry, w)
	}
	dry = append(dry[:1], append([]string{flag}, dry[1:]...)...)
	out, err := repo.Output(dry...)
	if err != nil {
		return Result{}
	}
	r := Result{Exact: true}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		f := strings.TrimPrefix(line, prefix)
		f = strings.TrimSuffix(f, "'")
		files = append(files, f)
	}
	if len(files) == 0 {
		r.add("Nothing to do.")
		return r
	}
	listFiles(&r, "Files git "+verb, files)
	if isClean {
		r.Save = files
	}
	return r
}

func mergeTree(repo *git.Repo, args ...string) (tree string, conflicts []string, err error) {
	out, err := repo.Output(append([]string{"merge-tree", "--write-tree", "--name-only", "--no-messages"}, args...)...)
	lines := strings.Split(out, "\n")
	if err != nil {
		if git.ExitCode(err) != 1 || len(lines) == 0 {
			return "", nil, err
		}
		return lines[0], lines[1:], nil
	}
	return lines[0], nil, nil
}

func merge(repo *git.Repo, from string, squash bool) Result {
	if !repo.OK("rev-parse", "--verify", "-q", from+"^{commit}") {
		return Result{}
	}
	r := Result{Exact: true}
	if repo.OK("merge-base", "--is-ancestor", from, "HEAD") {
		r.add("Already up to date: everything on %s is already here. Nothing changes.", from)
		return r
	}
	list, n := commits(repo, "HEAD.."+from)
	if repo.OK("merge-base", "--is-ancestor", "HEAD", from) && !squash {
		r.add("Fast-forward: your branch just moves ahead to %s. No merge commit.", short(repo, from))
		listCommits(&r, "Commits coming in", list, n)
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
	} else if squash {
		r.add("No conflicts. The changes are staged; you then make one commit.")
	} else {
		r.add("No conflicts. A merge commit is created.")
	}
	return r
}

// simulate replays commits (oldest first) onto base with merge-tree, without touching the repo.
// It returns the commit that would conflict and the conflicting files, if any.
func simulate(repo *git.Repo, base string, list []string, revert bool) (string, []string, error) {
	cur := base
	for _, c := range list {
		var tree string
		var conflicts []string
		var err error
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

func revList(repo *git.Repo, rng ...string) []string {
	out, err := repo.Output(append([]string{"rev-list", "--reverse", "--no-merges"}, append(rng, "--")...)...)
	if err != nil || out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

func reportSim(r *Result, repo *git.Repo, at string, files []string) {
	if at == "" {
		r.add("No conflicts expected.")
		return
	}
	r.add("CONFLICT expected at %s.", subject(repo, at))
	listFiles(r, "Conflicting files", files)
	r.add("git will stop there for you to fix it (gitwhy resolve), or you can abort.")
}

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

func revert(repo *git.Repo, rev string) Result {
	id, err := repo.Output("rev-parse", "--verify", "-q", rev+"^{commit}")
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

func branch(repo *git.Repo, a danger.Args) Result {
	pos := a.Positional()
	r := Result{Exact: true}
	switch {
	case a.Has("-d", "-D", "--delete") && len(pos) > 0:
		for _, b := range pos {
			ref := "refs/heads/" + b
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

func switchBranch(repo *git.Repo, a danger.Args) Result {
	pos := a.Positional()
	if a.Cmd == "checkout" && a.AfterDash() != nil {
		r := Result{Exact: true}
		listFiles(&r, "Files overwritten", a.AfterDash())
		return r
	}
	if len(pos) != 1 || a.Has("-c", "-C", "-b", "-B", "--orphan") {
		return Result{}
	}
	target := pos[0]
	if !repo.OK("rev-parse", "--verify", "-q", target+"^{commit}") {
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
		r.add("Your %d uncommitted change(s) come along; git stops if one would be overwritten.", len(dirty))
	}
	return r
}

func push(repo *git.Repo, a danger.Args) Result {
	pos := a.Positional()
	if len(pos) != 2 || a.Has("--delete", "-d", "--mirror", "--all", "--tags") {
		return Result{}
	}
	remote, src := pos[0], strings.TrimPrefix(pos[1], "+")
	src, dst, ok := strings.Cut(src, ":")
	if !ok {
		dst = src
	}
	rref := "refs/remotes/" + remote + "/" + dst
	r := Result{}
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
