# git check-attr

Show which `.gitattributes` settings apply to files: `unspecified`, `unset`, `set`, or a value.

Output: `<path>: <attribute>: <info>` per line.

**Usage:** `git check-attr [--source <tree-ish>] [-a | --all | <attr>...] [--] <pathname>...`
or `git check-attr --stdin [-z] [--source <tree-ish>] [-a | --all | <attr>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `-a`, `--[no-]all` | Show every attribute set on each path. |
| `--[no-]cached` | Read `.gitattributes` from the index only. |
| `--[no-]stdin` | Read paths from stdin, one per line. |
| `-z` | NUL-separated output (`<path> NUL <attribute> NUL <info> NUL`); NUL-separated input with `--stdin`. |
| `--[no-]source=<tree-ish>` | Use attributes from this commit, branch, or tag. |
| `--` | Separates attributes from paths. Without it (or `--stdin`/`--all`), the first argument is the attribute. |
| `<attr>...` | Attributes to check. |
| `<pathname>...` | Files to check. |
