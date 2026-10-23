# git diff-index

Plumbing: compare a tree (usually `HEAD`) with working files or the index. Default output `--raw`. Humans want `git diff`.

Without `--cached`, unstaged changes show an all-zero object ID (only timestamps were checked).

**Usage:** `git diff-index [-m] [--cached] [--merge-base] [<flags>] <tree-ish> [<path>...]`

## Flags

Also takes all [diff options](../shared/diff-options.md).

| Flag | What it does |
|------|--------------|
| `--cached` | Compare with the index only. `git diff-index --cached HEAD` = what you'd commit. |
| `--merge-base` | Compare from the merge base of `<tree-ish>` and `HEAD`. `<tree-ish>` must be a commit. |
| `-m` | Treat files missing from disk as unchanged, not deleted. |
| `<tree-ish>` | Tree or commit to compare against. |
| `<path>...` | Limit to these paths. |
