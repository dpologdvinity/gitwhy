# git remote

Manage remotes: named links to other copies of the repo (e.g. `origin`). Each has a URL and fetch rule; `git fetch <name>` stores their branches as `<name>/<branch>`. Config: `remote.<name>.url`, `remote.<name>.fetch`.

No subcommand: list remote names.

Exit codes: 2 remote not found (`add`, `rename`, `remove`), 3 already exists, other non-zero otherwise.

**Usage:** `git remote [-v] [<subcommand> [<flags>] ...]`

## Flags

| Flag | What it does |
|------|--------------|
| `-v`, `--[no-]verbose` | Also show URLs (and partial-clone filters). Goes before the subcommand. |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `add` | Add a remote. |
| `rename` | Rename a remote, its tracking branches, and settings. |
| `remove`, `rm` | Delete a remote, its tracking branches, and settings. |
| `set-head` | Set or remove `<name>/HEAD`. |
| `set-branches` | Change which branches are fetched. |
| `get-url` | Print a remote's URL(s). |
| `set-url` | Change, add, or delete a remote's URL(s). |
| `show` | Show remote details. |
| `prune` | Delete tracking branches gone from the remote. |
| `update` | Fetch from several remotes or groups. |

### git remote add

`git remote add [<flags>] <name> <url>`. Example: `git remote add upstream https://github.com/org/repo.git`.

| Flag | What it does |
|------|--------------|
| `-f`, `--[no-]fetch` | Fetch right away. |
| `--tags` | Fetch all tags. Default: only tags on fetched branches. |
| `--no-tags` | Fetch no tags. |
| `-t`, `--track <branch>` | Fetch only this branch. Repeatable. |
| `-m`, `--master <branch>` | Point `<name>/HEAD` at this branch. |
| `--mirror=<fetch\|push>` | `fetch`: copy remote refs straight into your `refs/`; bare repos only, since fetch overwrites local commits. `push`: `git push` always acts like `--mirror` (force-updates and deletes remote refs). |

### git remote rename

`git remote rename [--[no-]progress] <old> <new>`. Same `<old>` and `<new>` converts an old `.git/remotes` or `.git/branches` file to config.

| Flag | What it does |
|------|--------------|
| `--[no-]progress` | Show progress. |

### git remote remove

`git remote remove <name>` (or `rm`). No flags.

### git remote set-head

`git remote set-head <name> (-a | -d | <branch>)`. Lets `origin` mean e.g. `origin/main`. Target must already be fetched.

| Flag | What it does |
|------|--------------|
| `<branch>` | Point `<name>/HEAD` at `<name>/<branch>`. |
| `-a`, `--[no-]auto` | Use the remote's own default branch. |
| `-d`, `--[no-]delete` | Remove `<name>/HEAD`. |

### git remote set-branches

`git remote set-branches [--add] <name> <branch>...`. Replaces the list (like `add -t`) unless `--add`.

| Flag | What it does |
|------|--------------|
| `--[no-]add` | Add instead of replace. |

### git remote get-url

`git remote get-url [--push] [--all] <name>`. Shows URLs after `insteadOf`/`pushInsteadOf`. Default: first fetch URL.

| Flag | What it does |
|------|--------------|
| `--[no-]push` | Push URLs instead. |
| `--[no-]all` | All URLs. |

### git remote set-url

`git remote set-url [--push] <name> <newurl> [<oldurl>]`: replace the first URL matching regex `<oldurl>` (or the first URL). Errors if no match.
`git remote set-url --add [--push] <name> <newurl>`, `git remote set-url --delete [--push] <name> <url>`.

Push and fetch URLs must point at the same repo; otherwise use two remotes.

| Flag | What it does |
|------|--------------|
| `--[no-]push` | Change push URLs instead. |
| `--[no-]add` | Add `<newurl>`. |
| `--[no-]delete` | Delete URLs matching regex `<url>`. Can't delete every fetch URL. |

### git remote show

`git remote [-v] show [-n] <name>...`. Shows URLs, default branch, tracked branches, and local pull/push setup.

| Flag | What it does |
|------|--------------|
| `-n` | Don't contact the remote; use local cache. |

### git remote prune

`git remote prune [-n] <name>...`. Like `git fetch --prune <name>` without fetching. Config may also delete local tags missing remotely (see "Pruning" in [`git fetch`](fetch.md)).

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]dry-run` | Only report. |

### git remote update

`git remote [-v] update [-p] [(<group> | <remote>)...]`. Groups: `remotes.<group>` config. No names: `remotes.default`, else all remotes without `remote.<name>.skipDefaultUpdate`.

| Flag | What it does |
|------|--------------|
| `-p`, `--[no-]prune` | Also prune. |
