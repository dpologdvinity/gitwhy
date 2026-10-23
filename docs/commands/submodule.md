# git submodule

Manage submodules: other repos inside this one (the "superproject"), pinned to a commit. The superproject records each submodule's commit; `.gitmodules` records URLs. Background: [gitsubmodules](https://git-scm.com/docs/gitsubmodules).

No subcommand: same as `status`.

**Usage:** `git submodule [--quiet] [<subcommand>] [<flags>] [--] [<path>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `-q`, `--quiet` | Only errors. Any subcommand. |
| `--cached` | With no subcommand: `status --cached`. |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `add` | Clone a repo into a path and register it. |
| `status` | Show each submodule's commit and state. |
| `init` | Copy URLs from `.gitmodules` into `.git/config`. |
| `deinit` | Unregister submodules and remove their files. |
| `update` | Clone missing submodules; move each to the recorded commit. |
| `set-branch` | Set the branch `update --remote` follows. |
| `set-url` | Change a submodule's URL. |
| `summary` | Show submodule commits between a superproject commit and now. |
| `foreach` | Run a shell command in each checked-out submodule. |
| `sync` | Update submodule URLs in config to match `.gitmodules`. |
| `absorbgitdirs` | Move a submodule's `.git` into `.git/modules/`. |

### git submodule add

**Usage:** `git submodule add [<flags>] [--] <repository> [<path>]`

`<repository>`: URL, or `./`/`../` path relative to the superproject's default remote (sibling of `bar.git`: `../foo.git`). `<path>` defaults to the repo name. An existing repo at `<path>` is staged without cloning. Commit the result.

| Flag | What it does |
|------|--------------|
| `-b <branch>`, `--[no-]branch=<branch>` | Clone this branch and record it for `update --remote`. `.` = same name as current branch. Default: remote `HEAD`. |
| `-f`, `--[no-]force` | Add even if the path is ignored. If the name is taken, pick a new one (`child1`, ...). |
| `--[no-]name=<name>` | Config name instead of path. Valid directory name, no trailing `/`. |
| `--[no-]reference=<repository>` | Borrow objects from a local repo (see `git clone --reference` warnings). |
| `--[no-]dissociate` | After `--reference`, copy borrowed objects and drop the dependency. |
| `--[no-]depth=<n>` | Shallow clone: last `<n>` commits. |
| `--[no-]ref-format=<format>` | Clone ref format: `files` or `reftable`. |
| `--[no-]progress` | Show progress even off a terminal. |

### git submodule status

**Usage:** `git submodule status [<flags>] [--] [<path>...]`

Prints `<commit> <path> (<describe>)`. Prefix `-`: not initialized. `+`: checkout differs from recorded commit. `U`: conflicts.

| Flag | What it does |
|------|--------------|
| `--[no-]cached` | Show the recorded commit, not the checked-out one. |
| `--[no-]recursive` | Include nested submodules. |

### git submodule init

**Usage:** `git submodule init [--] [<path>...]`

Copies `submodule.<name>.url` (and `update`, unless a custom command) from `.gitmodules` to `.git/config`, without overwriting. No paths: all, or those in `submodule.active`. Edit URLs if needed, then `update`. Or use `update --init`.

No extra flags.

### git submodule deinit

**Usage:** `git submodule deinit [-f] (--all | [--] <path>...)`

Removes the submodule's `.git/config` section and empties its directory (refuses if it has changes or untracked files); `update`, `foreach`, `sync` then skip it. To delete it from the project, use `git rm`.

| Flag | What it does |
|------|--------------|
| `--[no-]all` | Unregister all. Required with no paths. |
| `-f`, `--[no-]force` | Remove files even with local changes or untracked files (**lost**). |

### git submodule update

**Usage:** `git submodule update [<flags>] [--] [<path>...]`

Update mode: flag, else `submodule.<name>.update`, else checkout. Config may also be `!<command>` (runs `<command> <commit>`; `.git/config` only, not `.gitmodules`) or `none` (skip). Neither works on the command line.

| Flag | What it does |
|------|--------------|
| `-i`, `--[no-]init` | (`-i` restored in 2.56+.) `init` uninitialized submodules first. |
| `--[no-]remote` | Use the latest commit of the remote branch, not the recorded one. Branch: `submodule.<name>.branch`, else remote `HEAD`. Fetches first. |
| `-N`, `--no-fetch` | Don't fetch. |
| `--fetch` | Fetch (default). Cancels `--no-fetch`. |
| `--[no-]checkout` | Check out on a detached HEAD (default). |
| `-m`, `--[no-]merge` | Merge the commit into the current branch. |
| `-r`, `--[no-]rebase` | Rebase the current branch onto the commit. |
| `-f`, `--[no-]force` | With checkout: discard local changes; check out even if already there. |
| `--[no-]recursive` | Include nested submodules. |
| `--[no-]reference=<repo>` | Borrow objects from a local repo when cloning. |
| `--[no-]dissociate` | Use `--reference` only during clone, then copy objects. |
| `--[no-]depth=<n>` | Shallow clone: last `<n>` commits. |
| `--[no-]recommend-shallow` | Honor `submodule.<name>.shallow` on first clone (default). |
| `--[no-]single-branch` | Clone only `HEAD` or the `--branch` branch. |
| `--[no-]filter=<filter-spec>` | Partial clone filter, e.g. `blob:none`. |
| `-j <n>`, `--[no-]jobs=<n>` | Clone `<n>` at once. Default `submodule.fetchJobs`. |
| `--[no-]ref-format=<format>` | New clone ref format: `files` or `reftable`. |
| `--[no-]require-init` | Refuse to clone into a non-empty directory. Implies `--init`. |
| `--[no-]progress` | Show progress even off a terminal. |

### git submodule set-branch

**Usage:** `git submodule set-branch (-b <branch> | -d) [--] <path>`

| Flag | What it does |
|------|--------------|
| `-b <branch>`, `--[no-]branch=<branch>` | Follow `<branch>` with `update --remote`. |
| `-d`, `--[no-]default` | Remove setting; follow remote `HEAD`. |

### git submodule set-url

**Usage:** `git submodule set-url [--] <path> <newurl>`

Changes the URL in `.gitmodules` and syncs config. No extra flags.

### git submodule summary

**Usage:** `git submodule summary [<flags>] [<commit>] [--] [<path>...]`

Lists submodule commits between `<commit>` (default `HEAD`) and the working tree. Like `git diff --submodule=log`.

| Flag | What it does |
|------|--------------|
| `--[no-]cached` | Compare against the index. |
| `--[no-]files` | Compare superproject index with each submodule's `HEAD`. Not with `--cached` or `<commit>`. |
| `-n <n>`, `--[no-]summary-limit=<n>` | Max `<n>` commits total. `0` disables; negative unlimited (default). |
| `--[no-]for-status` | Internal: skip `ignore = all` submodules. Used by `git status`. |

### git submodule foreach

**Usage:** `git submodule foreach [--recursive] <command>`

Run `<command>` in a shell in each checked-out submodule, printing its name first (unless `--quiet`). Stops at the first non-zero exit; add `|| :` to continue. Example: `git submodule foreach 'echo $sm_path $(git rev-parse HEAD)'`.

Variables:

| Variable | Value |
|----------|-------|
| `$name` | Name in `.gitmodules`. |
| `$sm_path` | Path in the parent repo (old alias `$path`). |
| `$displaypath` | Path from the current directory. |
| `$sha1` | Recorded commit. |
| `$toplevel` | Parent repo absolute path. |

| Flag | What it does |
|------|--------------|
| `--[no-]recursive` | Include nested submodules. |

### git submodule sync

**Usage:** `git submodule sync [--recursive] [--] [<path>...]`

Use after an upstream URL change. Only affects submodules with a URL in `.git/config`.

| Flag | What it does |
|------|--------------|
| `--[no-]recursive` | Include nested submodules. |

### git submodule absorbgitdirs

**Usage:** `git submodule absorbgitdirs [--] [<path>...]`

Move a submodule's embedded `.git` directory to `.git/modules/`, leaving a `.git` file pointing there. Always recursive. No extra flags.
