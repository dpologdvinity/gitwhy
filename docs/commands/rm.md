# git rm

Delete tracked files from the working tree and index, staging the removal. Plain `rm` deletes from disk only; `--cached` stops tracking but keeps the file.

Refuses unless the file matches the last commit with nothing staged (`-f` overrides). With `--cached`, staged version must match the last commit or the disk file. With sparse checkout, only paths inside the sparse area unless `--sparse`.

**Usage:** `git rm [<flags>] [--] [<pathspec>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<pathspec>...` | Files to remove. Folders need `-r`. Globs cross folders: `'d*'` also matches `d2/`, unlike `'d/*'`. Quote globs (`git rm 'docs/*.txt'`) so git expands them, including subfolders. |
| `-f`, `--[no-]force` | Remove even with uncommitted changes. Those changes are lost. |
| `-n`, `--[no-]dry-run` | Show what would be removed. |
| `-r` | Allow removing folders. |
| `--[no-]cached` | Remove from index only; keep file on disk. |
| `--[no-]ignore-unmatch` | Exit 0 if nothing matched. |
| `--[no-]sparse` | Allow paths outside the sparse-checkout area. |
| `-q`, `--[no-]quiet` | Don't print `rm '<file>'` lines. |
| `--[no-]pathspec-from-file=<file>` | Read pathspecs from `<file>` (`-` = stdin), one per line. Quoting per `core.quotePath`. |
| `--[no-]pathspec-file-nul` | With `--pathspec-from-file`: NUL-separated, literal. |
| `--` | Everything after is a file name. |

## Files you already deleted by hand

No flag stages "whatever is gone from disk". Instead:

- `git commit -a` or `git add -u`: stage all deletions and changes of tracked files.
- `git add -A`: also adds new files. Good for replacing a vendor drop.
- Only deletions: `git diff --name-only --diff-filter=D -z | xargs -0 git rm --cached`

## Submodules

Removes the submodule's working folder if its repo lives in the parent's `.git` (git 1.7.8+ clones). A submodule with its own `.git` folder has it moved into the parent's `.git` first, keeping history. Also removes and stages its `.gitmodules` section (not with `--cached` or `-n`). Refuses if the submodule has changes: different `HEAD`, modified tracked files, or untracked non-ignored files. `-f` removes it anyway, deleting those changes. To remove just the checkout, use `git submodule deinit`.
