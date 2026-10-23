# git write-tree

Low-level: save the index (staging area) as a tree object and print its ID. Index must have no conflicts. `git commit` does this for you.

**Usage:** `git write-tree [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]missing-ok` | Don't check that referenced objects exist. |
| `--[no-]prefix=<prefix>/` | Write only the tree for subdirectory `<prefix>`. |
