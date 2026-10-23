# git merge-octopus

Internal: the "octopus" merge strategy, run by `git merge -s octopus` (default when merging 2+ branches at once). Merges several heads into one commit with many parents. Refuses with staged changes, with only two heads, or if any head but the last conflicts.

**Usage:** `git merge-octopus <base>... -- <head> <remote> <remote>...`

## Flags

None. Takes merge base(s), `--`, the current head, then 2+ heads to merge.
