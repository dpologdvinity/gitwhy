# git merge-resolve

Internal: the "resolve" strategy, run by `git merge -s resolve`. Simple three-way merge of exactly two heads. Refuses with staged changes or no merge base.

**Usage:** `git merge-resolve <base>... -- <head> <remote>`

## Flags

None. Takes merge base(s), `--`, current head, and the head being merged.
