# git merge-base

Find the best common ancestor of commits (where histories split; a merge's starting point). Criss-cross merges can have several; without `--all`, one is picked arbitrarily.

With 3+ commits: base of the first and a hypothetical merge of the rest, so it may not be an ancestor of all; use `--octopus` for that.

**Usage:** `git merge-base [-a | --all] <commit> <commit>...`, `git merge-base [-a | --all] --octopus <commit>...`, `git merge-base --is-ancestor <commit> <commit>`, `git merge-base --independent <commit>...`, `git merge-base --fork-point <ref> [<commit>]`

## Flags

| Flag | What it does |
|------|--------------|
| `<commit>...` | Commits or branches to compare. |
| `-a`, `--[no-]all` | Print all best common ancestors. |
| `--octopus` | Best ancestor common to *all* commits. Same as `git show-branch --merge-base`. |
| `--independent` | Print the given commits not reachable from any other given one. Same as `git show-branch --independent`. |
| `--is-ancestor` | Exit 0 if first commit is an ancestor of second, 1 if not, other on error. Prints nothing. |
| `--fork-point` | Where `<commit>` (default `HEAD`) branched off `<ref>`, using `<ref>`'s reflog. See below. |

## Fork-point mode

If upstream was force-pushed after you branched, the plain merge base is too old and `git rebase` would replay commits upstream dropped. `--fork-point` uses `<ref>`'s reflog to find the commit you actually built on:

```sh
fork_point=$(git merge-base --fork-point origin/main topic)
git rebase --onto origin/main $fork_point topic
```

Fails if that reflog entry expired, or if you branched from an older commit rather than `<ref>`'s tip at the time.
