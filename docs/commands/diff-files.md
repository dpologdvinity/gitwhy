# git diff-files

Plumbing: compare working files with the index. Default output `--raw`. Humans want `git diff`.

**Usage:** `git diff-files [-q] [-0 | -1 | -2 | -3 | -c | --cc] [<flags>] [<path>...]`

## Flags

Also takes all [diff options](../shared/diff-options.md).

| Flag | What it does |
|------|--------------|
| `-1`, `--base` | Conflicted files: compare with the common ancestor (stage 1). Hides resolved files. |
| `-2`, `--ours` | Conflicted files: compare with your version (stage 2). Default. |
| `-3`, `--theirs` | Conflicted files: compare with their version (stage 3). |
| `-0` | Conflicted files: print only "Unmerged". |
| `-c`, `--cc` | Conflicted files: combined diff of ours, theirs, and working file. |
| `-q` | Don't complain about files missing from disk. |
| `<path>...` | Limit to these paths. |
