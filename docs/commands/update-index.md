# git update-index

Low-level: edit the index directly (stage files, set "assume unchanged" and "skip worktree" bits). For normal staging use [`git add`](add.md).

Flags apply to files after them, so order matters. New and deleted files are ignored unless `--add`/`--remove`.

**Usage:** `git update-index [<flags>] [--] [<file>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<file>...` | Files to act on. Paths with `.` parts, trailing `/`, or `//` are skipped. |
| `--[no-]add` | Add files not yet in the index. |
| `--[no-]remove` | Remove indexed files gone from disk. |
| `--[no-]force-remove` | Remove from the index even if still on disk. Implies `--remove`. |
| `--[no-]replace` | Allow a file to replace a same-named directory or vice versa; removes the conflicting entries with a warning. |
| `--refresh` | Re-check cached stat info so unchanged files stop looking changed. Doesn't stage content. |
| `--really-refresh` | `--refresh`, including "assume unchanged" files. |
| `-q` | With `--refresh`: continue instead of failing when updates are needed. |
| `--[no-]ignore-submodules` | With `--refresh` (put it first): skip submodules. |
| `--[no-]unmerged` | With `--refresh`: continue past unmerged entries. |
| `--[no-]ignore-missing` | With `--refresh`: skip missing files. |
| `--cacheinfo <mode>,<object>,<path>` | Insert an entry directly; file need not exist. Old form: three arguments. |
| `--index-info` | Read entries from stdin (see below). |
| `--chmod=(+\|-)x` | Set or clear the executable bit in the index. Useful with `core.filemode=false`. |
| `--[no-]assume-unchanged` | Set/clear "assume unchanged": git stops checking the file. Speed trick, not a way to ignore changes. |
| `--[no-]skip-worktree` | Set/clear "skip worktree": git won't write the file and treats it as unchanged if missing. Prefer `git sparse-checkout`. |
| `--[no-]ignore-skip-worktree-entries` | With `--remove`: keep skip-worktree entries. |
| `--[no-]fsmonitor-valid` | Set/clear the "fsmonitor valid" bit. |
| `-g`, `--again` | Re-run on paths whose index entry differs from `HEAD`. |
| `--unresolve` | Restore a file's conflicted state during a merge. |
| `--clear-resolve-undo` | Forget the data `--unresolve` needs. |
| `--[no-]info-only` | Following files: record IDs only, don't store contents. |
| `--stdin` | Read paths from stdin. |
| `-z` | With `--stdin` or `--index-info`: NUL-separated. |
| `--[no-]verbose` | Report adds and removes. |
| `--[no-]index-version=<n>` | Index format `2`, `3`, or `4` (compresses paths, 30-50% smaller). Default 2 or 3. |
| `--[no-]show-index-version` | Print index format version. |
| `--[no-]split-index` | Toggle split-index (below). Repeat while on to fold changes back. Overrides `core.splitIndex` until next read. |
| `--[no-]untracked-cache` | Toggle the untracked cache. Test with `--test-untracked-cache` first. Prefer `core.untrackedCache`. |
| `--[no-]test-untracked-cache` | Test file system support: prints `OK` and exits 0, or explains and exits 1. |
| `--[no-]force-untracked-cache` | Same as `--untracked-cache`. |
| `--[no-]fsmonitor` | Toggle the file system monitor. Prefer `core.fsmonitor`. |
| `--[no-]force-write-index` | Write the index even if unchanged. |
| `--` | Rest are file names. |

## --cacheinfo and --info-only

Both skip the object database. `--cacheinfo`: object exists, file not on disk. `--info-only`: file on disk, don't store it.

## --index-info input

One entry per line:

1. `<mode> <type> <id>TAB<path>`: `git ls-tree` output.
2. `<mode> <id> <stage>TAB<path>`: `git ls-files --stage` output; can set stages 1-3.
3. `<mode> <id>TAB<path>`: old format.

Mode `0` removes the path. For conflict stages, remove the path first, then add stages.

## Assume unchanged vs. skip worktree

| | Assume unchanged | Skip worktree |
|---|------------------|---------------|
| Purpose | Skip slow checks; you promise not to edit. | Sparse checkout: file left off disk. |
| File on disk | Present. | Usually absent. |
| If you edit it | Clear the bit so git notices. | Cleared automatically in sparse checkouts if the file appears. |
| See which files | `git ls-files -v` | `git ls-files -v` |

Neither reliably ignores changes to tracked files. For local config, commit a sample and ignore the real file.

With `core.ignorestat=true`, files git writes get "assume unchanged" automatically.

## Split index

For huge indexes: changes go to `.git/index`, the rest stays in `.git/sharedindex.<id>`, speeding writes. Folded back at `splitIndex.maxPercentChange`; old shared files expire per `splitIndex.sharedIndexExpire`.

## Untracked cache and fsmonitor

The untracked cache stores directory timestamps so `git status` skips unchanged directories. Needs a file system that updates directory timestamps. If `git status` misbehaves, run `git -c core.untrackedCache=false status` once.

The file system monitor (`git fsmonitor--daemon` or Watchman) tells git which files changed.

Flags apply immediately; config applies on the next index read.
