# git diff

Show changes between working files, the index, commits, blobs, or any two files on disk.

**Usage:**
`git diff [<flags>] [--] [<path>...]` working files vs. index (unstaged changes)
`git diff [<flags>] --cached [--merge-base] [<commit>] [--] [<path>...]` index vs. `<commit>` (default `HEAD`): what you'd commit
`git diff [<flags>] [--merge-base] <commit> [--] [<path>...]` working files vs. `<commit>`
`git diff [<flags>] [--merge-base] <commit> <commit> [--] [<path>...]` two commits (same as `A..B`)
`git diff [<flags>] <commit>...<commit> [--] [<path>...]` second side's changes since their merge base
`git diff [<flags>] <merge> <parent>... [--] [<path>...]` merge vs. its parents, combined diff
`git diff [<flags>] <blob> <blob>` two blobs
`git diff [<flags>] --no-index [--] <path> <path> [<pathspec>...]` two files or folders on disk, repo or not

Most `<commit>` spots take any tree, e.g. `AUTO_MERGE` to see your conflict fixes so far. `<path>` limits the diff.

## Flags

Also takes all [diff options](../shared/diff-options.md).

| Flag | What it does |
|------|--------------|
| `--cached`, `--staged` | Index vs. a commit (default `HEAD`). |
| `--merge-base` | Use the merge base with `HEAD` (or of the two commits) as the old side. |
| `--no-index` | Compare two paths outside git. Automatic outside a repo. Implies `--exit-code`. |
| `-1`, `--base` | Conflict: compare with the base version (stage 1). |
| `-2`, `--ours` | Conflict: compare with our version (stage 2). |
| `-3`, `--theirs` | Conflict: compare with their version (stage 3). |
| `-0` | Working vs. index: show conflicted files as just "Unmerged". |
| `-c` | Merges: combined diff of each parent against the result. |
| `--cc` | Like `-c`, hiding hunks taken unchanged from one parent. Default for merges. |
| `--combined-all-paths` | With `-c`/`--cc`: show each parent's file name (renames). |
