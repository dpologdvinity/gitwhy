# git mktree

Plumbing: build a tree object from `git ls-tree`-format lines on stdin; print its ID. Any order; one level only.

**Usage:** `git mktree [-z] [--missing] [--batch]`

## Flags

| Flag | What it does |
|------|--------------|
| `-z` | NUL-terminated input lines. |
| `--[no-]missing` | Don't check that listed objects exist. |
| `--[no-]batch` | Build several trees; blank line between each. One ID per tree. |
