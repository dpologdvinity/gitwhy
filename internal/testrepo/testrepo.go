// Package testrepo builds throwaway git repositories for tests.
package testrepo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Repo is a temp repository with helpers that fail the test on error.
type Repo struct {
	*git.Repo
	t *testing.T
}

// New makes an empty repo on branch main with a test identity.
func New(t *testing.T) *Repo {
	t.Helper()
	dir := t.TempDir()
	g := git.Runner{Dir: dir}
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.name", "Test"},
		{"config", "user.email", "test@example.com"},
		{"config", "commit.gpgsign", "false"},
		{"config", "core.hooksPath", t.TempDir()},
		{"config", "core.autocrlf", "false"},
	} {
		if _, err := g.Output(args...); err != nil {
			t.Fatal(err)
		}
	}
	r, err := git.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	return &Repo{Repo: r, t: t}
}

// Git runs git and returns stdout.
func (r *Repo) Git(args ...string) string {
	r.t.Helper()
	out, err := r.Output(args...)
	if err != nil {
		r.t.Fatal(err)
	}
	return out
}

// Write writes a file relative to the repo root.
func (r *Repo) Write(path, content string) {
	r.t.Helper()
	p := filepath.Join(r.Top, path)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// Read returns a file's content, or "<missing>".
func (r *Repo) Read(path string) string {
	data, err := os.ReadFile(filepath.Join(r.Top, path))
	if err != nil {
		return "<missing>"
	}
	return string(data)
}

// Commit writes files and commits them.
func (r *Repo) Commit(msg string, files ...string) string {
	r.t.Helper()
	for i := 0; i+1 < len(files); i += 2 {
		r.Write(files[i], files[i+1])
	}
	r.Git("add", "-A")
	r.Git("commit", "-q", "--allow-empty", "-m", msg)
	return r.Git("rev-parse", "HEAD")
}

// State is a comparable summary of everything a snapshot should restore.
func (r *Repo) State() string {
	r.t.Helper()
	var b strings.Builder
	b.WriteString(r.Git("status", "--porcelain=v2", "--branch", "--untracked-files=all", "--ignored"))
	b.WriteString("\n--refs\n" + r.Git("for-each-ref", "--format=%(objectname) %(refname)", "--exclude=refs/gitwhy/"))
	b.WriteString("\n--stash\n" + r.Git("stash", "list", "--format=%H %gs"))
	b.WriteString("\n--index\n" + r.Git("ls-files", "-s"))
	return b.String()
}
