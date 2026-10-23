# git ls-files

List tracked files, and optionally untracked, ignored, modified, or deleted ones. A file may print more than once. For scripts; `git status` is friendlier.

**Usage:** `git ls-files [<flags>] [--] [<file>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<file>...` | Only these paths (patterns ok). |
| `-c`, `--[no-]cached` | List tracked files. Default unless `-s`, `-d`, `-o`, `-u`, `-k`, `-m`, `--resolve-undo` given. |
| `-d`, `--[no-]deleted` | Tracked files deleted but not staged. |
| `-m`, `--[no-]modified` | Tracked files changed but not staged (includes deleted). |
| `-o`, `--[no-]others` | Untracked files. |
| `-i`, `--[no-]ignored` | Only ignored files. Needs `-c` or `-o` plus an `--exclude*` flag. |
| `-s`, `--[no-]stage` | Show `<mode> <object> <stage> <file>`. |
| `-u`, `--[no-]unmerged` | Only conflicted files, with stages (1 base, 2 ours, 3 theirs). Implies `-s`. |
| `-k`, `--[no-]killed` | Untracked files/folders blocking checkout of a tracked file. |
| `--[no-]resolve-undo` | Files with resolve-undo info (`git checkout -m <path>` can restore the conflict). |
| `--[no-]directory` | With `-o`: show a fully untracked folder as `name/`. |
| `--[no-]empty-directory` | With `--directory`: `--no-empty-directory` hides empty folders. |
| `--[no-]eol` | Show line endings: `i/<eolinfo>` (index), `w/<eolinfo>` (worktree), `attr/<eolattr>`. `<eolinfo>`: `-text`, `none`, `lf`, `crlf`, `mixed`, or empty. |
| `--[no-]deduplicate` | Print each name once. No effect with `-t`, `-u`, `-s`. |
| `-t` | Prefix a status letter: `H` tracked, `S` skip-worktree, `M` unmerged, `R` deleted, `C` modified, `K` killed, `?` untracked, `U` resolve-undo. Prefer `git status --short`. |
| `-v` | Like `-t`; lowercase for "assume unchanged" files. |
| `-f` | Like `-t`; lowercase for files fsmonitor says are clean. |
| `-z` | NUL-terminate lines; no quoting. |
| `-x`, `--exclude=<pattern>` | Skip untracked files matching this gitignore pattern. Repeatable. |
| `-X`, `--exclude-from=<file>` | Read skip patterns from `<file>`. |
| `--[no-]exclude-per-directory=<file>` | Read skip patterns from this file in each folder (usually `.gitignore`). |
| `--exclude-standard` | Use normal ignore rules (`.gitignore`, `.git/info/exclude`, global ignore). |
| `--[no-]error-unmatch` | Exit 1 if any `<file>` isn't tracked. |
| `--[no-]with-tree=<tree-ish>` | With `--error-unmatch`: treat paths removed since `<tree-ish>` as present. |
| `--full-name` | Paths from repo top, not current folder. |
| `--[no-]recurse-submodules` | Include active submodules. Only with `--cached` or `--stage`. |
| `--[no-]abbrev[=<n>]` | Shorten IDs to at least `<n>` digits. |
| `--[no-]debug` | Print raw index data per file. Format may change. |
| `--[no-]sparse` | With a sparse index, show collapsed folders as `dir/`. |
| `--format=<format>` | Custom output: `%(objectmode)`, `%(objecttype)`, `%(objectname)`, `%(objectsize[:padded])` (`-` for commits/trees), `%(stage)`, `%(eolinfo:index)`, `%(eolinfo:worktree)`, `%(eolattr)`, `%(path)`. `%%` is `%`, `%xXX` a hex byte. Not with `-s`, `-o`, `-k`, `-t`, `--resolve-undo`, `--eol`. |
| `--` | Rest are file names. |
