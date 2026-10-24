// Package recover reads git's reflogs to explain what happened and how to go back,
// including for commands run in plain git.
package recover

import (
	"bufio"
	"fmt"
	"io"
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

// entry is one reflog line: the commit the ref moved to, the reflog message and when.
type entry struct {
	id, msg string
	t       time.Time
}

// reflog reads the newest n reflog entries of ref, newest first.
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

// count returns how many non-merge commits are in the revision range rng, or 0 on error.
func count(repo *git.Repo, rng string) int {
	out, err := repo.Output("rev-list", "--count", "--no-merges", rng, "--")
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(out)
	return n
}

// describe sets e.Kind and e.Text from the reflog message in e.Raw.
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
	var commitIDs, blobIDs []string
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) != 3 || f[0] != "dangling" {
			continue
		}
		switch f[1] {
		case "commit":
			commitIDs = append(commitIDs, f[2])
		case "blob":
			blobIDs = append(blobIDs, f[2])
		}
	}
	commits, err := lostCommits(repo, commitIDs)
	if err != nil {
		return nil, err
	}
	if len(commits) >= max {
		return commits[:max], nil
	}
	blobs, err := lostBlobs(repo, blobIDs[:min(len(blobIDs), max-len(commits))])
	return append(commits, blobs...), err
}

// lostCommits describes dangling commits, newest first, with one git process. gitwhy's own
// snapshot and preview commits are left out.
func lostCommits(repo *git.Repo, ids []string) ([]Lost, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	in := strings.NewReader(strings.Join(ids, "\n") + "\n")
	out, err := repo.OutputIn(in, "log", "--no-walk=unsorted", "--stdin", "--format=%H%x00%s%x00%ct")
	if err != nil {
		return nil, err
	}
	var commits []Lost
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\x00")
		if len(f) != 3 || strings.HasPrefix(f[1], "gitwhy ") {
			continue
		}
		sec, _ := strconv.ParseInt(f[2], 10, 64)
		commits = append(commits, Lost{ID: f[0], Kind: "commit", Title: f[1], Time: time.Unix(sec, 0)})
	}
	sort.Slice(commits, func(i, j int) bool { return commits[i].Time.After(commits[j].Time) })
	return commits, nil
}

// titleBytes is how much of a blob is read for its title: a short first line is all we show.
const titleBytes = 256

// lostBlobs describes dangling blobs by their first line, with one git process. Only the first
// titleBytes of each blob are kept, so a huge file doesn't fill memory.
func lostBlobs(repo *git.Repo, ids []string) ([]Lost, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var blobs []Lost
	read := func(stdout io.Reader) error {
		br := bufio.NewReader(stdout)
		for range ids {
			header, err := br.ReadString('\n')
			if err != nil {
				return err
			}
			id, size, ok := strings.Cut(strings.TrimSuffix(header, "\n"), " ")
			n, err := strconv.ParseInt(size, 10, 64)
			if !ok || err != nil {
				continue // "<id> missing": no content follows
			}
			head := make([]byte, min(n, titleBytes))
			if _, err := io.ReadFull(br, head); err != nil {
				return err
			}
			if _, err := br.Discard(int(n-int64(len(head))) + 1); err != nil { // the rest and "\n"
				return err
			}
			first, _, _ := strings.Cut(string(head), "\n")
			if len(first) > 70 {
				first = first[:70] + "..."
			}
			blobs = append(blobs, Lost{ID: id, Kind: "blob", Title: first})
		}
		return nil
	}
	in := strings.NewReader(strings.Join(ids, "\n") + "\n")
	err := repo.Stream(in, read, "cat-file", "--batch=%(objectname) %(objectsize)")
	return blobs, err
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
