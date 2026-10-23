# git grep

Search tracked files for a pattern. Can also search the index, any commit or tree, untracked files, or non-git folders. Empty pattern matches every line. Exit 0 on match, 1 if none.

**Usage:** `git grep [<flags>] [-e] <pattern> [<tree>...] [[--] <pathspec>...]`

- `<tree>...`: search these commits/trees instead (e.g. `git grep foo HEAD~3`).
- `<pathspec>...`: limit to matching paths (globs; `:^docs` excludes). Use `--` if ambiguous.
- Examples: `git grep 'time_t' -- '*.[ch]'`; `git grep -e '#define' --and \( -e MAX_PATH -e PATH_MAX \)`.

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]cached` | Search the index, not working files. |
| `--no-index` | Search the current folder ignoring git (like `grep -r`). Not with `--cached` or `--untracked`. See `grep.fallbackToNoIndex`. |
| `--index` | Undo `--no-index`. |
| `--[no-]untracked` | Also search untracked files. |
| `--[no-]exclude-standard` | Honor `.gitignore`. `--no-exclude-standard` with `--untracked`: include ignored files. `--exclude-standard` with `--no-index`: skip them. |
| `--[no-]recurse-submodules` | Also search active checked-out submodules. Not with `--untracked`; no effect with `--no-index`. |
| `--parent-basename <basename>` | Internal: path prefix for submodule results. |
| `-v`, `--[no-]invert-match` | Show non-matching lines. |
| `-i`, `--[no-]ignore-case` | Ignore case. |
| `-w`, `--[no-]word-regexp` | Whole words only. |
| `-a`, `--[no-]text` | Treat binary files as text. |
| `-I` | Skip binary files. |
| `--[no-]textconv` | Apply textconv filters first. Default off. |
| `-r`, `--[no-]recursive` | Recurse (default; `--max-depth=-1`). `--no-recursive` = `--max-depth=0`. |
| `--max-depth <n>` | Max `<n>` levels below each pathspec; `-1` unlimited. Ignored with active wildcards. |
| `-E`, `--[no-]extended-regexp` | POSIX extended regex. |
| `-G`, `--[no-]basic-regexp` | POSIX basic regex (default). |
| `-F`, `--[no-]fixed-strings` | Literal strings. |
| `-P`, `--[no-]perl-regexp` | Perl regex. Needs PCRE build. |
| `-n`, `--[no-]line-number` | Line numbers. |
| `--[no-]column` | Byte column (from 1) of first match. |
| `-h` | Hide file names. |
| `-H` | Show file names (default). |
| `--[no-]full-name` | Paths from repo root. |
| `-l`, `--[no-]files-with-matches` | List files with matches. |
| `--[no-]name-only` | Same as `-l`. |
| `-L`, `--[no-]files-without-match` | List files without matches. |
| `-z`, `--[no-]null` | NUL after file names, unquoted. |
| `-o`, `--[no-]only-matching` | Print only matched parts, one per line. |
| `-c`, `--[no-]count` | Matching line count per file. |
| `--[no-]color[=<when>]` | Highlight matches: `always` (default), `never`, `auto`. |
| `--[no-]break` | Blank line between files. |
| `--[no-]heading` | File name once above its matches. |
| `-C`, `--[no-]context <n>` | `<n>` lines before and after; `--` between groups. |
| `-<n>` | Same as `-C <n>`. |
| `-B`, `--before-context <n>` | `<n>` lines before. |
| `-A`, `--after-context <n>` | `<n>` lines after. |
| `--[no-]threads <n>` | Worker threads. Default `grep.threads`, else one per CPU. Ignored with `-O`. |
| `-p`, `--[no-]show-function` | Show the enclosing function line before each match. |
| `-W`, `--[no-]function-context` | Show the whole enclosing function. |
| `-f <file>` | Read patterns from a file, one per line (may contain NUL if regex type allows). |
| `-e <pattern>` | Next arg is a pattern. Needed for patterns starting with `-` and with `--and`/`--or`/`--not`. Multiple `-e` are OR-ed. |
| `--and`, `--or`, `--not`, `(`, `)` | Combine `-e` patterns. `--or` default; `--and` binds tighter. Quote parentheses. |
| `--[no-]all-match` | With OR-ed patterns, only files matching all. |
| `-q`, `--[no-]quiet` | No output; exit code only. |
| `-O`, `--[no-]open-files-in-pager[=<pager>]` | Open matching files in a pager (default `core.pager`). `less`/`vi` with one pattern jump to first match. Attach `<pager>` (`-Ovi`). |
| `--[no-]ext-grep` | No-op. |
| `-m`, `--[no-]max-count <n>` | Stop after `<n>` matches per file (`-v`: non-matches). `-1` unlimited (default); `0` exits at once, non-zero. |
