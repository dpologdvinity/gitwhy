# git check-ignore

Debug ignore rules: print the given paths that `.gitignore`, `.git/info/exclude`, or `core.excludesFile` ignores. Skips tracked files unless `--no-index`.

**Usage:** `git check-ignore [<flags>] <pathname>...`, `git check-ignore [<flags>] --stdin`

Exit code: `0` if any path is ignored, `1` if none, `128` on error.

## Flags

| Flag | What it does |
|------|--------------|
| `<pathname>...` | Paths to check; needn't exist. |
| `-q`, `--[no-]quiet` | No output, exit code only. One path only. |
| `-v`, `--[no-]verbose` | Show the matching rule: `<source>:<line>:<pattern><TAB><path>`. A matched `!pattern` means *not* ignored. |
| `--[no-]stdin` | Read paths from stdin, one per line. |
| `-z` | NUL-separated output (and input with `--stdin`). With `-v`, NUL replaces `:` and tab too. |
| `-n`, `--[no-]non-matching` | Also print unmatched paths, with empty rule fields. Use with `-v`. |
| `--no-index` | Check tracked files too. Finds why `git add .` picked up a file. |
| `--index` | Use the index (default). Undoes `--no-index`. |

`<source>` in `-v` output is absolute for `core.excludesFile`, else repo-relative. Buffering follows `GIT_FLUSH`.
