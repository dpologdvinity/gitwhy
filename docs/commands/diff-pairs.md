# git diff-pairs

Plumbing: read NUL-separated raw file pairs on stdin (e.g. from `git diff-tree -z -r --raw`) and print their diffs, as a patch by default. Example: `git diff-tree -z -r -M A B | git diff-pairs -z`.

Output comes when stdin closes. A lone NUL in input flushes; output batches are NUL-separated. Input can't have pathspecs, tree entries, or short IDs.

**Usage:** `git diff-pairs -z [<flags>]`

## Flags

Also takes all [diff options](../shared/diff-options.md).

| Flag | What it does |
|------|--------------|
| `-z` | NUL-separated input and output. Required. |
