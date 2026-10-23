# git merge-ours

Internal command run by `git merge -s ours`. Merge that keeps the current branch's files unchanged and discards all changes from the other branches; the commit only records the merge. Unlike `-X ours` (see [`git merge-recursive-ours`](merge-recursive-ours.md)), which keeps the other side's non-conflicting changes. Fails if the index differs from `HEAD`.

**Usage:** `git merge-ours <base>... -- HEAD <remote>...`

## Flags

None. Arguments ignored; only checks the index matches `HEAD`.
