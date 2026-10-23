# git gui

GUI for committing: stage, write messages, amend, branch, merge, fetch, push. No history view; use [gitk](gitk.md) (Repository > Visualize).

Often a separate package (`git-gui` on Debian/Ubuntu). Needs Tcl/Tk.

**Usage:** `git gui [<command>] [<arguments>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| (none) | Open the main window. |
| `blame` | Show who wrote each line of a file. |
| `browser` | Browse a commit's files; opens picks in blame. |
| `citool` | Make one commit, then exit. |
| `version` | Print git gui version. |

### git gui blame

**Usage:** `git gui blame [<rev>] <file>`

Shows who wrote and who last moved each line. Without `<rev>`, uses the working copy (uncommitted lines marked). With `<rev>`, the file at that commit.

| Flag | What it does |
|------|--------------|
| `--line=<n>` | Open at line `<n>`. |

### git gui browser

**Usage:** `git gui browser <rev>`

No flags. Example: `git gui browser maint`.

### git gui citool

Same as [`git citool`](citool.md). Exits nonzero if closed without committing.

| Flag | What it does |
|------|--------------|
| `--amend` | Amend the last commit. |
| `--nocommit` | Don't commit; exit 0 when done. Still refuses with unresolved conflicts, so works as a conflict helper. |

### git gui version

No flags.
