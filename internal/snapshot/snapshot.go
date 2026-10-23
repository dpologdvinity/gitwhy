// Package snapshot saves and restores repository state around gitwhy-run commands.
//
// A snapshot holds every ref and HEAD, a byte copy of the index, a commit of the whole working
// tree (tracked + untracked, not ignored), the state files of an in-progress merge/rebase/etc.,
// and the stash list. It lives in <gitdir>/gitwhy/snapshots/<id>/; the working-tree commit is
// kept alive by refs/gitwhy/snapshots/<id>.
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
	Refs    map[string]string `json:"refs"`  // refname -> object ID (no refs/gitwhy, no refs/stash)
	Work    string            `json:"work"`  // commit of the working tree
	Stashes []Stash           `json:"stashes"`
	State   []string          `json:"state"`    // state files that existed
	HasIdx  bool              `json:"hasIndex"` // whether .git/index existed

	dir string
}

func root(r *git.Repo) string { return filepath.Join(r.GitDir, "gitwhy", "snapshots") }

// Take saves the current state. extra lists ignored paths to include because the next command
// will delete them (from its own dry run).
func Take(r *git.Repo, label string, extra []string) (*Snapshot, error) {
	if _, err := os.Stat(filepath.Join(r.GitDir, "index.lock")); err == nil {
		return nil, errors.New("another git process is running (.git/index.lock exists)")
	}
	now := time.Now()
	s := &Snapshot{ID: strconv.FormatInt(now.UnixNano(), 36), Time: now, Label: label, Refs: map[string]string{}}
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

	head, err := readHead(r)
	if err != nil {
		return nil, err
	}
	s.Head = head
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
	if _, err := w.Output("add", "-A", "--ignore-errors", "--", ":/"); err != nil {
		return nil, fmt.Errorf("saving working files: %w", err)
	}
	if len(extra) > 0 {
		args := append([]string{"add", "-f", "--ignore-errors", "--"}, extra...)
		if _, err := w.Output(args...); err != nil {
			return nil, fmt.Errorf("saving ignored files: %w", err)
		}
	}
	tree, err := w.Output("write-tree")
	if err != nil {
		return nil, fmt.Errorf("saving working files: %w", err)
	}
	os.Remove(tmpIdx)
	s.Work, err = r.With(commitEnv...).Output("commit-tree", tree, "-m", "gitwhy snapshot: "+label)
	if err != nil {
		return nil, err
	}
	if _, err := r.Output("update-ref", refPrefix+s.ID, s.Work); err != nil {
		return nil, err
	}

	// In-progress operation state.
	for _, f := range stateFiles {
		src := filepath.Join(r.GitDir, f)
		if _, err := os.Lstat(src); err != nil {
			continue
		}
		if err := copyAll(src, filepath.Join(s.dir, "state", f)); err != nil {
			return nil, err
		}
		s.State = append(s.State, f)
	}

	if err := s.save(); err != nil {
		return nil, err
	}
	ok = true
	prune(r)
	return s, nil
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
	return os.WriteFile(filepath.Join(s.dir, "meta.json"), data, 0o644)
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
		dir := filepath.Join(root(r), e.Name())
		data, err := os.ReadFile(filepath.Join(dir, "meta.json"))
		if err != nil {
			continue // incomplete snapshot
		}
		var s Snapshot
		if json.Unmarshal(data, &s) != nil {
			continue
		}
		s.dir = dir
		out = append(out, &s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	return out, nil
}

// Delete removes a snapshot and its ref.
func Delete(r *git.Repo, s *Snapshot) {
	r.Output("update-ref", "-d", refPrefix+s.ID)
	os.RemoveAll(s.dir)
}

func prune(r *git.Repo) {
	all, err := List(r)
	if err != nil {
		return
	}
	for i := Keep; i < len(all); i++ {
		Delete(r, all[i])
	}
}

// Restore puts the repository back to s. It takes a snapshot of the current state first and
// returns it, so the restore itself can be undone.
func Restore(r *git.Repo, s *Snapshot) (*Snapshot, error) {
	cur, err := Take(r, "before undo of: "+s.Label, nil)
	if err != nil {
		return nil, fmt.Errorf("saving current state first: %w", err)
	}

	// 1. Operation state: clear what's there now, put back what was saved.
	for _, f := range stateFiles {
		if err := os.RemoveAll(filepath.Join(r.GitDir, f)); err != nil {
			return cur, err
		}
	}
	for _, f := range s.State {
		if err := copyAll(filepath.Join(s.dir, "state", f), filepath.Join(r.GitDir, f)); err != nil {
			return cur, err
		}
	}

	// 2. Refs: one transaction, touching only refs that differ.
	var tx strings.Builder
	for name, id := range s.Refs {
		if cur.Refs[name] != id {
			fmt.Fprintf(&tx, "update %s %s\n", name, id)
		}
	}
	for name := range cur.Refs {
		if _, ok := s.Refs[name]; !ok {
			fmt.Fprintf(&tx, "delete %s\n", name)
		}
	}
	if tx.Len() > 0 {
		if _, err := r.OutputIn(strings.NewReader(tx.String()), "update-ref", "--stdin"); err != nil {
			return cur, fmt.Errorf("restoring branches: %w", err)
		}
	}

	// 3. HEAD.
	if ref, ok := strings.CutPrefix(s.Head, "ref: "); ok {
		_, err = r.Output("symbolic-ref", "HEAD", ref)
	} else {
		_, err = r.Output("update-ref", "--no-deref", "HEAD", s.Head)
	}
	if err != nil {
		return cur, fmt.Errorf("restoring HEAD: %w", err)
	}

	// 4. Working files: delete files added since, rewrite files that differ.
	if err := restoreFiles(r, s.Work, cur.Work); err != nil {
		return cur, err
	}

	// 5. Index.
	idx := filepath.Join(r.GitDir, "index")
	if s.HasIdx {
		if err := copyFile(filepath.Join(s.dir, "index"), idx+".gitwhy-tmp"); err != nil {
			return cur, err
		}
		if err := os.Rename(idx+".gitwhy-tmp", idx); err != nil {
			return cur, err
		}
		r.Output("update-index", "-q", "--refresh") // refresh stat info; dirty entries are expected
	} else {
		os.Remove(idx)
	}

	// 6. Stash list.
	if !sameStashes(s.Stashes, cur.Stashes) {
		r.Output("update-ref", "-d", "refs/stash")
		for i := len(s.Stashes) - 1; i >= 0; i-- {
			st := s.Stashes[i]
			if _, err := r.Output("stash", "store", "-m", st.Message, st.Commit); err != nil {
				return cur, fmt.Errorf("restoring stash: %w", err)
			}
		}
	}
	return cur, nil
}

func restoreFiles(r *git.Repo, want, have string) error {
	added, err := r.Output("diff-tree", "-r", "-z", "--name-only", "--no-renames", "--diff-filter=A", want, have)
	if err != nil {
		return err
	}
	for _, p := range splitZ(added) {
		full := filepath.Join(r.Top, filepath.FromSlash(p))
		if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		removeEmptyParents(r.Top, filepath.Dir(full))
	}
	changed, err := r.Output("diff-tree", "-r", "-z", "--name-only", "--no-renames", "--diff-filter=DMT", want, have)
	if err != nil {
		return err
	}
	paths := splitZ(changed)
	if len(paths) == 0 {
		return nil
	}
	tmp := filepath.Join(r.GitDir, "gitwhy", "restore-index")
	defer os.Remove(tmp)
	w := r.With("GIT_INDEX_FILE=" + tmp)
	if _, err := w.Output("read-tree", want); err != nil {
		return err
	}
	// Remove first so a file replaced by a folder (or the reverse) can be written.
	for _, p := range paths {
		os.RemoveAll(filepath.Join(r.Top, filepath.FromSlash(p)))
	}
	in := strings.Join(paths, "\x00") + "\x00"
	if _, err := w.OutputIn(strings.NewReader(in), "checkout-index", "-f", "-z", "--stdin"); err != nil {
		return fmt.Errorf("restoring files: %w", err)
	}
	return nil
}

func readHead(r *git.Repo) (string, error) {
	if ref, err := r.Output("symbolic-ref", "-q", "HEAD"); err == nil {
		return "ref: " + ref, nil
	}
	return r.Output("rev-parse", "HEAD")
}

func refs(r *git.Repo) (map[string]string, error) {
	out, err := r.Output("for-each-ref", "--format=%(objectname) %(refname)")
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		id, name, ok := strings.Cut(line, " ")
		if !ok || strings.HasPrefix(name, "refs/gitwhy/") || name == "refs/stash" {
			continue
		}
		m[name] = id
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

func splitZ(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "\x00") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func removeEmptyParents(top, dir string) {
	for dir != top && strings.HasPrefix(dir, top) {
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
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
