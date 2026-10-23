# git lfs

Git LFS (Large File Storage), a separate add-on: big files are stored as small "pointers" in the repo, with content on an LFS server and in `.git/lfs/objects`. Content downloads on checkout and uploads on push via filters and hooks from `git lfs install`.

Typical start: `git lfs install` once per machine, then `git lfs track "*.iso"`, `git add .gitattributes`, commit and push.

Documented here: git-lfs 3.4.1. Later additions are marked, e.g. "(git-lfs 3.7+)".

**Usage:** `git lfs <subcommand> [<flags>] [<args>]`

`--include` / `--exclude` (`-I` / `-X`) take comma-separated gitignore-style patterns. They override `lfs.fetchinclude` / `lfs.fetchexclude` for one run; empty clears.

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `install` | Set up LFS filters and repo hooks. |
| `uninstall` | Remove LFS filters and hooks. |
| `track` | Store a pattern in LFS (via `.gitattributes`); no arguments lists patterns. |
| `untrack` | Stop storing a pattern in LFS. |
| `status` | Show unpushed, staged, or modified LFS files. |
| `ls-files` | List LFS files in a commit. |
| `fetch` | Download LFS content without touching files. |
| `checkout` | Fill working files from downloaded content. |
| `pull` | `fetch` then `checkout`. |
| `push` | Upload LFS content. |
| `clone` | Old LFS clone; plain `git clone` is as fast now. |
| `prune` | Delete old local LFS content. |
| `fsck` | Check LFS content and pointers. |
| `migrate` | Rewrite history into or out of LFS, or report large files. |
| `lock` | Lock a file on the server. |
| `unlock` | Release a lock. |
| `locks` | List locks. |
| `dedup` | Make working files copy-on-write clones of stored content. |
| `env` | Print LFS settings and paths. |
| `ext` | Show LFS extensions. |
| `logs` | Show crash logs. |
| `update` | Update LFS hooks. |
| `version` | Print git-lfs version. |
| `completion` | Print a shell completion script. |
| `pointer` | Build or check pointer files. |
| `merge-driver` | Merge driver for LFS text files. |
| `clean` | Internal: filter, content to pointer. |
| `smudge` | Internal: filter, pointer to content. |
| `filter-process` | Internal: long-running clean and smudge filter. |
| `pre-push` | Internal: hook that uploads LFS content. |
| `post-checkout` | Internal: hook making lockable files read-only. |
| `post-commit` | Internal: hook making lockable files read-only. |
| `post-merge` | Internal: hook making lockable files read-only. |
| `standalone-file` | Internal: transfer adapter for `file:///` URLs. |

### git lfs install

Sets `lfs` filters in global config (if unset). Inside a repo, also installs the hooks (to `core.hooksPath` if set).

| Flag | What it does |
|------|--------------|
| `-f`, `--force` | Overwrite existing `lfs` filter settings. |
| `-l`, `--local` | Use repo config. |
| `-w`, `--worktree` | Use worktree config. Needs git 2.20+, and `extensions.worktreeConfig` with several worktrees. |
| `--system` | Use system config. |
| `--file=<file>` | Use this config file. |
| `-m`, `--manual` | Print how to add LFS to your own hooks instead. |
| `-s`, `--skip-smudge` | Don't download on checkout, clone, or pull; run `git lfs pull` yourself. |
| `--skip-repo` | Filters only; no repo hooks. |

### git lfs uninstall

Removes `lfs` filters from global config and LFS hooks from the repo.

| Flag | What it does |
|------|--------------|
| `-l`, `--local` | Use repo config. |
| `-w`, `--worktree` | Use worktree config. |
| `--system` | Use system config. |
| `--file=<file>` | Use this config file. |
| `--skip-repo` | Leave repo hooks and config alone. |

### git lfs track

`git lfs track [<flags>] [<pattern>...]`. Adds patterns to `.gitattributes` (commit it). Gitignore rules: `*`, `?`, `[ ]` are wildcards unless `\`-escaped or `--filename`. No patterns: list them.

| Flag | What it does |
|------|--------------|
| `-v`, `--verbose` | Log touched files. |
| `-d`, `--dry-run` | Show what would change. |
| `--filename` | Arguments are literal file names. |
| `-l`, `--lockable` | Mark lockable: read-only unless you hold the lock. |
| `--not-lockable` | Remove the lockable mark. |
| `--no-excluded` | When listing, hide excluded patterns. |
| `--no-modify-attrs` | Don't edit `.gitattributes`; just mark matches changed so `git add` re-adds them via LFS. |
| `--json` | List patterns as JSON. Not with pattern arguments. `-j` short form (git-lfs 3.8+). |

### git lfs untrack

`git lfs untrack <pattern>...`. Removes patterns from `.gitattributes`. No flags.

### git lfs status

Lists unpushed LFS files, staged and unstaged changes. Not in bare repos.

| Flag | What it does |
|------|--------------|
| `-p`, `--porcelain` | Script output. |
| `-j`, `--json` | Stable JSON output. |

### git lfs ls-files

`git lfs ls-files [<flags>] [<ref> [<ref>]]`. LFS files in `<ref>` (default current branch). Two refs: files changed between them (not deletions). `*` = content local; `-` = pointer only.

| Flag | What it does |
|------|--------------|
| `-l`, `--long` | Full 64-character IDs. |
| `-s`, `--size` | Show sizes. |
| `-d`, `--debug` | Show all details (format unstable). |
| `-a`, `--all` | Search all history. |
| `--deleted` | Include deleted files. |
| `-I`, `--include=<paths>` | Only these paths. |
| `-X`, `--exclude=<paths>` | Skip these paths. |
| `-n`, `--name-only` | Names only. |
| `--json` | JSON output (`--debug` wins; `--long`, `--size`, `--name-only` ignored). `-j` short form (git-lfs 3.7+). |

### git lfs fetch

`git lfs fetch [<flags>] [<remote> [<ref>...]]`. Download content for the current branch (or refs) from the default remote (tracked, else `origin`). Files untouched; `checkout` after.

| Flag | What it does |
|------|--------------|
| `-I`, `--include=<paths>` | Only these paths. |
| `-X`, `--exclude=<paths>` | Skip these paths. |
| `-r`, `--recent` | Also recent branches and commits (`lfs.fetchrecent*`; default: branches with commits in the last 7 days). Always on with `lfs.fetchrecentalways`. |
| `-a`, `--all` | Every commit reachable from the refs (or all refs). For backups. Ignores include/exclude; not with `--recent`, `-I`, `-X`. |
| `-p`, `--prune` | Run `git lfs prune` after. |
| `--stdin` | Refs from stdin. (git-lfs 3.8+) |
| `--refetch` | Re-download existing content. (git-lfs 3.7+) |
| `-d`, `--dry-run` | Print what would download. (git-lfs 3.7+) |
| `-j`, `--json` | Transfer details as JSON. (git-lfs 3.7+) |

### git lfs checkout

`git lfs checkout [<pattern>...]` or `git lfs checkout --to <file> (--base | --ours | --theirs) <conflicted-path>`. Replace pointer and missing files with downloaded content. Never overwrites modified files; never downloads.

| Flag | What it does |
|------|--------------|
| `--to <path>` | Merge conflict: write one side of an LFS file to `<path>` (may be outside the repo). |
| `--base` | With `--to`: merge base version. |
| `--ours` | With `--to`: current branch's version. |
| `--theirs` | With `--to`: other branch's version. |

### git lfs pull

`git lfs pull [<flags>] [<remote>]`. `git lfs fetch` plus `git lfs checkout` for the current branch.

| Flag | What it does |
|------|--------------|
| `-I`, `--include=<paths>` | Only these paths. |
| `-X`, `--exclude=<paths>` | Skip these paths. |

### git lfs push

`git lfs push [<flags>] <remote> [<ref>...]` or `git lfs push --object-id <remote> <oid>...`. Upload content not already referenced by the remote's refs. Usually done by the `pre-push` hook.

| Flag | What it does |
|------|--------------|
| `-d`, `--dry-run` | Print what would upload. |
| `-a`, `--all` | Content for every commit reachable from the refs (or all local branches and tags). |
| `-o`, `--object-id` | Arguments are LFS object IDs. |
| `--stdin` | Refs (or IDs) from stdin. |

### git lfs clone

`git lfs clone [<git clone flags>] <repo> [<dir>]`. Clone without LFS downloads, then one batched `git lfs pull`, and install hooks (in `--separate-git-dir` if given). Obsolete: `git clone` is as fast.

| Flag | What it does |
|------|--------------|
| Any [`git clone`](clone.md) flag | Passed to `git clone`. |
| `-I`, `--include=<paths>` | Only download these paths. |
| `-X`, `--exclude=<paths>` | Don't download these paths. |
| `--skip-repo` | Don't install hooks. |

### git lfs prune

Delete local content not needed by: the checkout, stashes, other worktrees, recent refs (fetch-recent settings plus `lfs.pruneoffsetdays`, default 3), or unpushed commits. "Pushed" means on `origin` (or `lfs.pruneremotetocheck`); without it nothing is pruned. Reflogs don't count. `lfs.fetchexclude` paths are deleted unless in a stash or unpushed commit. Unsafe with a shared `lfs.storage`.

| Flag | What it does |
|------|--------------|
| `-d`, `--dry-run` | Report only. |
| `-v`, `--verbose` | List each file. |
| `-f`, `--force` | Delete all but unpushed content, even what the checkout needs. Implies `--recent`. |
| `--recent` | Also delete content kept only for being recent. |
| `-c`, `--verify-remote` | Confirm each file is on the server first (git-lfs 3.5+: reachable files only). Default `lfs.pruneverifyremotealways`. |
| `--no-verify-remote` | Don't verify. |
| `--verify-unreachable` | With `--verify-remote`: also verify unreachable content. (git-lfs 3.5+) |
| `--no-verify-unreachable` | Override `lfs.pruneverifyunreachablealways`. (git-lfs 3.5+) |
| `--when-unverified=<halt\|continue>` | Unconfirmed file: stop (default) or delete the confirmed ones anyway. (git-lfs 3.5+) |

### git lfs fsck

`git lfs fsck [<flags>] [<commit> | <A>..<B>]`. Check LFS files in `HEAD` (plus index for objects), a commit, or a range. Corrupt content moves to `.git/lfs/bad`. Skips `lfs.fetchexclude` paths. Default: both checks.

| Flag | What it does |
|------|--------------|
| `--objects` | Objects exist and match their hash. |
| `--pointers` | Pointers are canonical; files that should be in LFS are. |
| `-d`, `--dry-run` | Don't move bad files. (documented in git-lfs 3.8+; accepted by 3.4.1) |

### git lfs migrate

`git lfs migrate <info | import | export> [<flags>] [--] [<branch>...]`. Rewrite history (new IDs) to convert files to or from LFS, or summarize sizes by type. Default: current branch, unpushed commits. Rewrites local branches only; force-push after. Commit or stash first. Edits `.gitattributes`. `-I`/`-X` use `.gitattributes` matching here (`dir/*` recursive). `^<branch>` excludes.

Shared flags:

| Flag | What it does |
|------|--------------|
| `-I`, `--include=<paths>` | Only matching files. |
| `-X`, `--exclude=<paths>` | Skip matching files. |
| `--include-ref=<ref>` | Include commits reachable from `<ref>`. Repeatable. Current branch no longer implied. |
| `--exclude-ref=<ref>` | Exclude commits reachable from `<ref>`. Repeatable. |
| `--everything` | All local and remote refs (only local rewritten). Case-only ref differences clash on case-insensitive file systems. |
| `--skip-fetch` | Don't refresh remote refs first. |
| `--yes` | Answer yes to prompts (overwriting working-tree changes). Can lose data. |

#### info

| Flag | What it does |
|------|--------------|
| `--above=<size>` | Only files bigger than `<size>` (`1b`, `20 MB`, `3 TiB`, ...). |
| `--top=<n>` | Top `<n>` types (default 5). |
| `--unit=<unit>` | `b`, `kib`, `mib`, `gib`, `tib`, `pib` or `kb`, `mb`, `gb`, `tb`, `pb`. Default auto. |
| `--pointers=<follow\|no-follow\|ignore>` | Existing pointers: `follow` (default) totals real objects separately; `ignore` skips; `no-follow` counts pointer files. |
| `--fixup` | Only files `.gitattributes` says belong in LFS but aren't. Not with `-I`/`-X`; implies `--pointers=ignore`. |

#### import

Convert matches to LFS and add patterns to `.gitattributes`. Without `-I` or `--fixup`, converts every file type.

| Flag | What it does |
|------|--------------|
| `--verbose` | Print each converted commit and file. |
| `--above=<size>` | Only files bigger than `<size>`. Not with `-I`, `-X`, `--fixup`. |
| `--object-map=<path>` | Write `OLD-SHA,NEW-SHA` CSV. |
| `--fixup` | Convert files that belong in LFS per `.gitattributes`. Not with `-I`/`-X`. |
| `--no-rewrite` | Convert named files (already in `.gitattributes`) in one new commit, no history rewrite. Other migrate flags ignored. |
| `-m`, `--message=<msg>` | With `--no-rewrite`: commit message. |

#### export

Convert matching LFS files back to normal files. `-I` required. Adds "don't track" entries to `.gitattributes`.

| Flag | What it does |
|------|--------------|
| `--verbose` | Print each converted commit and file. |
| `--object-map=<path>` | Write `OLD-SHA,NEW-SHA` CSV. |
| `--remote=<remote>` | Fetch missing content from here (default `origin`). |

### git lfs lock

`git lfs lock [<flags>] <path>`. Lock a file on the server so others' pushes changing it are refused (see `lfs.<url>.locksverify`). File must exist locally.

| Flag | What it does |
|------|--------------|
| `-r`, `--remote=<name>` | Server. Ignored if `lfs.url` is set. |
| `--json` | JSON output. `-j` short form (git-lfs 3.7+). |

### git lfs unlock

`git lfs unlock [<flags>] (<path> | --id=<id>)`. File must exist and be unmodified unless `--force`.

| Flag | What it does |
|------|--------------|
| `-r`, `--remote=<name>` | Server. Ignored if `lfs.url` is set. |
| `-f`, `--force` | Unlock even if someone else owns it; skip local checks. |
| `-i`, `--id=<id>` | Pick lock by ID. |
| `--json` | JSON output. `-j` short form (git-lfs 3.7+). |

### git lfs locks

Lists server locks.

| Flag | What it does |
|------|--------------|
| `-r`, `--remote=<name>` | Server. Ignored if `lfs.url` is set. |
| `-i`, `--id=<id>` | Only this lock ID. |
| `-p`, `--path=<path>` | Only this path. |
| `-l`, `--limit=<n>` | At most `<n>` results. |
| `--local` | Your locks from the local cache; no server call. |
| `--cached` | Everyone's locks from the last server call. May be stale. |
| `--verify` | Ask the server for owners; mark yours `O`. |
| `--json` | JSON output. `-j` short form (git-lfs 3.7+). |

### git lfs dedup

Make working LFS files copy-on-write clones of stored copies. Fails without copy-on-write support or with any LFS extension configured.

| Flag | What it does |
|------|--------------|
| `-t`, `--test` | Only check copy-on-write support. (documented in git-lfs 3.8+; accepted by 3.4.1) |

### git lfs env

Prints version, endpoints, paths, settings. No flags.

### git lfs ext

`git lfs ext [list [<name>...]]`. Shows LFS extensions (programs transforming content in clean and smudge). No flags.

### git lfs logs

`git lfs logs [<file> | last | clear | boomtown]`. None: list logs in `.git/lfs/logs`. `<file>` or `last`: show one. `clear`: delete all. `boomtown`: test crash. No flags.

### git lfs update

Update repo LFS hooks.

| Flag | What it does |
|------|--------------|
| `-m`, `--manual` | Print how to merge LFS into your hooks. |
| `-f`, `--force` | Overwrite existing hooks. |

### git lfs version

Prints the git-lfs version. No flags.

### git lfs completion

`git lfs completion <bash | fish | zsh>`. Print a completion script. Bash needs `bash-completion` and git's completion; zsh needs `compinit`. E.g. `source <(git lfs completion bash)`; `git lfs completion zsh > "${fpath[1]}/_git-lfs"`. No flags.

### git lfs pointer

Build a file's pointer and optionally compare it, or check data is a valid pointer.

| Flag | What it does |
|------|--------------|
| `-f`, `--file=<path>` | File to build a pointer from (or check with `--check`). |
| `-p`, `--pointer=<path>` | Pointer to compare with. |
| `--stdin` | Read the pointer to compare or check from stdin. |
| `--check` | Exit 0 if `--file` or `--stdin` (exactly one) is a valid pointer, else 1. |
| `--[no-]strict` | With `--check`: require canonical form, else exit 2. Default off. |
| `--no-extensions` | Omit extension lines. (git-lfs 3.8+) |

### git lfs merge-driver

Merge LFS text files with `git merge-file` or your program. Off by default; enable for `merge=lfs-text` files:

`git config merge.lfs-text.driver 'git lfs merge-driver --ancestor %O --current %A --other %B --marker-size %L --output %A'`

| Flag | What it does |
|------|--------------|
| `--ancestor <path>` | Base version. |
| `--current <path>` | Our version. |
| `--other <path>` | Their version. |
| `--output <path>` | Result file. |
| `--marker-size <n>` | Conflict marker length. |
| `--program <cmd>` | Merge command instead of `git merge-file`. Placeholders: `%A` current, `%B` other, `%O` ancestor, `%D` destination, `%L` marker size (`%%` in git config). Exit 0 if clean, non-zero on conflict. |

### git lfs clean

Internal: clean filter. Content on stdin is stored; pointer to stdout. Preview with `git lfs pointer`. `git lfs clean <path>`. No flags.

### git lfs smudge

Internal: smudge filter. Pointer on stdin, content out (downloading if needed). Paths excluded by `lfs.fetchinclude`/`lfs.fetchexclude` stay pointers. `git lfs smudge [<path>]` (path only labels progress).

| Flag | What it does |
|------|--------------|
| `-s`, `--skip` | Output the pointer; no download. Same as `GIT_LFS_SKIP_SMUDGE=1`. |

### git lfs filter-process

Internal: long-running clean and smudge filter. Respects `lfs.fetchinclude`/`lfs.fetchexclude`.

| Flag | What it does |
|------|--------------|
| `-s`, `--skip` | No downloads on checkout. Same as `GIT_LFS_SKIP_SMUDGE=1`. |

### git lfs pre-push

Internal: `pre-push` hook. `git lfs pre-push <remote> [<url>]`, reading `<local-ref> <local-sha> <remote-ref> <remote-sha>` lines from stdin. Uploads content in pushed commits; nothing for deletions. `GIT_LFS_SKIP_PUSH=1` disables it.

| Flag | What it does |
|------|--------------|
| `-d`, `--dry-run` | Print what would upload. |

### git lfs post-checkout

Internal: `post-checkout` hook. `git lfs post-checkout <old-rev> <new-rev> <is-branch-checkout>`. Makes lockable files read-only unless you hold the lock. No flags.

### git lfs post-commit

Internal: `post-commit` hook. Makes lockable files changed in `HEAD` read-only unless you hold the lock. No flags.

### git lfs post-merge

Internal: `post-merge` hook. `git lfs post-merge <is-squash>`. Makes lockable files read-only unless you hold the lock. No flags.

### git lfs standalone-file

Internal: transfer adapter for `file:///` URLs (local folders as server), JSON over stdin/stdout. Automatic. No flags.
