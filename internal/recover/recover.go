// Package recover reads git's reflogs to explain what happened and how to go back,
// including for commands run in plain git.
package recover

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dpologdvinity/gitwhy/internal/git"
)

// Event is one thing that happened to a branch (or to HEAD).
type Event struct {
	Time   time.Time
	Ref    string // "refs/heads/main" or "HEAD"
	Before string // commit before the event ("" if the ref was created)
	After  string
	Kind   string // commit, amend, reset, merge, pull, rebase, cherry-pick, revert, checkout, branch, other
	Text   string // plain English
	Raw    string // reflog message
}

// Undo returns the git args that put the ref back to Before, for the current state.
func (e Event) Undo(repo *git.Repo) []string {
	if e.Before == "" {
		return nil
	}
	if e.Kind == "checkout" {
		return []string{"switch", "--detach", e.Before}
	}
	cur, _ := repo.Output("symbolic-ref", "-q", "HEAD")
	if e.Ref == cur || e.Ref == "HEAD" {
		return []string{"reset", "--keep", e.Before}
	}
	return []string{"branch", "-f", strings.TrimPrefix(e.Ref, "refs/heads/"), e.Before}
}

type entry struct {
	id, msg string
	t       time.Time
}

func reflog(repo *git.Repo, ref string, n int) ([]entry, error) {
	out, err := repo.Output("log", "-g", "-n", strconv.Itoa(n), "--format=%H%x00%gs%x00%gt", ref, "--")
	if err != nil {
		return nil, err
	}
	var es []entry
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\x00")
		if len(f) != 3 {
			continue
		}
		sec, _ := strconv.ParseInt(f[2], 10, 64)
		es = append(es, entry{id: f[0], msg: f[1], t: time.Unix(sec, 0)})
	}
	return es, nil
}

// Events lists recent events on the current branch, newest first, with rebases grouped into one.
// If HEAD is detached, HEAD's reflog is used.
func Events(repo *git.Repo, n int) ([]Event, error) {
	ref, err := repo.Output("symbolic-ref", "-q", "HEAD")
	if err != nil || !repo.OK("rev-parse", "--verify", "-q", ref) {
		ref = "HEAD"
	}
	es, err := reflog(repo, ref, n*4)
	if err != nil {
		return nil, err
	}
	var out []Event
	for i := 0; i < len(es) && len(out) < n; i++ {
		e := es[i]
		before := ""
		if i+1 < len(es) {
			before = es[i+1].id
		}
		ev := Event{Time: e.t, Ref: ref, Before: before, After: e.id, Raw: e.msg}
		// Branch reflogs record a whole rebase as one "rebase (finish)" entry; HEAD's reflog has
		// start/pick/finish entries. Group a HEAD rebase back to its start.
		if ref == "HEAD" && strings.HasPrefix(e.msg, "rebase") && !strings.Contains(e.msg, "(start)") {
			j := i
			for j+1 < len(es) && strings.HasPrefix(es[j+1].msg, "rebase") && !strings.Contains(es[j].msg, "(start)") {
				j++
			}
			if j+1 < len(es) {
				ev.Before = es[j+1].id
			} else {
				ev.Before = ""
			}
			i = j
		}
		describe(repo, &ev)
		out = append(out, ev)
	}
	return out, nil
}

func count(repo *git.Repo, rng string) int {
	out, err := repo.Output("rev-list", "--count", "--no-merges", rng, "--")
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(out)
	return n
}

func describe(repo *git.Repo, e *Event) {
	action, detail, _ := strings.Cut(e.Raw, ": ")
	action = strings.TrimSpace(action)
	name := strings.TrimPrefix(e.Ref, "refs/heads/")
	switch {
	case action == "commit" || action == "commit (initial)":
		e.Kind, e.Text = "commit", fmt.Sprintf("Committed %q", detail)
	case action == "commit (amend)":
		e.Kind, e.Text = "amend", fmt.Sprintf("Amended the last commit (now %q)", detail)
	case action == "commit (merge)":
		e.Kind, e.Text = "merge", "Finished a merge after resolving conflicts"
	case strings.HasPrefix(action, "reset"):
		lost := 0
		if e.Before != "" {
			lost = count(repo, e.After+".."+e.Before)
		}
		e.Kind, e.Text = "reset", fmt.Sprintf("Reset %s %s", name, strings.TrimPrefix(detail, "moving "))
		if lost > 0 {
			e.Text += fmt.Sprintf(" (%d commit(s) left the branch)", lost)
		}
	case strings.HasPrefix(action, "pull"):
		in := 0
		if e.Before != "" {
			in = count(repo, e.Before+".."+e.After)
		}
		e.Kind, e.Text = "pull", fmt.Sprintf("Pulled %d commit(s) into %s (%s)", in, name, detail)
		if strings.Contains(action, "rebase") {
			e.Text = fmt.Sprintf("Pulled with rebase into %s", name)
		}
	case strings.HasPrefix(action, "merge"):
		in := 0
		if e.Before != "" {
			in = count(repo, e.Before+".."+e.After)
		}
		e.Kind, e.Text = "merge", fmt.Sprintf("Merged %s into %s, bringing %d commit(s) (%s)", strings.TrimPrefix(action, "merge "), name, in, detail)
	case strings.HasPrefix(action, "rebase"):
		e.Kind, e.Text = "rebase", fmt.Sprintf("Rebased %s (commits rewritten with new IDs)", name)
	case strings.HasPrefix(action, "cherry-pick"):
		e.Kind, e.Text = "cherry-pick", fmt.Sprintf("Copied commit %q onto %s", detail, name)
	case strings.HasPrefix(action, "revert"):
		e.Kind, e.Text = "revert", fmt.Sprintf("Reverted: %q", detail)
	case strings.HasPrefix(action, "checkout") || strings.HasPrefix(action, "switch"):
		e.Kind, e.Text = "checkout", "Switched "+strings.TrimPrefix(detail, "moving ")
	case strings.HasPrefix(action, "branch"):
		e.Kind, e.Text = "branch", "Branch "+detail
	default:
		e.Kind, e.Text = "other", e.Raw
	}
}

// Lost is a dangling object git still has: a commit no branch points to, or a staged file that
// was never committed.
type Lost struct {
	ID    string
	Kind  string // "commit" or "blob"
	Title string // commit subject or the file's first line
	Time  time.Time
}

// LostObjects lists dangling commits and blobs, newest commits first.
func LostObjects(repo *git.Repo, max int) ([]Lost, error) {
	out, err := repo.Output("fsck", "--no-reflogs", "--dangling", "--no-progress")
	if err != nil && out == "" {
		return nil, err
	}
	var commits, blobs []Lost
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) != 3 || f[0] != "dangling" {
			continue
		}
		switch f[1] {
		case "commit":
			info, err := repo.Output("log", "-1", "--format=%s%x00%ct", f[2])
			if err != nil {
				continue
			}
			subj, ts, _ := strings.Cut(info, "\x00")
			if strings.HasPrefix(subj, "gitwhy ") { // our own snapshots and previews
				continue
			}
			sec, _ := strconv.ParseInt(ts, 10, 64)
			commits = append(commits, Lost{ID: f[2], Kind: "commit", Title: subj, Time: time.Unix(sec, 0)})
		case "blob":
			head, _ := repo.Output("cat-file", "-p", f[2])
			first, _, _ := strings.Cut(head, "\n")
			if len(first) > 70 {
				first = first[:70] + "..."
			}
			blobs = append(blobs, Lost{ID: f[2], Kind: "blob", Title: first})
		}
	}
	sort.Slice(commits, func(i, j int) bool { return commits[i].Time.After(commits[j].Time) })
	all := append(commits, blobs...)
	if len(all) > max {
		all = all[:max]
	}
	return all, nil
}

// Ago formats a time as "5m ago".
func Ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}
