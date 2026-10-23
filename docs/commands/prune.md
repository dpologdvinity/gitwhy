# git prune

Delete unreachable loose objects (not reachable from any ref, reflog, or the index), loose objects already packed, and stale `.git/shallow` entries. Packed objects stay (see `git repack`). Usually run via `git gc`, which keeps recent objects; plain `git prune` doesn't. Don't run it while other git commands are running.

**Usage:** `git prune [<flags>] [--] [<head>...]`

`<head>...`: also keep objects reachable from these.

## Flags

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]dry-run` | List what would be removed. |
| `-v`, `--[no-]verbose` | List each removed object. |
| `--[no-]progress` | Show progress. |
| `--[no-]expire=<time>` | Only remove loose objects older than `<time>`. Default: any age, even just-created ones. |
| `--[no-]exclude-promisor-objects` | Partial clones: don't walk objects in promisor packs. |
| `--` | End of flags; rest are `<head>`s. |
