# git merge-recursive-theirs

Internal command: [`git merge-recursive`](merge-recursive.md) with `--theirs` (`git merge -X theirs`). Conflicts take the other side's version; non-conflicting changes from both sides are kept. No "theirs" strategy discards your side entirely.

**Usage:** `git merge-recursive-theirs [--<option>...] <base>... -- <head> <remote>...`

## Flags

Same `--<option>` flags as [`git merge-recursive`](merge-recursive.md).
