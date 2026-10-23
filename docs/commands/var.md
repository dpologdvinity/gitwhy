# git var

Print a computed git setting (editor, pager, author identity, ...). Exits 1 if it has no value.

**Usage:** `git var (-l | <variable>)`

Unreleased (on git `master`, not in 2.56.0): `git var [-z] (-l | <variable>...)`; several variables print as `NAME=value`, skipping empty ones.

## Flags

| Flag | What it does |
|------|--------------|
| `-l` | List logical variables, then config (deprecated; use `git config list`). |
| `-z` | NUL-terminate entries. With `-l` or several variables, newline splits name and value. (Unreleased: on git `master`, not in 2.56.0.) |

## Variables

| Variable | Value |
|----------|-------|
| `GIT_AUTHOR_IDENT` | Current author name, email, timestamp. |
| `GIT_COMMITTER_IDENT` | Current committer name, email, timestamp. |
| `GIT_EDITOR` | Editor: `$GIT_EDITOR`, `core.editor`, `$VISUAL`, `$EDITOR`, then default (usually `vi`). |
| `GIT_SEQUENCE_EDITOR` | `rebase -i` editor: `$GIT_SEQUENCE_EDITOR`, `sequence.editor`, then `GIT_EDITOR`. |
| `GIT_PAGER` | Pager: `$GIT_PAGER`, `core.pager`, `$PAGER`, then default (usually `less`). |
| `GIT_DEFAULT_BRANCH` | First branch name in new repos. |
| `GIT_SHELL_PATH` | POSIX shell path git uses. |
| `GIT_ATTR_SYSTEM` | System attributes file, if enabled. |
| `GIT_ATTR_GLOBAL` | Per-user attributes file. |
| `GIT_CONFIG_SYSTEM` | System config file, if enabled. |
| `GIT_CONFIG_GLOBAL` | Per-user config file(s). |

Path variables may print several lines, highest priority first, even for missing files.
