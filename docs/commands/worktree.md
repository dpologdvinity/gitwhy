# git worktree

Have several working folders for one repo, each on a different branch. They share history, branches, and config; each has its own `HEAD`, index, and files. The original is the main worktree; added ones are linked. A branch can be checked out in only one worktree.

Example: urgent fix without stashing:

```sh
git worktree add -b emergency-fix ../temp main
cd ../temp   # fix, git commit -a -m 'fix'
cd -
git worktree remove ../temp
```

**Usage:** `git worktree <subcommand> [<flags>]`

`<worktree>`: a path, or any unique path ending (`ghi` for `/abc/def/ghi`).

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `add` | Create a worktree and check something out. |
| `list` | Show all worktrees. |
| `lock` | Protect from prune, move, remove. |
| `unlock` | Undo `lock`. |
| `move` | Move a linked worktree. |
| `remove` | Delete a linked worktree. |
| `prune` | Forget worktrees whose folders were deleted. |
| `repair` | Fix links after moving worktrees by hand. |

### git worktree add

Create a worktree at `<path>` and check out `<commit-ish>`. Guessing:

- `<commit-ish>` is a missing branch that exactly one remote has: `--track -b <branch> <path> <remote>/<branch>`. Several remotes: `checkout.defaultRemote` picks.
- No `<commit-ish>` (nor `-b`/`-B`/`--detach`): use a branch named after `<path>`'s last part, created from `HEAD` if missing. `git worktree add ../hotfix` makes `hotfix`.
- No `<commit-ish>` and no branches yet: unborn branch, as `--orphan`.

**Usage:** `git worktree add [<flags>] <path> [<commit-ish>]`

| Flag | What it does |
|------|--------------|
| `<path>` | New worktree folder. |
| `<commit-ish>` | Branch or commit to check out. `-` = previous branch. |
| `-f`, `--[no-]force` | Allow a branch checked out elsewhere, or a `<path>` of a missing worktree. Twice: even if locked. |
| `-b <new-branch>` | Create and check out `<new-branch>` at `<commit-ish>` (default `HEAD`). Fails if it exists. |
| `-B <new-branch>` | Like `-b`, resetting an existing branch. |
| `-d`, `--[no-]detach` | Detached `HEAD`, no branch. |
| `--[no-]checkout` | Check out files (default). `--no-`: empty, e.g. to set up sparse checkout. |
| `--[no-]lock` | Lock it immediately. |
| `--[no-]reason <string>` | With `--lock`: reason. |
| `--[no-]orphan` | Empty index and folder on a new unborn branch (named by `-b`/`-B` or `<path>`). |
| `-q`, `--[no-]quiet` | No messages. |
| `--[no-]track` | Set the source branch as upstream. Default for remote-tracking sources. See [`git branch`](branch.md). |
| `--[no-]guess-remote` | No `<commit-ish>`: base on and track the one remote branch named like `<path>`. Default: `worktree.guessRemote`. |
| `--[no-]relative-paths` | Link with relative paths (survive moving the setup; older git can't read). Default: `worktree.useRelativePaths`. |

### git worktree list

List worktrees, main first: path, commit, branch (or `detached HEAD`/`bare`), plus `locked`/`prunable`.

**Usage:** `git worktree list [-v | --porcelain [-z]]`

| Flag | What it does |
|------|--------------|
| `-v`, `--[no-]verbose` | Also show `locked`/`prunable` reasons. |
| `--[no-]porcelain` | Script output: `label value` lines, blank line between worktrees. Labels: `worktree`, `HEAD`, `branch`, `bare`, `detached`, `locked`, `prunable`. |
| `-z` | With `--porcelain`: NUL line ends (safe for odd paths). |
| `--[no-]expire <time>` | Mark missing worktrees `prunable` only if older than `<time>`. |

### git worktree lock

Lock a worktree (e.g. on removable media) so it isn't pruned, moved, or removed without extra `--force`.

**Usage:** `git worktree lock [--reason <string>] <worktree>`

| Flag | What it does |
|------|--------------|
| `<worktree>` | Worktree to lock. |
| `--[no-]reason <string>` | Reason, shown by `list -v`. |

### git worktree unlock

Unlock a worktree.

**Usage:** `git worktree unlock <worktree>`

No flags besides `<worktree>`.

### git worktree move

Move a linked worktree. Not the main one (move by hand, then `repair`) or ones with submodules.

**Usage:** `git worktree move [<flags>] <worktree> <new-path>`

| Flag | What it does |
|------|--------------|
| `<worktree>` | Worktree to move. |
| `<new-path>` | Destination. |
| `-f`, `--[no-]force` | Allow a missing worktree's path. Twice: move locked ones too. |
| `--[no-]relative-paths` | Link with relative paths. |

### git worktree remove

Delete a linked worktree. Refuses with untracked files or changes unless `-f`. Ignored files are deleted without asking. Not the main one.

**Usage:** `git worktree remove [-f] <worktree>`

| Flag | What it does |
|------|--------------|
| `<worktree>` | Worktree to remove. |
| `-f`, `--[no-]force` | Remove despite changes (**lost**) or submodules. Twice: locked too. |

### git worktree prune

Delete admin data for worktrees whose folders are gone. Also automatic (`gc.worktreePruneExpire`). Skips locked ones.

**Usage:** `git worktree prune [-n] [-v] [--expire <time>]`

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]dry-run` | Report only. |
| `-v`, `--[no-]verbose` | Report each removal. |
| `--[no-]expire <time>` | Only those older than `<time>`. |

### git worktree repair

Reconnect after moving folders by hand. Moved main: run in it. Moved linked: run inside it, or anywhere with the new paths. Moved both: run in main with linked paths.

**Usage:** `git worktree repair [<flags>] [<path>...]`

| Flag | What it does |
|------|--------------|
| `<path>...` | New linked worktree locations. |
| `--[no-]relative-paths` | Rewrite links as relative (or absolute). |

## Refs and config across worktrees

- `refs/` is shared. Per-worktree: `HEAD` and top-level refs, `refs/bisect/`, `refs/worktree/`, `refs/rewritten/`. Reach others via `main-worktree/HEAD` or `worktrees/<id>/HEAD`.
- Config is shared. Per-worktree: `git config extensions.worktreeConfig true`, then `git config --worktree <key> <value>`. Move `core.worktree`, `core.bare`, and usually `core.sparseCheckout` out of shared config. Older git refuses such repos.
- Linked worktrees have a `.git` file pointing to `$GIT_DIR/worktrees/<id>/`. Find files with `git rev-parse --git-path <file>`.
- Multiple checkouts with submodules aren't well supported.
