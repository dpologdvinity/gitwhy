# git stash

Set uncommitted changes aside, leaving files matching `HEAD`; reapply later on any commit. Untracked files stay unless `-u`. Handy before `git pull` or an urgent fix.

Stashes are a stack: newest `stash@{0}` (in `refs/stash`), then `stash@{1}`, etc.; `stash@{2.hours.ago}` and bare `<n>` work too. Plain `git stash` = `git stash push`. Default description: `WIP on <branch>: <commit>`.

Each stash is a commit: parent 1 is `HEAD`, parent 2 holds the index state, parent 3 (if any) untracked files.

**Usage:** `git stash [<subcommand>] [<flags>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `push` | Save changes as a new stash, reset files to `HEAD`. Default. |
| `save` | Deprecated `push`; takes a message, not paths. |
| `list` | List stashes. |
| `show` | Show a stash's changes. |
| `pop` | Apply and remove a stash. |
| `apply` | Apply, keep the stash. |
| `branch` | Branch from where a stash was made and apply it. |
| `drop` | Delete one stash. |
| `clear` | Delete all stashes. |
| `create` | Scripts: make a stash commit, print ID, don't list it. |
| `store` | Scripts: list a `create` commit. |
| `export` | Pack stashes into a pushable commit chain. |
| `import` | Add stashes from an `export` chain. |

### git stash push

`git stash [push] [<flags>] [--] [<pathspec>...]`. Without `push`, paths need `--` (so a typo'd subcommand doesn't stash).

| Flag | What it does |
|------|--------------|
| `<pathspec>...` | Stash only matching files. |
| `-m`, `--[no-]message=<message>` | Description. |
| `-u`, `--[no-]include-untracked` | Also stash untracked files, then delete them. |
| `-a`, `--[no-]all` | Also stash untracked and ignored files, then delete them. |
| `-k`, `--[no-]keep-index` | Leave staged changes in place (still saved). |
| `-S`, `--[no-]staged` | Stash only staged changes. `-p` overrides. |
| `-p`, `--[no-]patch` | Pick hunks interactively (`git add -p` keys). Implies `--keep-index`. |
| `--[no-]auto-advance` | With `-p`: go to the next file after its last hunk (default). `--no-auto-advance`: stay; `<`/`>` switch files. |
| `-U<n>`, `--unified=<n>` | With `-p`: context lines. Default `diff.context` or 3. |
| `--inter-hunk-context=<n>` | With `-p`: join hunks up to `<n>` lines apart. Default `diff.interHunkContext` or 0. |
| `-q`, `--[no-]quiet` | No messages. |
| `--[no-]pathspec-from-file=<file>` | Read paths from `<file>` (`-` = stdin), one per line. |
| `--[no-]pathspec-file-nul` | With `--pathspec-from-file`: NUL-separated, literal. |
| `--` | Paths follow. |

### git stash save

Deprecated: use `push`. `git stash save [<flags>] [<message>]`. Non-flag words form the message; no paths. Flags as `push` minus pathspec ones: `-m`, `-u`, `-a`, `-k`, `-S`, `-p`, `--auto-advance`, `-U`, `--inter-hunk-context`, `-q`.

### git stash list

`git stash list [<log-options>]`. Lines like `stash@{0}: WIP on main: 6ebd0e2 Fix typo`. Takes `git log` flags (`--stat`, `-p`, `--date=relative`; see [`git log`](log.md), [commit selection options](../shared/revision-options.md)).

### git stash show

`git stash show [<flags>] [<diff-options>] [<stash>]`. Diff against the stash's base commit. Default diffstat (`stash.showStat`); `-p` for patch (`stash.showPatch`). Also takes all [diff options](../shared/diff-options.md).

| Flag | What it does |
|------|--------------|
| `<stash>` | Default `stash@{0}`. |
| `-u`, `--[no-]include-untracked` | Include stashed untracked files. Default from `stash.showIncludeUntracked`. |
| `--only-untracked` | Only stashed untracked files. |

### git stash pop

`git stash pop [--index] [-q] [<stash>]`. Apply and remove. On conflict the stash is kept: resolve, then `git stash drop`. Refuses if it would overwrite local edits.

| Flag | What it does |
|------|--------------|
| `<stash>` | Default `stash@{0}`. |
| `--[no-]index` | Also restore staged state. May fail on conflicts. Default from `stash.index`. |
| `-q`, `--[no-]quiet` | Errors only. |

### git stash apply

`git stash apply [<flags>] [<stash>]`. Like `pop`, keeps the stash. `<stash>` can be any stash-like commit.

| Flag | What it does |
|------|--------------|
| `<stash>` | Default `stash@{0}`. |
| `--[no-]index` | Also restore staged state. May fail on conflicts. Default from `stash.index`. |
| `-q`, `--[no-]quiet` | Errors only. |
| `--[no-]label-ours=<label>` | Conflict label for your side. Default `Updated upstream`. |
| `--[no-]label-theirs=<label>` | Conflict label for the stash. Default `Stashed changes`. |
| `--[no-]label-base=<label>` | Conflict label for the base. Default `Stash base`. Only with `merge.conflictStyle=diff3`. |

### git stash branch

`git stash branch <branch> [<stash>]`. Create `<branch>` at the stash's base commit, switch, apply with staged state, drop if a `stash@{...}` entry. For when `apply` conflicts. No flags.

### git stash drop

`git stash drop [-q] [<stash>]`. Delete one stash (default `stash@{0}`).

| Flag | What it does |
|------|--------------|
| `-q`, `--[no-]quiet` | Errors only. |

### git stash clear

Delete all stashes. No flags. Not normally recoverable. Last resort:
`git fsck --unreachable | grep commit | cut -d' ' -f3 | xargs git log --merges --no-walk --grep=WIP`

### git stash create

`git stash create [<message>]`. Make a stash commit and print its ID; changes no files or list. For scripts. No flags.

### git stash store

`git stash store [-m <message>] [-q] <commit>`. Add a `create` commit to the list. For scripts.

| Flag | What it does |
|------|--------------|
| `-m`, `--[no-]message=<message>` | Description. |
| `-q`, `--[no-]quiet` | No messages. |

### git stash export

`git stash export (--print | --to-ref <ref>) [<stash>...]`. Turn stashes into a normal commit chain to push/fetch. Restore with `import`.

| Flag | What it does |
|------|--------------|
| `<stash>...` | Stashes to export. Default all. |
| `--print` | Print the chain ID; no ref. |
| `--[no-]to-ref=<ref>` | Save chain to this ref, e.g. `refs/heads/my-stashes`. |

### git stash import

`git stash import <commit>`. Add stashes from an `export` chain. `clear` first to replace. No flags.
