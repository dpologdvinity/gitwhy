# git merge-index

Plumbing: run `<merge-program>` (usually `git merge-one-file`) on each unmerged index file. Stops at the first failure by default.

The program gets 7 args: base, ours, theirs object IDs (empty if missing), file name, then the three modes. Base comes first, unlike RCS `merge`.

**Usage:** `git merge-index [-o] [-q] <merge-program> (-a | [--] <file>...)`

## Flags

| Flag | What it does |
|------|--------------|
| `-a` | Merge every unmerged file. |
| `-o` | Continue past failures; report at the end. |
| `-q` | Don't complain when the program fails. |
| `--` | Rest are file names. |
