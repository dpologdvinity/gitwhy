# git merge-recursive

Internal: `git merge` runs it for the `recursive` strategy, which since git 2.50 means `ort` (the default). Use [`git merge`](merge.md).

Takes merge base(s), `--`, the current commit, then commits to merge. Writes result to index and working tree. Exits 0 clean, 1 on conflicts.

**Usage:** `git merge-recursive [--<option>...] <base>... -- <head> <remote>...`

## Flags

Options go before `--`; same as `git merge -X<option>`.

| Flag | What it does |
|------|--------------|
| `--ours` | Conflicting lines: take ours. Other side's clean changes still merge. |
| `--theirs` | Conflicting lines: take theirs. |
| `--subtree[=<path>]` | Shift one tree to line up as a subdirectory of the other; `<path>` sets the prefix. |
| `--ignore-space-change` | Ignore whitespace amount changes. |
| `--ignore-all-space` | Ignore all whitespace. |
| `--ignore-space-at-eol` | Ignore whitespace at line ends. |
| `--ignore-cr-at-eol` | Ignore CR at line ends. |
| `--[no-]renormalize` | Normalize line endings and filters on all three sides first. |
| `--find-renames[=<n>]` | Detect renames at `<n>`% similarity. |
| `--rename-threshold=<n>` | Old name for `--find-renames=<n>`. |
| `--no-renames` | No rename detection. |
| `--patience` | Patience diff algorithm. |
| `--histogram` | Histogram diff algorithm. |
| `--diff-algorithm=<algo>` | `myers`, `minimal`, `patience`, or `histogram`. |
