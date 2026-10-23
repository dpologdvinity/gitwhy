# git history

Experimental: edit one past commit and auto-replay the commits after it. Simpler than `git rebase -i` for single fixes. Moves every local branch built on the changed commit by default. Rewrites commit IDs: already-pushed commits then need a force-push.

Unlike [`git rebase`](rebase.md): most subcommands work in bare repos (no working tree; `fixup` reads the index), no hooks. Can't handle merge commits (use `git rebase --rebase-merges`) or conflicts: a conflicting replay errors and changes nothing.

**Usage:** `git history <subcommand> <commit> [<flags>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `drop` | (2.56+) Delete a commit; replay descendants onto its parent. |
| `fixup` | Fold staged changes into an older commit. |
| `reword` | Change an older commit's message. |
| `split` | Split a commit in two by picking hunks. |

### git history drop

(2.56+) Remove `<commit>`; replay later commits onto its parent. Not for root or merge commits. Updates index and working tree if `HEAD` changes; stops first if that would overwrite local changes.

**Usage:** `git history drop <commit> [--dry-run] [--update-refs=(branches|head)] [--empty=(drop|keep|abort)]`

| Flag | What it does |
|------|--------------|
| `<commit>` | Commit to remove. |
| `--dry-run` | Move no refs; print updates in `git update-ref --stdin` format. |
| `--update-refs=(branches\|head)` | Refs to move: `branches` (default, all built on it) or `head` (current only). |
| `--empty=(drop\|keep\|abort)` | Replayed commit ends up empty: `drop` (default), `keep`, or `abort`. |

### git history fixup

Fold staged changes into `<commit>`, like `git commit --fixup` plus `git rebase --autosquash`. Keeps message and author. Errors on conflict.

**Usage:** `git history fixup <commit> [--dry-run] [--update-refs=(branches|head)] [--reedit-message] [--empty=(drop|keep|abort)]`

| Flag | What it does |
|------|--------------|
| `<commit>` | Commit to fix. |
| `--dry-run` | Move no refs; print updates in `git update-ref --stdin` format. |
| `--update-refs=(branches\|head)` | Refs to move: `branches` (default, all built on it) or `head` (current only). |
| `--reedit-message` | Also edit the target's message. |
| `--empty=(drop\|keep\|abort)` | Target or later commit becomes empty: `drop` (default; descendants go to its parent), `keep`, or `abort`. Can't drop the root. |

### git history reword

Edit `<commit>`'s message only.

**Usage:** `git history reword <commit> [--dry-run] [--update-refs=(branches|head)]`

| Flag | What it does |
|------|--------------|
| `<commit>` | Commit to reword. |
| `--dry-run` | Move no refs; print updates in `git update-ref --stdin` format. |
| `--update-refs=(branches\|head)` | Refs to move: `branches` (default, all built on it) or `head` (current only). |

### git history split

Split `<commit>` in two: picked hunks (as in `git add -p`) go into a new commit just before it. Edit both messages; both keep the author. Picking all or no hunks is an error.

**Usage:** `git history split <commit> [--dry-run] [--update-refs=(branches|head)] [--] [<pathspec>...]`

| Flag | What it does |
|------|--------------|
| `<commit>` | Commit to split. |
| `<pathspec>...` | Only offer hunks from these files. |
| `--dry-run` | Move no refs; print updates in `git update-ref --stdin` format. |
| `--update-refs=(branches\|head)` | Refs to move: `branches` (default, all built on it) or `head` (current only). |
| `--` | Rest are file names. |
