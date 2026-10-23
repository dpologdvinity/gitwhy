# git merge-recursive-ours

Internal: [`git merge-recursive`](merge-recursive.md) with `--ours`. Conflicting lines keep our side; other non-conflicting changes still merge. Unlike `git merge -s ours`, which ignores the other side.

**Usage:** `git merge-recursive-ours [--<option>...] <base>... -- <head> <remote>...`

## Flags

Same `--<option>` flags as [`git merge-recursive`](merge-recursive.md).
