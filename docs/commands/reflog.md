# git reflog

Show and manage reflogs: local logs of where `HEAD` and each branch pointed. Find "lost" commits; enables names like `HEAD@{2}` or `main@{one.week.ago}`. Never pushed.

**Usage:** `git reflog [<subcommand>] [<flags>]`

Default subcommand: `show`.

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `show` | Print a ref's reflog (default `HEAD`). Same as `git log -g --abbrev-commit --pretty=oneline`. |
| `list` | List refs with a reflog. |
| `exists` | Exit 0 if a ref has a reflog. |
| `write` | Append one entry. |
| `delete` | Delete single entries (e.g. `main@{2}`). |
| `drop` | Delete whole reflogs. |
| `expire` | Delete old entries. Usually run by `git gc`. |

### git reflog show

`git reflog [show] [<log-options>] [<ref>]`. Takes every `git log` flag; see [`log`](log.md) and [revision options](../shared/revision-options.md).

### git reflog list

No flags.

### git reflog exists

`git reflog exists <ref>`. No flags.

### git reflog write

`git reflog write <ref> <old-oid> <new-oid> <message>`. Full ref name (`refs/heads/main`); full-length, existing IDs. No flags.

### git reflog delete

`git reflog delete [<flags>] <ref>@{<specifier>}...`

| Flag | What it does |
|------|--------------|
| `--rewrite` | Fix the next entry's old ID to keep the chain continuous. |
| `--updateref` | If the newest entry is deleted, move the ref to the new newest. Ignored for symbolic refs. |
| `-n`, `--dry-run` | Show what would be deleted. |
| `--verbose` | Extra details. |

### git reflog drop

`git reflog drop [--all [--single-worktree] | <refs>...]`

| Flag | What it does |
|------|--------------|
| `--all` | All refs in all worktrees. |
| `--single-worktree` | With `--all`: current worktree only. |

### git reflog expire

`git reflog expire [<flags>] [--all [--single-worktree] | <refs>...]`

| Flag | What it does |
|------|--------------|
| `--all` | Every ref's reflog. |
| `--single-worktree` | With `--all`: current worktree only. |
| `--expire=<time>` | Delete entries older than `<time>`. Default `gc.reflogExpire` (90 days). `all` = any age; `never` = keep reachable. |
| `--expire-unreachable=<time>` | Same, for entries unreachable from the current tip. Default `gc.reflogExpireUnreachable` (30 days). Takes `all`/`never`. |
| `--rewrite` | Fix the next entry's old ID to keep the chain continuous. |
| `--updateref` | If the newest entry is deleted, move the ref to the new newest. Ignored for symbolic refs. |
| `--stale-fix` | Also delete entries pointing at broken commits. Slow. Repairs old-gc damage. |
| `-n`, `--dry-run` | Show what would be deleted. |
| `--verbose` | Extra details. |
