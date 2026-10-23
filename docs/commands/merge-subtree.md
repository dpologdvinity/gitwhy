# git merge-subtree

Internal: the "subtree" merge strategy, run by `git merge -s subtree`. [`git merge-recursive`](merge-recursive.md) with `--subtree`: shifts one tree to line up as a subdirectory of the other. For a project kept in a subfolder of yours.

**Usage:** `git merge-subtree [--<option>...] <base>... -- <head> <remote>...`

## Flags

No man page. Takes merge base(s), `--`, the current head, then heads to merge. Options before `--` are [`git merge-recursive`](merge-recursive.md)'s; `--subtree=<path>` names the subfolder instead of guessing.
