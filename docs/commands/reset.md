# git reset

Two jobs: move the current branch to another commit (optionally resetting index and files), or unstage files (opposite of `git add`). To discard file changes use [`git restore`](restore.md); to undo shared commits use `git revert`.

**Usage:**
- `git reset [<mode>] [<commit>]`: move the branch to `<commit>` (default `HEAD`). Old tip saved in `ORIG_HEAD`.
- `git reset [-q] [<tree-ish>] [--] <pathspec>...`: reset these paths' staged versions to `<tree-ish>` (default `HEAD`). Same as `git restore --staged`.
- `git reset [-q] --pathspec-from-file=<file> [--pathspec-file-nul] [<tree-ish>]`: same, paths from a file.
- `git reset (-p | --patch) [<tree-ish>] [--] [<pathspec>...]`: pick hunks to unstage.

Common recipes:
- Undo last commit, keep changes staged: `git reset --soft HEAD^`.
- Squash the last 5 commits: `git reset --soft HEAD~5 && git commit`.
- Throw away the last 3 commits and all local changes: `git reset --hard HEAD~3` (not on shared commits).
- Undo a merge or pull: `git reset --hard ORIG_HEAD`, or `--merge ORIG_HEAD` to keep unrelated unstaged changes.
- Split the last commit: `git reset -N HEAD^`, then `git add -p` and commit in pieces.

## Modes

Default: `--mixed`.

| Flag | What it does |
|------|--------------|
| `--soft` | Move the branch only; changes stay staged. |
| `--mixed` | Move the branch and reset the index; changes become unstaged. |
| `--hard` | Move the branch; make index and files match. **Deletes uncommitted changes to tracked files** (including newly added ones) and may overwrite untracked ones. |
| `--merge` | Like `--hard`, keeping unstaged changes in untouched files. **Staged changes are lost.** Aborts if a changed file has unstaged edits. For backing out conflicted merges. |
| `--keep` | Move the branch, updating changed files but keeping local edits. Aborts if a changed file has local edits. |

## Flags

| Flag | What it does |
|------|--------------|
| `-q`, `--[no-]quiet` | Errors only. |
| `--[no-]refresh` | Refresh the index after a mixed reset (default). |
| `-N`, `--[no-]intent-to-add` | With `--mixed`: mark newly untracked paths intent-to-add, so `git diff` and `git add -p` see them. |
| `--[no-]recurse-submodules[=<reset>]` | Also reset active submodules to the recorded commit. |
| `-p`, `--[no-]patch` | Pick hunks to unstage. Keys: see [`add`](add.md). |
| `--[no-]auto-advance` | With `-p`: go to the next file after the last hunk (default). `--no-`: stay; use `<`/`>`. |
| `-U`, `--unified=<n>` | With `-p`: `<n>` context lines (default `diff.context` or 3). |
| `--inter-hunk-context=<n>` | With `-p`: join hunks up to `<n>` lines apart (default `diff.interHunkContext` or 0). |
| `--[no-]pathspec-from-file=<file>` | Pathspecs from `<file>` (`-` = stdin). |
| `--[no-]pathspec-file-nul` | With `--pathspec-from-file`: NUL-separated, literal. |
| `--` | Rest are paths. |
