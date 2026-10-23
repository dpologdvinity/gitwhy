# git merge-resolve

Internal: the "resolve" strategy, run by `git merge -s resolve`. Simple three-way merge of two heads; handles only one merge base cleanly.

**Usage:** `git merge-resolve <base>... -- <head> <remote>`

## Flags

None. Takes merge base(s), `--`, current head, and the head being merged.
