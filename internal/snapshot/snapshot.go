// Package snapshot saves and restores repository state around gitwhy-run commands.
//
// A snapshot holds every ref and HEAD, a byte copy of the index, a commit of the whole working
// tree (tracked + untracked, not ignored), the state files of an in-progress merge/rebase/etc.,
// sparse-checkout settings, and the stash list. It lives in <gitdir>/gitwhy/snapshots/<id>/. The
// working-tree commit is kept alive by refs/gitwhy/snapshots/<id>. Its parents are every commit
// HEAD, a ref (branches, tags, remote-tracking branches) or a stash entry points to, with tags
// peeled to their commit. Other objects refs point to, such as annotated tag objects, are kept
// alive by refs/gitwhy/snapshots/<id>-<object ID>. So gc can't delete what undo needs.
package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Keep is how many snapshots are kept per worktree.
const Keep = 20

// refPrefix starts the refs that keep each snapshot's objects alive. The refs are shared by
// every worktree; the snapshot folders are not.
const refPrefix = "refs/gitwhy/snapshots/"

// stateFiles are the per-worktree files that describe an operation in progress.
var stateFiles = []string{
	"MERGE_HEAD", "MERGE_MSG", "MERGE_MODE", "MERGE_RR", "AUTO_MERGE", "ORIG_HEAD",
	"CHERRY_PICK_HEAD", "REVERT_HEAD", "REBASE_HEAD", "sequencer", "rebase-merge", "rebase-apply",
	"BISECT_LOG", "BISECT_START", "BISECT_TERMS", "BISECT_EXPECTED_REV", "BISECT_ANCESTORS_OK", "BISECT_NAMES",
	"BISECT_FIRST_PARENT", "BISECT_RUN",
}

// sparseFiles are per-worktree sparse-checkout settings (relative to the git dir).
var sparseFiles = []string{"info/sparse-checkout", "config.worktree"}

// sparseKeys are config keys that turn sparse checkout on.
var sparseKeys = []string{"core.sparseCheckout", "core.sparseCheckoutCone", "index.sparse"}

// Stash is one stash entry.
type Stash struct {
	Commit  string `json:"commit"`
	Message string `json:"message"` // the reflog subject, e.g. "On main: my wip"
}

// Snapshot describes one saved state.
type Snapshot struct {
	ID      string            `json:"id"` // base-36 creation time in nanoseconds; also the folder name
	Time    time.Time         `json:"time"`
	Label   string            `json:"label"`    // what was about to run, e.g. "git reset --hard HEAD~1"
	Head    string            `json:"head"`     // "ref: refs/heads/x" or a commit ID
	Refs    map[string]string `json:"refs"`     // refname -> object ID (no symrefs, refs/gitwhy, refs/stash)
	Work    string            `json:"work"`     // commit of the working tree
	Stashes []Stash           `json:"stashes"`  // newest first
	State   []string          `json:"state"`    // state files that existed
	HasIdx  bool              `json:"hasIndex"` // whether .git/index existed
	Sparse  map[string]string `json:"sparse"`   // sparse config key -> value ("" = unset)
	Files   []string          `json:"files"`    // sparseFiles that existed

	// Finished is true when Finish recorded the after-state (refs and stashes).
	Finished bool `json:"finished,omitempty"`

	// Refs right after the command ran (see Finish). Undo only reverts refs still at these values.
	After map[string]string `json:"after,omitempty"`
	// Stash list right after the command ran (see Finish). Undo keeps stashes made after that.
	StashesAfter []Stash `json:"stashesAfter,omitempty"`

	dir string // the snapshot's folder
}

// root returns the folder holding this worktree's snapshots.
func root(r *git.Repo) string { return filepath.Join(r.GitDir, "gitwhy", "snapshots") }

// Take saves the current state. extra lists ignored paths to include because the next command
// will delete or overwrite them.
func Take(r *git.Repo, label string, extra []string) (*Snapshot, error) {
	s, err := take(r, label, extra)
	if err == nil {
		prune(r, "")
	}
	return s, err
}

// take saves the current state like Take, without pruning old snapshots. If it fails, the
// snapshot's folder is removed.
func take(r *git.Repo, label string, extra []string) (*Snapshot, error) {
	if _, err := os.Stat(filepath.Join(r.GitDir, "index.lock")); err == nil {
		return nil, errors.New("another git process is running (.git/index.lock exists)")
	}
	now := time.Now()
	s := &Snapshot{ID: strconv.FormatInt(now.UnixNano(), 36), Time: now, Label: label, Sparse: map[string]string{}}
	s.dir = filepath.Join(root(r), s.ID)
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			os.RemoveAll(s.dir)
		}
	}()

	var err error
	if s.Head, err = readHead(r); err != nil {
		return nil, err
	}
	if s.Refs, err = refs(r); err != nil {
		return nil, err
	}
	if s.Stashes, err = stashes(r); err != nil {
		return nil, err
	}

	// Index: byte copy, so conflict stages and flags survive exactly.
	idx := filepath.Join(r.GitDir, "index")
	if _, err := os.Stat(idx); err == nil {
		s.HasIdx = true
		if err := copyFile(idx, filepath.Join(s.dir, "index")); err != nil {
			return nil, err
		}
	}

	// Working tree: a temp index seeded from the real one keeps add -A fast (stat cache).
	tmpIdx := filepath.Join(s.dir, "work-index")
	if s.HasIdx {
		if err := copyFile(idx, tmpIdx); err != nil {
			return nil, err
		}
	}
	w := r.With("GIT_INDEX_FILE=" + tmpIdx)
	if _, err := w.Output("add", "-A", "--ignore-errors"); err != nil { // no pathspec: the whole tree
		return nil, fmt.Errorf("saving working files: %w", err)
	}
	if len(extra) > 0 {
		in := strings.Join(extra, "\x00") + "\x00"
		if _, err := w.OutputIn(strings.NewReader(in), "add", "-f", "--ignore-errors", "--pathspec-from-file=-", "--pathspec-file-nul"); err != nil {
			return nil, fmt.Errorf("saving ignored files: %w", err)
		}
	}
	tree, err := w.Output("write-tree")
	if err != nil {
		return nil, fmt.Errorf("saving working files: %w", err)
	}
	os.Remove(tmpIdx)

	// Keep alive everything HEAD, the refs and the stash point to, so gc can't delete it: commits
	// (and the commits tags point to) become parents of the working-tree commit, and other
	// objects, such as annotated tags, get a ref of their own.
	args := []string{"commit-tree", tree, "-m", "gitwhy snapshot: " + label}
	var ids []string
	if !strings.HasPrefix(s.Head, "ref: ") {
		ids = append(ids, s.Head)
	}
	names := make([]string, 0, len(s.Refs))
	for n := range s.Refs {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		ids = append(ids, s.Refs[n])
	}
	for _, st := range s.Stashes {
		ids = append(ids, st.Commit)
	}
	commits, others := pins(r, ids)
	for _, id := range commits {
		args = append(args, "-p", id)
	}
	s.Work, err = r.With(commitEnv...).Output(args...)
	if err != nil {
		return nil, err
	}
	var tx strings.Builder
	fmt.Fprintf(&tx, "update %s %s\n", refPrefix+s.ID, s.Work)
	for _, id := range others {
		fmt.Fprintf(&tx, "update %s-%s %s\n", refPrefix+s.ID, id, id)
	}
	if _, err := r.OutputIn(strings.NewReader(tx.String()), "update-ref", "--stdin"); err != nil {
		return nil, err
	}

	// In-progress operation state and sparse-checkout settings.
	for _, f := range stateFiles {
		if copied, err := copyIfExists(filepath.Join(r.GitDir, f), filepath.Join(s.dir, "state", f)); err != nil {
			return nil, err
		} else if copied {
			s.State = append(s.State, f)
		}
	}
	for _, f := range sparseFiles {
		if copied, err := copyIfExists(filepath.Join(r.GitDir, f), filepath.Join(s.dir, "sparse", f)); err != nil {
			return nil, err
		} else if copied {
			s.Files = append(s.Files, f)
		}
	}
	// Only the shared config file: config.worktree is copied above, and writing the effective
	// value back to the shared file would turn sparse checkout on in every worktree.
	shared := filepath.Join(r.Common, "config")
	for _, k := range sparseKeys {
		s.Sparse[k], _ = r.Output("config", "--file", shared, "--get", k)
	}

	if err := s.save(); err != nil {
		return nil, err
	}
	ok = true
	return s, nil
}

// Finish records the refs right after the command ran, so undo can tell which refs changed later
// (by you, another worktree, or another tool) and leave those alone.
func Finish(r *git.Repo, s *Snapshot) error {
	after, err := refs(r)
	if err != nil {
		return err
	}
	if s.StashesAfter, err = stashes(r); err != nil {
		return err
	}
	s.After = after
	s.Finished = true
	return s.save()
}

// FinishedRecording reports whether Finish recorded the after-state (refs and stashes).
func (s *Snapshot) FinishedRecording() bool {
	return s.Finished || s.After != nil
}

// commitEnv gives commit-tree an identity even when the user has none configured.
var commitEnv = []string{
	"GIT_AUTHOR_NAME=gitwhy", "GIT_AUTHOR_EMAIL=gitwhy@localhost",
	"GIT_COMMITTER_NAME=gitwhy", "GIT_COMMITTER_EMAIL=gitwhy@localhost",
}

// save writes meta.json through a temporary file, so a crash never leaves half a file.
func (s *Snapshot) save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(s.dir, "meta.json.tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(s.dir, "meta.json"))
}

// List returns snapshots, newest first.
func List(r *git.Repo) ([]*Snapshot, error) {
	entries, err := os.ReadDir(root(r))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []*Snapshot
	for _, e := range entries {
		if s, err := read(filepath.Join(root(r), e.Name())); err == nil {
			out = append(out, s) // skip incomplete snapshots
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	return out, nil
}

// Get returns the snapshot with this ID.
func Get(r *git.Repo, id string) (*Snapshot, error) {
	if id == "" || id != filepath.Base(id) || strings.HasPrefix(id, ".") {
		return nil, fmt.Errorf("bad snapshot ID %q", id)
	}
	return read(filepath.Join(root(r), id))
}

// read loads the snapshot saved in dir.
func read(dir string) (*Snapshot, error) {
	data, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return nil, err
	}
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	s.dir = dir
	return &s, nil
}

// Delete removes a snapshot and the refs that keep its objects alive.
func Delete(r *git.Repo, s *Snapshot) {
	// Best effort: refs left behind lose their folder here, so the next prune deletes them.
	_ = deleteRefs(r, s.ID)
	os.RemoveAll(s.dir)
}

// deleteRefs removes refs/gitwhy/snapshots/<id> and the extra refs Take made for that ID.
func deleteRefs(r *git.Repo, id string) error {
	out, err := r.Output("for-each-ref", "--format=%(refname)", refPrefix)
	if err != nil {
		return err
	}
	var tx strings.Builder
	for _, name := range strings.Split(out, "\n") {
		if name == refPrefix+id || strings.HasPrefix(name, refPrefix+id+"-") {
			fmt.Fprintf(&tx, "delete %s\n", name)
		}
	}
	return updateRefs(r, tx.String())
}

// prune deletes all but the newest Keep snapshots, never the one with ID keep, then removes
// orphans (see pruneOrphans).
func prune(r *git.Repo, keep string) {
	all, err := List(r)
	if err != nil {
		return
	}
	n := 0
	for _, s := range all {
		if s.ID == keep {
			continue
		}
		if n++; n > Keep {
			Delete(r, s)
		}
	}
	pruneOrphans(r)
}

// staleAfter is how old a snapshot folder without meta.json must be before prune treats it as
// left over from an interrupted Take rather than one still being written.
const staleAfter = 10 * time.Minute

// pruneOrphans removes what an interrupted Take or a removed worktree leaves behind: this
// worktree's old snapshot folders without meta.json, and snapshot refs whose folder exists in no
// worktree (refs are shared by all worktrees, folders are not).
func pruneOrphans(r *git.Repo) {
	entries, _ := os.ReadDir(root(r))
	for _, e := range entries {
		dir := filepath.Join(root(r), e.Name())
		if _, err := os.Lstat(filepath.Join(dir, "meta.json")); !errors.Is(err, os.ErrNotExist) {
			continue
		}
		if fi, err := e.Info(); err == nil && time.Since(fi.ModTime()) > staleAfter {
			os.RemoveAll(dir)
		}
	}

	// Every worktree's git dir: the main one and <common>/worktrees/<name>. If any can't be
	// read, keep every ref rather than risk deleting one still in use.
	gitDirs := []string{r.Common, r.GitDir}
	wts, err := os.ReadDir(filepath.Join(r.Common, "worktrees"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}
	for _, w := range wts {
		gitDirs = append(gitDirs, filepath.Join(r.Common, "worktrees", w.Name()))
	}
	ids := map[string]bool{}
	for _, g := range gitDirs {
		es, err := os.ReadDir(filepath.Join(g, "gitwhy", "snapshots"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return
		}
		for _, e := range es {
			ids[e.Name()] = true
		}
	}
	out, err := r.Output("for-each-ref", "--format=%(refname)", refPrefix)
	if err != nil {
		return
	}
	var tx strings.Builder
	for _, name := range strings.Split(out, "\n") {
		id, _, _ := strings.Cut(strings.TrimPrefix(name, refPrefix), "-")
		if name != "" && !ids[id] {
			fmt.Fprintf(&tx, "delete %s\n", name)
		}
	}
	_ = updateRefs(r, tx.String()) // best effort: the next prune tries again
}

// Result is what a restore did.
type Result struct {
	Before  *Snapshot // the state just before the restore (undo it with Restore(Before))
	Skipped []string  // plain-English notes about what was left alone
	// HeadLeftAlone is true when HEAD, the index and the files were not restored because the
	// branch HEAD goes back to was left alone. The undo is then only partly done.
	HeadLeftAlone bool
}

// Restore puts the repository back to s. It snapshots the current state first, so the restore
// itself can be undone. Refs changed since the command ran, and stashes made since, are left
// alone and reported. If the branch HEAD goes back to is left alone, HEAD, the index and the
// files are too.
func Restore(r *git.Repo, s *Snapshot) (_ *Result, err error) {
	res := &Result{}

	// Check everything we need still exists before touching anything.
	if !r.OK("cat-file", "-e", s.Work+"^{tree}") {
		return nil, errors.New("this undo point's saved files are gone (git garbage-collected them)")
	}
	if !strings.HasPrefix(s.Head, "ref: ") && !r.OK("cat-file", "-e", s.Head) {
		return nil, errors.New("the commit HEAD was on is gone (history was rewritten and cleaned up)")
	}
	missing := missingObjects(r, s)

	cur, err := take(r, "before undo of: "+s.Label, nil)
	if err != nil {
		return nil, fmt.Errorf("saving current state first: %w", err)
	}
	// Files undo will overwrite that the current snapshot didn't save (ignored now): save them too.
	if extra := overwrittenUnsaved(r, s.Work, cur.Work); len(extra) > 0 {
		Delete(r, cur)
		if cur, err = take(r, "before undo of: "+s.Label, extra); err != nil {
			return nil, fmt.Errorf("saving current state first: %w", err)
		}
	}
	res.Before = cur
	// Record the refs this restore leaves, even if it fails partway, so a redo leaves alone
	// refs changed after it.
	defer func() {
		if ferr := Finish(r, cur); ferr != nil {
			err = errors.Join(err, ferr)
		}
	}()

	// 1. Refs: deletes, then updates, each one transaction. If the updates fail, the deleted refs
	// are put back, so a failure here changes nothing.
	elsewhere := checkedOutElsewhere(r)
	var dels, ups, undel strings.Builder
	skipped := map[string]bool{}
	names := map[string]bool{}
	for n := range s.Refs {
		names[n] = true
	}
	for n := range cur.Refs {
		names[n] = true
	}
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)
	for _, name := range sorted {
		want, had := s.Refs[name]
		now, has := cur.Refs[name]
		if want == now && had == has {
			continue
		}
		short := shortRef(name)
		if s.FinishedRecording() {
			afterRefs := s.After
			if afterRefs == nil {
				afterRefs = map[string]string{}
			}
			if after, ok := afterRefs[name]; after != now || ok != has {
				res.Skipped = append(res.Skipped, short+" changed after this action, so it was left as is")
				skipped[name] = true
				continue
			}
		}
		if elsewhere[name] {
			res.Skipped = append(res.Skipped, short+" is checked out in another worktree, so it was left as is")
			skipped[name] = true
			continue
		}
		if had && missing[want] {
			res.Skipped = append(res.Skipped, short+" can't go back: its old commits are gone")
			skipped[name] = true
			continue
		}
		switch {
		case !had:
			// Deleted first, in its own transaction: "foo" can't be written while "foo/bar" exists.
			fmt.Fprintf(&dels, "delete %s %s\n", name, now)
			fmt.Fprintf(&undel, "create %s %s\n", name, now)
		case has:
			fmt.Fprintf(&ups, "update %s %s %s\n", name, want, now) // update keeps the reflog
		default:
			fmt.Fprintf(&ups, "create %s %s\n", name, want)
		}
	}
	if err := updateRefs(r, dels.String()); err != nil {
		return res, fmt.Errorf("restoring branches (nothing was changed): %w", err)
	}
	if err := updateRefs(r, ups.String()); err != nil {
		if uerr := updateRefs(r, undel.String()); uerr != nil {
			return res, fmt.Errorf("restoring branches failed and some branches were deleted: %w", errors.Join(err, uerr))
		}
		return res, fmt.Errorf("restoring branches (nothing was changed): %w", err)
	}

	// 2-5. HEAD, files, index and operation state, unless the branch HEAD goes back to was left
	// alone: old files under a newer commit, or a branch checked out twice, would be wrong.
	if ref, ok := strings.CutPrefix(s.Head, "ref: "); ok && (skipped[ref] || elsewhere[ref]) {
		res.HeadLeftAlone = true
		res.Skipped = append(res.Skipped, "HEAD, the index and your files were left as they are, because "+shortRef(ref)+" was left as is")
	} else if err := restoreCheckout(r, s, cur); err != nil {
		return res, fmt.Errorf("%w\nThe undo is only partly done. Run \"gitwhy undo\" again to go back to how things were before it", err)
	}

	// 6. Stash list. Stashes made after the command ran stay on top of the restored list.
	var later []Stash
	if s.FinishedRecording() {
		known := map[string]bool{}
		for _, st := range append(append([]Stash{}, s.Stashes...), s.StashesAfter...) {
			known[st.Commit] = true
		}
		for _, st := range cur.Stashes {
			if !known[st.Commit] {
				later = append(later, st)
				res.Skipped = append(res.Skipped, "a stash made after this action ("+st.Message+") was kept")
			}
		}
	}
	if want := append(append([]Stash{}, later...), s.Stashes...); !sameStashes(want, cur.Stashes) {
		r.Output("update-ref", "-d", "refs/stash")
		for i := len(want) - 1; i >= 0; i-- {
			st := want[i]
			if missing[st.Commit] {
				res.Skipped = append(res.Skipped, "a stash ("+st.Message+") can't come back: it's gone")
				continue
			}
			if _, err := r.Output("stash", "store", "-m", st.Message, st.Commit); err != nil {
				return res, fmt.Errorf("restoring stash: %w", err)
			}
		}
	}
	prune(r, cur.ID)
	return res, nil
}

// restoreCheckout puts back HEAD, the working files, sparse-checkout settings, the index and the
// state of an operation in progress from s. cur is the snapshot of the current state.
func restoreCheckout(r *git.Repo, s, cur *Snapshot) error {
	// 2. HEAD.
	var err error
	if ref, ok := strings.CutPrefix(s.Head, "ref: "); ok {
		_, err = r.Output("symbolic-ref", "HEAD", ref)
	} else {
		_, err = r.Output("update-ref", "--no-deref", "HEAD", s.Head)
	}
	if err != nil {
		return fmt.Errorf("restoring HEAD: %w", err)
	}

	// 3. Working files: delete files added since, rewrite files that differ. A file that can't be
	// changed doesn't stop the rest: the index and state are still put back, then it's reported.
	fileErr := restoreFiles(r, s.Work, cur.Work)

	// 4. Sparse-checkout settings, then the index.
	for _, f := range sparseFiles {
		os.Remove(filepath.Join(r.GitDir, f))
	}
	for _, f := range s.Files {
		if err := copyFile(filepath.Join(s.dir, "sparse", f), filepath.Join(r.GitDir, f)); err != nil {
			return err
		}
	}
	shared := filepath.Join(r.Common, "config")
	for _, k := range sparseKeys {
		if v := s.Sparse[k]; v != "" {
			if _, err := r.Output("config", "--file", shared, k, v); err != nil {
				return fmt.Errorf("restoring %s: %w", k, err)
			}
		} else if r.OK("config", "--file", shared, "--get", k) {
			if _, err := r.Output("config", "--file", shared, "--unset", k); err != nil {
				return fmt.Errorf("restoring %s: %w", k, err)
			}
		}
	}
	idx := filepath.Join(r.GitDir, "index")
	if s.HasIdx {
		if err := copyFile(filepath.Join(s.dir, "index"), idx+".gitwhy-tmp"); err != nil {
			return err
		}
		if err := os.Rename(idx+".gitwhy-tmp", idx); err != nil {
			return err
		}
		r.Output("update-index", "-q", "--refresh") // refresh stat info; dirty entries are expected
	} else {
		os.Remove(idx)
	}
	if err := restoreMissing(r, s); err != nil {
		return err
	}

	// 5. Operation state.
	for _, f := range stateFiles {
		if err := os.RemoveAll(filepath.Join(r.GitDir, f)); err != nil {
			return err
		}
	}
	for _, f := range s.State {
		if err := copyAll(filepath.Join(s.dir, "state", f), filepath.Join(r.GitDir, f)); err != nil {
			return err
		}
	}
	return fileErr
}

// updateRefs runs one all-or-nothing "git update-ref --stdin" transaction, if tx isn't empty.
func updateRefs(r *git.Repo, tx string) error {
	if tx == "" {
		return nil
	}
	_, err := r.OutputIn(strings.NewReader(tx), "update-ref", "--no-deref", "--stdin")
	return err
}

// shortRef turns a ref name into the name people use: "main", "origin/main" or "tag v1".
func shortRef(name string) string {
	for _, p := range []string{"refs/heads/", "refs/remotes/", "refs/tags/"} {
		if s, ok := strings.CutPrefix(name, p); ok {
			if p == "refs/tags/" {
				return "tag " + s
			}
			return s
		}
	}
	return name
}

// missingObjects returns object IDs from s that no longer exist.
func missingObjects(r *git.Repo, s *Snapshot) map[string]bool {
	var ids []string
	for _, id := range s.Refs {
		ids = append(ids, id)
	}
	for _, st := range s.Stashes {
		ids = append(ids, st.Commit)
	}
	out, _ := r.OutputIn(strings.NewReader(strings.Join(ids, "\n")+"\n"), "cat-file", "--batch-check")
	m := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		if id, ok := strings.CutSuffix(line, " missing"); ok {
			m[id] = true
		}
	}
	return m
}

// pins sorts objects to keep alive, with one git process: commits (including the commit each tag
// points to) can be parents of the working-tree commit; other objects, such as annotated tags,
// need a ref of their own. Missing objects and repeats are left out.
func pins(r *git.Repo, ids []string) (commits, others []string) {
	var in strings.Builder
	var uniq []string
	seen := map[string]bool{}
	for _, id := range ids {
		if id != "" && !seen[id] {
			seen[id] = true
			uniq = append(uniq, id)
			fmt.Fprintf(&in, "%s\n%s^{commit}\n", id, id)
		}
	}
	if len(uniq) == 0 {
		return nil, nil
	}
	out, _ := r.OutputIn(strings.NewReader(in.String()), "cat-file", "--batch-check=%(objectname) %(objecttype)")
	lines := strings.Split(out, "\n")
	parent := map[string]bool{}
	for i, id := range uniq {
		if 2*i+1 >= len(lines) {
			break
		}
		if typ := strings.TrimPrefix(lines[2*i], id+" "); typ != "missing" && typ != "commit" {
			others = append(others, id)
		}
		if c, typ, _ := strings.Cut(lines[2*i+1], " "); typ == "commit" && !parent[c] {
			parent[c] = true
			commits = append(commits, c)
		}
	}
	return commits, others
}

// checkedOutElsewhere lists branches checked out in other worktrees.
func checkedOutElsewhere(r *git.Repo) map[string]bool {
	out, _ := r.Output("worktree", "list", "--porcelain")
	m := map[string]bool{}
	var path string
	for _, line := range strings.Split(out, "\n") {
		if p, ok := strings.CutPrefix(line, "worktree "); ok {
			path = p
		}
		if b, ok := strings.CutPrefix(line, "branch "); ok && filepath.Clean(path) != filepath.Clean(r.Top) {
			m[b] = true
		}
	}
	return m
}

// change is one entry of "git diff-tree -r --raw".
type change struct {
	oldMode, newMode string // file modes, such as "100644", or gitlink
	status           byte   // 'A' added, 'D' deleted, 'M' changed, 'T' type changed (from a to b)
	path             string // relative to the repository root, with forward slashes
}

// diffRaw lists the files that differ between the trees of commits a and b.
func diffRaw(r *git.Repo, a, b string) ([]change, error) {
	out, err := r.Output("diff-tree", "-r", "-z", "--no-renames", a, b)
	if err != nil {
		return nil, err
	}
	f := strings.Split(out, "\x00")
	var cs []change
	for i := 0; i+1 < len(f); i += 2 {
		meta := strings.Fields(strings.TrimPrefix(f[i], ":"))
		if len(meta) < 5 {
			continue
		}
		cs = append(cs, change{oldMode: meta[0], newMode: meta[1], status: meta[4][0], path: f[i+1]})
	}
	return cs, nil
}

// gitlink is the file mode of a submodule or nested repository: never touched by restore.
const gitlink = "160000"

// isGitlink reports whether a change is to a submodule or nested repository.
func isGitlink(c change) bool { return c.oldMode == gitlink || c.newMode == gitlink }

// hasRepo reports whether a path is or contains a nested repository.
func hasRepo(full string) bool {
	found := false
	filepath.WalkDir(full, func(p string, d os.DirEntry, err error) error {
		if err == nil && d.Name() == ".git" {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// overwrittenUnsaved lists files restore will write over or delete that exist on disk but aren't
// in the current snapshot (because they're ignored now): a file at a path the snapshot needs,
// ignored files inside a folder at such a path, and a file where one of its folders must go.
func overwrittenUnsaved(r *git.Repo, want, have string) []string {
	cs, err := diffRaw(r, want, have)
	if err != nil {
		return nil
	}
	var out, dirs []string
	seen := map[string]bool{}
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for _, c := range cs {
		if c.status != 'D' || isGitlink(c) {
			continue
		}
		if fi, err := os.Lstat(filepath.Join(r.Top, filepath.FromSlash(c.path))); err == nil {
			if fi.IsDir() {
				dirs = append(dirs, c.path)
			} else {
				add(c.path)
			}
		}
		for d := path.Dir(c.path); d != "."; d = path.Dir(d) {
			if fi, err := os.Lstat(filepath.Join(r.Top, filepath.FromSlash(d))); err == nil && !fi.IsDir() {
				add(d)
			}
		}
	}
	if len(dirs) > 0 {
		top := r.Runner
		top.Dir = r.Top // ls-files prints paths relative to the current folder
		list, _ := top.Output(append([]string{"ls-files", "-z", "--others", "--ignored", "--exclude-standard", "--"}, dirs...)...)
		for _, p := range strings.Split(list, "\x00") {
			if p != "" {
				add(p)
			}
		}
	}
	return out
}

// restoreFiles makes the working files match commit want, given that commit have holds them as
// they are now: files added since are deleted, the others are written from want. Nested
// repositories are never touched.
func restoreFiles(r *git.Repo, want, have string) error {
	cs, err := diffRaw(r, want, have)
	if err != nil {
		return err
	}
	// Errors are collected, not returned at once: stopping here would leave files deleted
	// below and never written back.
	var errs []error
	var write []string
	for _, c := range cs {
		if isGitlink(c) {
			continue
		}
		full := filepath.Join(r.Top, filepath.FromSlash(c.path))
		switch c.status {
		case 'A': // added since: delete (files only; never a folder holding a repo)
			fi, err := os.Lstat(full)
			if err != nil || fi.IsDir() {
				continue
			}
			if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
				errs = append(errs, err)
				continue
			}
			removeEmptyParents(r.Top, filepath.Dir(full))
		default: // changed, deleted or type changed: rewrite from the snapshot
			if fi, err := os.Lstat(full); err == nil && fi.IsDir() {
				if hasRepo(full) {
					continue
				}
				os.RemoveAll(full)
			} else {
				os.Remove(full)
			}
			// A parent that became a file must go before the file can be written.
			for d := filepath.Dir(full); d != r.Top && strings.HasPrefix(d, r.Top); d = filepath.Dir(d) {
				if fi, err := os.Lstat(d); err == nil && !fi.IsDir() {
					os.Remove(d)
				}
			}
			write = append(write, c.path)
		}
	}
	return errors.Join(append(errs, checkout(r, want, write))...)
}

// checkout writes paths from a commit's tree to the working tree, leaving the index alone.
func checkout(r *git.Repo, commit string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	tmp := filepath.Join(r.GitDir, "gitwhy", "restore-index")
	defer os.Remove(tmp)
	w := r.With("GIT_INDEX_FILE=" + tmp)
	if _, err := w.Output("read-tree", commit); err != nil {
		return err
	}
	in := strings.Join(paths, "\x00") + "\x00"
	if _, err := w.OutputIn(strings.NewReader(in), "checkout-index", "-f", "-z", "--stdin"); err != nil {
		return fmt.Errorf("restoring files: %w", err)
	}
	return nil
}

// restoreMissing brings back files the snapshot had that aren't on disk (for example after a
// sparse checkout hid them), or reapplies sparse checkout if it's on.
func restoreMissing(r *git.Repo, s *Snapshot) error {
	// The settings are already restored, so this reads the value s had, from either file.
	if v, _ := r.Output("config", "--bool", "--get", "core.sparseCheckout"); v == "true" {
		r.Output("sparse-checkout", "reapply")
		return nil
	}
	top := r.Runner
	top.Dir = r.Top // ls-files lists the current folder; this lists the whole tree
	out, err := top.Output("ls-files", "-z", "--deleted")
	if err != nil || out == "" {
		return nil
	}
	saved, err := r.Output("ls-tree", "-r", "-z", "--name-only", s.Work)
	if err != nil {
		return err
	}
	inWork := map[string]bool{}
	for _, p := range strings.Split(saved, "\x00") {
		inWork[p] = true
	}
	var paths []string
	for _, p := range strings.Split(out, "\x00") {
		if p != "" && inWork[p] {
			paths = append(paths, p)
		}
	}
	return checkout(r, s.Work, paths)
}

// readHead returns "ref: <branch ref>", or the commit ID if HEAD is detached.
func readHead(r *git.Repo) (string, error) {
	if ref, err := r.Output("symbolic-ref", "-q", "HEAD"); err == nil {
		return "ref: " + ref, nil
	}
	return r.Output("rev-parse", "HEAD")
}

// refs returns every real ref; symbolic refs (like origin/HEAD) are skipped, since writing them
// would move the ref they point to.
func refs(r *git.Repo) (map[string]string, error) {
	out, err := r.Output("for-each-ref", "--format=%(objectname) %(refname) %(symref)")
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) != 2 || strings.HasPrefix(f[1], "refs/gitwhy/") || f[1] == "refs/stash" {
			continue
		}
		m[f[1]] = f[0]
	}
	return m, nil
}

// stashes returns the stash list, newest first.
func stashes(r *git.Repo) ([]Stash, error) {
	if !r.OK("rev-parse", "--verify", "-q", "refs/stash") {
		return nil, nil
	}
	out, err := r.Output("log", "-g", "--format=%H%x00%gs", "refs/stash")
	if err != nil {
		return nil, err
	}
	var st []Stash
	for _, line := range strings.Split(out, "\n") {
		if id, msg, ok := strings.Cut(line, "\x00"); ok {
			st = append(st, Stash{Commit: id, Message: msg})
		}
	}
	return st, nil
}

// sameStashes reports whether two stash lists hold the same commits in the same order.
func sameStashes(a, b []Stash) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Commit != b[i].Commit {
			return false
		}
	}
	return true
}

// removeEmptyParents removes dir and its parents while they are empty, stopping at top.
func removeEmptyParents(top, dir string) {
	for dir != top && strings.HasPrefix(dir, top) {
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

// copyIfExists copies the file or folder src to dst if src exists, and reports whether it did.
func copyIfExists(src, dst string) (bool, error) {
	if _, err := os.Lstat(src); err != nil {
		return false, nil
	}
	return true, copyAll(src, dst)
}

// copyFile copies the file src to dst, creating dst's folder if needed.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// copyAll copies the file or folder src to dst, with everything inside it.
func copyAll(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		return copyFile(p, target)
	})
}
