# Diff options

Flags controlling how changes are shown. Taken by `git diff`, `diff-files`, `diff-index`, `diff-tree`, `diff-pairs`, `difftool`, `range-diff`, `log`, `show`, `whatchanged`, `stash show`, and mostly `format-patch`. For *which commits*, see [revision options](revision-options.md).

Terms: **hunk** = one block of changed lines. **Context** = unchanged lines around it. **Pre-image**/**post-image** = old/new version.

## Output format

| Flag | What it does |
|------|--------------|
| `-p`, `-u`, `--patch` | Full patch (`+`/`-` lines). Default for `git diff`. |
| `-s`, `--no-patch` | No diff. Cancels earlier `--patch`/`--stat`. |
| `-U<n>`, `--unified=<n>` | `<n>` context lines (default 3, `diff.context`). Implies `--patch`. |
| `-W`, `--[no-]function-context` | Whole enclosing function as context. |
| `--inter-hunk-context=<n>` | Join hunks up to `<n>` lines apart. Default 0. |
| `--raw` | One machine-readable line per file: modes, IDs, status, paths. |
| `--patch-with-raw` | `-p --raw`. |
| `-t` | In `log`/`show` diffs, also list tree entries. |
| `--stat[=<width>[,<name-width>[,<count>]]]` | Per-file added/removed bar graph. Caps total width (default terminal or 80), name width, and file count. |
| `--stat-width=<width>` | `--stat` total width. |
| `--stat-name-width=<width>` | `--stat` name width. |
| `--stat-graph-width=<width>` | `--stat` graph width. |
| `--stat-count=<count>` | `--stat` first `<count>` files only. |
| `--patch-with-stat` | `-p --stat`. |
| `--[no-]compact-summary` | In `--stat`, mark created/deleted files and mode changes (`new`, `gone`, `+x`). Implies `--stat`. |
| `--numstat` | Like `--stat`, plain numbers for scripts. Binary: `-` `-`. |
| `--shortstat` | Only the `--stat` totals line. |
| `-X[<param>,...]`, `--dirstat[=<param>,...]` | Percent of change per folder. Params: `changes` (default), `lines` (slower, counts moves), `files` (fastest), `cumulative` (include subfolders), `noncumulative`, and a minimum percent (default 3). E.g. `--dirstat=files,10,cumulative`. |
| `--cumulative` | `--dirstat=cumulative`. |
| `--dirstat-by-file[=<param>,...]` | `--dirstat=files,<param>,...`. |
| `--summary` | List created, deleted, renamed, copied files and mode changes. |
| `--name-only` | Changed file names only. |
| `--name-status` | Names with status letters (see `--diff-filter`). |
| `-z` | With `--raw`, `--numstat`, `--name-only`, `--name-status`: unquoted paths, NUL-terminated fields. |
| `--submodule[=<format>]` | Submodule changes: `short` (IDs; default), `log` (commits between; bare), `diff` (inline file diff). |
| `--[no-]color[=<when>]` | `always` (bare), `never`, or `auto` (terminal only). |
| `--[no-]color-moved[=<mode>]` | Color moved lines differently. `no`, `plain`, `blocks` (20+ alphanumerics), `zebra` (alternating; bare), `default` (= `zebra`), `dimmed-zebra`. |
| `--[no-]color-moved-ws=<mode>,...` | Whitespace ignored when matching moves: `no`, `ignore-space-at-eol`, `ignore-space-change`, `ignore-all-space`, or `allow-indentation-change` (uniform re-indent; alone only). |
| `--word-diff[=<mode>]` | Word-level diff: `plain` (`[-old-]{+new+}`; default), `color`, `porcelain` (for scripts), `none`. |
| `--word-diff-regex=<regex>` | Regex for a word (default: non-whitespace runs; `.` = per character). Implies `--word-diff`. |
| `--color-words[=<regex>]` | `--word-diff=color` plus optional `--word-diff-regex`. |
| `--check` | Warn on added conflict markers and whitespace errors; exit non-zero if any. Not with `--exit-code`. |
| `--ws-error-highlight=<kind>` | Highlight whitespace errors on `context`, `old`, `new` (comma list); `all`, `none`, `default` (= `new`). |
| `--[no-]full-index` | Full IDs on patch `index` lines. |
| `--binary` | Include binary changes `git apply` can use. Implies `--full-index`, `--patch`. |
| `--[no-]abbrev[=<n>]` | Short unique IDs (at least `<n>` digits) in `--raw` and `diff-tree` headers. |
| `--src-prefix=<prefix>` | Old-name prefix instead of `a/`. |
| `--dst-prefix=<prefix>` | New-name prefix instead of `b/`. |
| `--no-prefix` | No `a/`/`b/` prefixes. |
| `--default-prefix` | Force `a/`/`b/`, ignoring config. |
| `--line-prefix=<prefix>` | Prefix every output line. |
| `--output=<file>` | Write to `<file>`. |
| `--output-indicator-new=<char>` | Added-line marker instead of `+`. |
| `--output-indicator-old=<char>` | Removed-line marker instead of `-`. |
| `--output-indicator-context=<char>` | Context-line marker instead of space. |

## Renames and copies

| Flag | What it does |
|------|--------------|
| `-M[<n>]`, `--find-renames[=<n>]` | Detect renames. `<n>` = required similarity (`-M90%`; `-M5` = 50%; `-M100%` = exact). Default 50%. |
| `-C[<n>]`, `--find-copies[=<n>]` | Also detect copies, from files changed in the same change. `<n>` as `-M`. |
| `--[no-]find-copies-harder` | Unchanged files can be copy sources too. Slow. Same as `-C -C`. |
| `-B[<n>][/<m>]`, `--break-rewrites[=[<n>][/<m>]]` | Show near-total rewrites as delete plus add. `<m>` (60%): change needed to count. `<n>` (50%): with `-M`, change needed to be a rename source. |
| `-D`, `--irreversible-delete` | Deleted files: header only. Can't be applied. |
| `-l<n>` | Skip thorough rename/copy detection above `<n>` files. Default `diff.renameLimit` (1000); 0 = no limit. |
| `--no-renames` | No rename detection, despite config. |
| `--[no-]rename-empty` | Empty files can be rename sources. |
| `--[no-]follow` | Follow one file across renames. `git log` with a single path only. |

## Algorithm and whitespace

| Flag | What it does |
|------|--------------|
| `--diff-algorithm=<algo>` | `myers`/`default` (default), `minimal`, `patience`, `histogram`. `default` overrides config. |
| `--minimal` | Smallest diff; slower. |
| `--patience` | Patience algorithm; clearer for moved blocks. |
| `--histogram` | Histogram algorithm (faster patience). |
| `--anchored=<text>` | Keep unique lines starting with `<text>` unchanged. Uses patience. Repeatable. |
| `--[no-]indent-heuristic` | Shift hunk edges for readability. Default on. |
| `-w`, `--ignore-all-space` | Ignore all whitespace. |
| `-b`, `--ignore-space-change` | Ignore whitespace amount and line-end changes. |
| `--ignore-space-at-eol` | Ignore line-end whitespace. |
| `--ignore-cr-at-eol` | Ignore trailing `\r`. |
| `--ignore-blank-lines` | Ignore added/removed blank lines. |
| `-I<regex>`, `--[no-]ignore-matching-lines=<regex>` | Ignore changes where every line matches `<regex>`. Repeatable. |
| `-a`, `--[no-]text` | Treat all files as text. |
| `--[no-]textconv` | Run textconv filters so binaries diff as text. Default only in `git diff` and `git log`. |
| `--[no-]ext-diff` | Allow a configured external diff program. |

## Choosing what to show

| Flag | What it does |
|------|--------------|
| `--diff-filter=[ACDMRTUXB...][*]` | Only these change types: `A`dded, `C`opied, `D`eleted, `M`odified, `R`enamed, `T`ype changed, `U`nmerged, `X` unknown, `B` broken. Lowercase excludes. `*`: all files if any match. |
| `-S<string>` | Only changes that alter the count of `<string>`. Searches binaries too. |
| `-G<regex>` | Only changes whose added/removed lines match `<regex>` (includes edits near a match). |
| `--pickaxe-regex` | Treat `-S` as an extended regex. |
| `--pickaxe-all` | On `-S`/`-G` match, show every file in that change. |
| `--find-object=<id>` | Only changes adding or removing this blob or submodule commit. |
| `--ignore-submodules[=<when>]` | Hide submodule changes: `all` (default), `dirty` (new commits only), `untracked`, `none`. |
| `--[no-]relative[=<path>]` | Only changes under the current folder (or `<path>`), paths relative to it. |
| `--max-depth=<depth>` | Descend at most `<depth>` levels per pathspec (-1 = unlimited). No wildcards; tree diffs only. |
| `-O<orderfile>` | Order files by globs in `<orderfile>`; unmatched last. `-O/dev/null` cancels `diff.orderFile`. |
| `--skip-to=<file>` | Start at `<file>`, dropping earlier files. |
| `--rotate-to=<file>` | Start at `<file>`, moving earlier files to the end. |
| `-R` | Swap old and new. |
| `--ita-invisible-in-index` | Show `git add -N` files as new in `git diff`, absent in `--cached`. Experimental. |
| `--ita-visible-in-index` | Undo `--ita-invisible-in-index`. Experimental. |

## Exit status

| Flag | What it does |
|------|--------------|
| `--[no-]exit-code` | Exit 1 if differences, 0 if none. |
| `--[no-]quiet` | No output; implies `--exit-code`. |
