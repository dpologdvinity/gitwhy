# git repo

Experimental: print repo facts (bare, shallow, hash type, paths) and size stats in script-friendly formats.

**Usage:** `git repo <subcommand> [<flags>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `info` | Print the requested keys' values, in order. `--keys` lists keys. |
| `structure` | Stats by type: ref counts, reachable object counts, uncompressed and on-disk sizes, largest objects. |

### git repo info

`git repo info [<flags>] [--all | <key>...]` or `git repo info --keys [<flags>]`

| Flag | What it does |
|------|--------------|
| `--all` | Print every key. |
| `--keys` | List key names, one per line. |
| `--format=<format>` | `lines` (default): `key=value`, odd characters quoted. `nul`: key, newline, value, NUL; unquoted. With `--keys`: name per line, or NUL after each. |
| `-z` | `--format=nul`. |

Keys:

| Key | Value |
|-----|-------|
| `layout.bare` | `true` if bare. |
| `layout.shallow` | `true` if shallow clone. |
| `object.format` | Hash algorithm (`sha1`, `sha256`). |
| `references.format` | Ref storage: `files` or `reftable`. |
| `path.gitdir.absolute` | (2.56+) Absolute `.git` directory path. |
| `path.gitdir.relative` | (2.56+) `.git` directory path relative to current directory. |
| `path.commondir.absolute` | (2.56+) Absolute path of the git directory shared by all worktrees. |
| `path.commondir.relative` | (2.56+) Shared git directory relative to current directory. |

### git repo structure

| Flag | What it does |
|------|--------------|
| `--format=<format>` | `table` (default, human only), `lines` (`key=value`), or `nul` (key, newline, value, NUL; unquoted). |
| `-z` | `--format=nul`. |
