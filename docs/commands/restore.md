# git restore

Restore files: discard unstaged edits (from the index), unstage (from `HEAD`), or take a file from any commit. Tracked files missing from the source are deleted. Doesn't move branches; see [`reset`](reset.md).

**Usage:** `git restore [<flags>] [--source=<tree>] [--staged] [--worktree] [--] <pathspec>...`, `git restore [<flags>] --pathspec-from-file=<file> [--pathspec-file-nul]`, `git restore -p [<flags>] [--] [<pathspec>...]`

Examples:
- Discard unstaged changes: `git restore file.c` (`.` here, `:/` whole repo).
- Unstage: `git restore --staged file.c`.
- File from 2 commits ago: `git restore --source=HEAD~2 Makefile`.
- Reset staged and working copy: `git restore -s@ -SW file.c`.
- Quote globs (`'*.c'`) to match tracked files, including deleted ones.

**Warning:** restoring the working tree discards uncommitted edits with no undo.

## Flags

| Flag | What it does |
|------|--------------|
| `-s`, `--[no-]source=<tree>` | Source commit, branch, or tag. Default: `HEAD` with `--staged`, else the index. `A...B` = merge base. |
| `-W`, `--[no-]worktree` | Restore working files (default if neither `-W` nor `-S`). |
| `-S`, `--[no-]staged` | Restore the index. With `-W`: both. |
| `-p`, `--[no-]patch` | Pick hunks. Keys: see [`add`](add.md). |
| `-U`, `--unified=<n>` | With `-p`: context lines (default `diff.context` or 3). |
| `--inter-hunk-context=<n>` | With `-p`: join hunks up to `<n>` lines apart (default `diff.interHunkContext` or 0). |
| `-q`, `--[no-]quiet` | No feedback. Implies `--no-progress`. |
| `--[no-]progress` | Force progress off a terminal. |
| `-2`, `--ours` | Conflicted files: our version (stage 2). Not with `--source`. Swapped in rebase. |
| `-3`, `--theirs` | Conflicted files: their version (stage 3). Not with `--source`. Swapped in rebase. |
| `-m`, `--[no-]merge` | Conflicted files: recreate conflict markers. Not with `--source`. |
| `--[no-]conflict=<style>` | `--merge` with style `merge` (default), `diff3`, or `zdiff3`. Overrides `merge.conflictStyle`. |
| `--[no-]ignore-unmerged` | Skip conflicted files instead of aborting (without `--ours`/`--theirs`/`--merge`/`--conflict`). |
| `--[no-]ignore-skip-worktree-bits` | Sparse checkout: restore all matches, not just in-pattern ones. |
| `--[no-]recurse-submodules[=<checkout>]` | Also restore active submodules to the recorded commit, discarding their changes (detached `HEAD`). Default: untouched. |
| `--[no-]overlay` | Overlay: never delete. Default no-overlay deletes tracked files missing from `--source`. |
| `--[no-]pathspec-from-file=<file>` | Read pathspecs from `<file>` (`-` = stdin). |
| `--[no-]pathspec-file-nul` | With `--pathspec-from-file`: NUL-separated, literal. |
| `--` | Rest are paths. |
