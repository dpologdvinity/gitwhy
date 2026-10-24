// Package snapshot saves and restores repository state around gitwhy-run commands.
//
// A snapshot holds every ref and HEAD, a byte copy of the index, a commit of the whole working
// tree (tracked + untracked, not ignored), the state files of an in-progress merge/rebase/etc.,
// sparse-checkout settings, and the stash list. It lives in <gitdir>/gitwhy/snapshots/<id>/. The
// working-tree commit, whose parents are every local branch tip, HEAD and stash entry, is kept
// alive by refs/gitwhy/snapshots/<id>, so gc can't delete what undo needs.
package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Keep is how many snapshots are kept per worktree.
const Keep = 20

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
	Message string `json:"message"`
}

// Snapshot describes one saved state.
type Snapshot struct {
	ID      string            `json:"id"`
	Time    time.Time         `json:"time"`
	Label   string            `json:"label"` // what was about to run, e.g. "git reset --hard HEAD~1"
	Head    string            `json:"head"`  // "ref: refs/heads/x" or a commit ID
	Refs    map[string]string `json:"refs"`  // refname -> object ID (no symrefs, refs/gitwhy, refs/stash)
	Work    string            `json:"work"`  // commit of the working tree
	Stashes []Stash           `json:"stashes"`
	State   []string          `json:"state"`    // state files that existed
	HasIdx  bool              `json:"hasIndex"` // whether .git/index existed
	Sparse  map[string]string `json:"sparse"`   // sparse config key -> value ("" = unset)
	Files   []string          `json:"files"`    // sparseFiles that existed

	// Refs right after the command ran (see Finish). Undo only reverts refs still at these values.
	After map[string]string `json:"after,omitempty"`

	dir string
}

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

	// Parents pin every local branch tip, HEAD and stash commit, so gc can't delete them.
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
		if strings.HasPrefix(n, "refs/heads/") {
			ids = append(ids, s.Refs[n])
		}
	}
	for _, st := range s.Stashes {
		ids = append(ids, st.Commit)
	}
	types := objectTypes(r, ids)
	seen := map[string]bool{}
	for _, id := range ids {
		if !seen[id] && types[id] == "commit" {
			seen[id] = true
			args = append(args, "-p", id)
		}
	}
	s.Work, err = r.With(commitEnv...).Output(args...)
	if err != nil {
		return nil, err
	}
	if _, err := r.Output("update-ref", refPrefix+s.ID, s.Work); err != nil {
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
	for _, k := range sparseKeys {
		s.Sparse[k], _ = r.Output("config", "--get", k)
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
	s.After = after
	return s.save()
}

// commitEnv gives commit-tree an identity even when the user has none configured.
var commitEnv = []string{
	"GIT_AUTHOR_NAME=gitwhy", "GIT_AUTHOR_EMAIL=gitwhy@localhost",
	"GIT_COMMITTER_NAME=gitwhy", "GIT_COMMITTER_EMAIL=gitwhy@localhost",
}

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

// Delete removes a snapshot and its ref.
func Delete(r *git.Repo, s *Snapshot) {
	r.Output("update-ref", "-d", refPrefix+s.ID)
	os.RemoveAll(s.dir)
}

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
}

// Result of a restore.
type Result struct {
	Before  *Snapshot // the state just before the restore (undo it with Restore(Before))
	Skipped []string  // plain-English notes about what was left alone
}

// Restore puts the repository back to s. It snapshots the current state first, so the restore
// itself can be undone. Refs changed since the command ran are left alone and reported.
func Restore(r *git.Repo, s *Snapshot) (*Result, error) {
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

	// 1. Refs: one transaction, so a failure here changes nothing.
	elsewhere := checkedOutElsewhere(r)
	var tx strings.Builder
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
		if s.After != nil {
			if after, ok := s.After[name]; after != now || ok != has {
				res.Skipped = append(res.Skipped, short+" changed after this action, so it was left as is")
				continue
			}
		}
		if elsewhere[name] {
			res.Skipped = append(res.Skipped, short+" is checked out in another worktree, so it was left as is")
			continue
		}
		if had && missing[want] {
			res.Skipped = append(res.Skipped, short+" can't go back: its old commits are gone")
			continue
		}
		if had {
			fmt.Fprintf(&tx, "update %s %s\n", name, want)
		} else {
			fmt.Fprintf(&tx, "delete %s\n", name)
		}
	}
	if tx.Len() > 0 {
		if _, err := r.OutputIn(strings.NewReader(tx.String()), "update-ref", "--no-deref", "--stdin"); err != nil {
			return res, fmt.Errorf("restoring branches (nothing was changed): %w", err)
		}
	}

	// 2. HEAD.
	if ref, ok := strings.CutPrefix(s.Head, "ref: "); ok {
		_, err = r.Output("symbolic-ref", "HEAD", ref)
	} else {
		_, err = r.Output("update-ref", "--no-deref", "HEAD", s.Head)
	}
	if err != nil {
		return res, fmt.Errorf("restoring HEAD: %w", err)
	}

	// 3. Working files: delete files added since, rewrite files that differ.
	if err := restoreFiles(r, s.Work, cur.Work); err != nil {
		return res, err
	}

	// 4. Sparse-checkout settings, then the index.
	for _, f := range sparseFiles {
		os.Remove(filepath.Join(r.GitDir, f))
	}
	for _, f := range s.Files {
		if err := copyFile(filepath.Join(s.dir, "sparse", f), filepath.Join(r.GitDir, f)); err != nil {
			return res, err
		}
	}
	for _, k := range sparseKeys {
		if v := s.Sparse[k]; v != "" {
			if _, err := r.Output("config", k, v); err != nil {
				return res, fmt.Errorf("restoring %s: %w", k, err)
			}
		} else if r.OK("config", "--get", k) {
			if _, err := r.Output("config", "--unset", k); err != nil {
				return res, fmt.Errorf("restoring %s: %w", k, err)
			}
		}
	}
	idx := filepath.Join(r.GitDir, "index")
	if s.HasIdx {
		if err := copyFile(filepath.Join(s.dir, "index"), idx+".gitwhy-tmp"); err != nil {
			return res, err
		}
		if err := os.Rename(idx+".gitwhy-tmp", idx); err != nil {
			return res, err
		}
		r.Output("update-index", "-q", "--refresh") // refresh stat info; dirty entries are expected
	} else {
		os.Remove(idx)
	}
	if err := restoreMissing(r, s); err != nil {
		return res, err
	}

	// 5. Operation state.
	for _, f := range stateFiles {
		if err := os.RemoveAll(filepath.Join(r.GitDir, f)); err != nil {
			return res, err
		}
	}
	for _, f := range s.State {
		if err := copyAll(filepath.Join(s.dir, "state", f), filepath.Join(r.GitDir, f)); err != nil {
			return res, err
		}
	}

	// 6. Stash list.
	if !sameStashes(s.Stashes, cur.Stashes) {
		r.Output("update-ref", "-d", "refs/stash")
		for i := len(s.Stashes) - 1; i >= 0; i-- {
			st := s.Stashes[i]
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

// objectTypes returns the type of each object ("commit", "tag", ...) with one git process.
// Missing objects are left out.
func objectTypes(r *git.Repo, ids []string) map[string]string {
	m := map[string]string{}
	if len(ids) == 0 {
		return m
	}
	out, _ := r.OutputIn(strings.NewReader(strings.Join(ids, "\n")+"\n"), "cat-file", "--batch-check=%(objectname) %(objecttype)")
	for _, line := range strings.Split(out, "\n") {
		if id, typ, ok := strings.Cut(line, " "); ok && typ != "missing" {
			m[id] = typ
		}
	}
	return m
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
	oldMode, newMode string
	status           byte
	path             string
}

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

const gitlink = "160000" // a submodule or nested repository: never touched by restore

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

// overwrittenUnsaved lists files restore will write over that exist on disk but aren't in the
// current snapshot (because they're ignored now).
func overwrittenUnsaved(r *git.Repo, want, have string) []string {
	cs, err := diffRaw(r, want, have)
	if err != nil {
		return nil
	}
	var out []string
	for _, c := range cs {
		if c.status == 'D' && !isGitlink(c) {
			if fi, err := os.Lstat(filepath.Join(r.Top, filepath.FromSlash(c.path))); err == nil && !fi.IsDir() {
				out = append(out, c.path)
			}
		}
	}
	return out
}

func restoreFiles(r *git.Repo, want, have string) error {
	cs, err := diffRaw(r, want, have)
	if err != nil {
		return err
	}
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
				return err
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
	return checkout(r, want, write)
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
	if v := s.Sparse["core.sparseCheckout"]; v == "true" {
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

func removeEmptyParents(top, dir string) {
	for dir != top && strings.HasPrefix(dir, top) {
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func copyIfExists(src, dst string) (bool, error) {
	if _, err := os.Lstat(src); err != nil {
		return false, nil
	}
	return true, copyAll(src, dst)
}

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
