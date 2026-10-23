# git blame

Show the commit, author, and date that last changed each line of a file. Follows whole-file renames. Doesn't show deleted lines; use `git log -S'<text>'` for those.

Each line: short commit ID, author, date, line number, text. `^` before the ID marks a boundary commit (line is at least that old).

**Usage:** `git blame [<flags>] [<rev> | --reverse <start>..<end>] [--] <file>`

`<rev>` starts from that commit instead of the working tree. Ranges limit how far back: `git blame v2.0.. -- foo`, `git blame --since=3.weeks -- foo`; older lines land on the boundary commit.

## Flags

Also takes [commit selection options](../shared/revision-options.md) such as `--since` and `--first-parent`.

| Flag | What it does |
|------|--------------|
| `-L <start>,<end>`, `-L :<funcname>` | Only these lines. Repeatable. Each side: line number, `/regex/` (from previous `-L`; `^/regex/` from file start), or `+N`/`-N` for `<end>`; may be blank. `:<funcname>`: function matching regex. |
| `--reverse <start>..<end>` | Walk forward: show the last commit where each line still existed. `<start>` alone means `<start>..HEAD`. |
| `--first-parent` | Follow only first parents of merges: when the line reached this branch. |
| `-S <revs-file>` | Use the commit list in this file instead of walking history. |
| `--[no-]contents=<file>` | Blame this file's contents as the newest version (`-` = stdin). |
| `-M[<num>]` | Blame lines moved/copied within the file on their origin. `<num>`: min alphanumeric chars (default 20). |
| `-C[<num>]` | Like `-M`, plus lines from other files changed in the same commit. Twice: files present at creation. Thrice: any file, any commit. Default 40. |
| `--[no-]ignore-rev=<rev>` | Skip this commit; its lines go to the previous change. Repeatable. For reformat commits. |
| `--[no-]ignore-revs-file=<file>` | Skip every commit listed in the file (full IDs, `#` comments). Repeatable; `""` clears. Applied after `blame.ignoreRevsFile`. |
| `-w` | Ignore whitespace changes. |
| `--diff-algorithm=<algorithm>` | `myers` (`default`), `minimal`, `patience`, or `histogram`. |
| `-b` | Blank IDs for boundary commits. Default: `blame.blankBoundary`. |
| `--[no-]root` | Don't mark root commits as boundaries. Default: `blame.showRoot`. |
| `-l` | Full commit IDs. |
| `--[no-]abbrev=<n>` | At least `<n>` hex digits per ID. |
| `-t` | Raw timestamps. |
| `--date=<format>` | Date style, as in `git log --date`. Default: `blame.date`, else `iso`. |
| `-s` | Hide author and date. |
| `-e`, `--[no-]show-email` | Author email instead of name. Default: `blame.showEmail`. |
| `-f`, `--[no-]show-name` | Show the original file name. Default: only if some lines came from a rename. |
| `-n`, `--[no-]show-number` | Show the original line number. |
| `-c` | `git annotate`'s layout. |
| `-p`, `--[no-]porcelain` | Machine-readable; full commit info only on a commit's first appearance. |
| `--[no-]line-porcelain` | Like `--porcelain`, full commit info on every line. |
| `--[no-]incremental` | Machine-readable, streamed out of order, no line text. For GUIs. |
| `--encoding=<encoding>` | Encoding for author names and summaries. `none` = no conversion. |
| `--[no-]color-lines` | Color repeated-commit annotations differently (`color.blame.repeatedLines`). |
| `--[no-]color-by-age` | Color annotations by line age (`color.blame.highlightRecent`). |
| `--[no-]progress` | Force progress output. Not with `--porcelain` or `--incremental`. |
| `--[no-]show-stats` | Print work stats at the end. |
| `--[no-]score-debug` | Show `-M`/`-C` detection scores. |
| `-h` | Help. |

Config: `blame.markIgnoredLines` marks lines reassigned by `--ignore-rev` with `?`; `blame.markUnblamableLines` marks unreassignable ones with `*`.
