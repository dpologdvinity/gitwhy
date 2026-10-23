# git rev-list

List commit IDs reachable from given commits, newest first. Engine behind `log`, `bisect`, packing; mostly for scripts.

**Usage:** `git rev-list [<flags>] <commit>... [--] [<path>...]`

`git rev-list A B ^C` = reachable from `A` or `B` but not `C`. `A..B` and `A...B` work too.

## Flags

Also takes all [revision options](../shared/revision-options.md): limiting (`--max-count`, `--since`, `--all`, ...), simplification, ordering, object traversal (`--objects`, `--filter`, ...), formatting (`--pretty`, `--graph`, ...).

| Flag | What it does |
|------|--------------|
| `--quiet` | Print nothing; exit code tells if the range is fully connected. |
| `--disk-usage[=human]` | Print total on-disk bytes of selected commits (or objects with `--objects`). `human`: `12.24 MiB` style. |
| `--use-bitmap-index` | Use pack bitmap for speed. With `--objects`, no paths printed. |
| `--progress=<header>` | Progress on stderr, labelled `<header>`. |
| `-z` | NUL-separated; extra info as `<token>=<value>`. Only with `--objects`, `--boundary`, `--missing`. |
| `--bisect` | Print only the commit about halfway between included and excluded. Uses `refs/bisect/bad`, `refs/bisect/good-*` if present. |
| `--bisect-vars` | Like `--bisect` but ignores `refs/bisect/`; prints shell variables (`bisect_rev`, `bisect_nr`, `bisect_good`, `bisect_bad`, `bisect_all`) for `eval`. |
| `--bisect-all` | All commits in range, best bisect midpoint first. Ignores `refs/bisect/`. Combines with `--bisect-vars`. |
| `--header` | Raw commit format, NUL-separated. |
| `--no-commit-header` | With custom `--format`, drop the `commit <id>` line. |
| `--commit-header` | Undo `--no-commit-header`. |
| `--timestamp` | Print raw commit timestamps. |
| `--count` | Print only the count. With `--left-right`: left and right counts. With `--cherry-mark`: also duplicates. |
| `--abbrev=<n>`, `--no-abbrev` | Minimum short-ID length / full IDs. |
