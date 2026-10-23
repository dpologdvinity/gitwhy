# git cherry

Show which of your commits are already upstream. For each commit not in `<upstream>`: `-` if upstream has an equivalent change (cherry-picked, rebased, applied patch), `+` if not. Compares diffs, ignoring whitespace and line numbers.

E.g. after `git fetch`, `git cherry -v` shows which sent patches were applied.

**Usage:** `git cherry [-v] [<upstream> [<head> [<limit>]]]`

## Flags

| Flag | What it does |
|------|--------------|
| `<upstream>` | Branch to search. Default: current branch's upstream. |
| `<head>` | Your branch. Default `HEAD`. |
| `<limit>` | Skip commits up to and including this one. |
| `-v`, `--[no-]verbose` | Show subject lines. |
| `--[no-]abbrev[=<n>]` | Shorten IDs to `<n>` chars. |
