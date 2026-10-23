# git annotate

Show each line of a file with the commit that last changed it. `git blame` with an older output format (`<commit> (<author> <date> <line-no>)<line>`), kept for old scripts. Use `git blame` instead.

**Usage:** `git annotate [<flags>] [<rev-opts>] [<rev>] [--] <file>`

## Flags

Also takes [commit selection options](../shared/revision-options.md) to limit searched history.

| Flag | What it does |
|------|--------------|
| `<rev>` | Annotate the file as of this commit, not the working tree. |
| `<file>` | File to annotate. |
| `-L <start>,<end>`, `-L :<funcname>` | Only these lines. `<start>`/`<end>`: line number, `/<regex>/` (first match; `^/<regex>/` from file start), or for `<end>` `+<n>`/`-<n>` lines from start. Omitted = file start/end. `:<funcname>` = whole function by name regex. Repeatable. |
| `-b` | Blank commit ID for boundary commits (where search stopped). Same as `blame.blankBoundary`. |
| `--[no-]root` | Treat the root commit as normal, not a boundary. Same as `blame.showRoot`. |
| `--[no-]show-stats` | Print work-cost stats at the end. |
| `-l` | Show full commit IDs. |
| `-t` | Raw timestamps. |
| `-s` | Hide author name and date. |
| `-e`, `--[no-]show-email` | Author email instead of name. |
| `-f`, `--[no-]show-name` | Show original file name (automatic when renames were followed). |
| `-n`, `--[no-]show-number` | Show line number in the original commit. |
| `-c` | `git annotate` output format. Already default here; meant for `git blame`. |
| `-p`, `--[no-]porcelain` | Machine-readable; commit details once per commit. |
| `--[no-]line-porcelain` | `--porcelain`, repeating commit details every line. |
| `--[no-]incremental` | Machine-readable, printed as found. |
| `--[no-]abbrev[=<n>]` | `<n>`-digit commit IDs. |
| `--encoding=<encoding>` | Encoding for author names and summaries. `none` = unconverted. |
| `--date=<format>` | Date format, as `git log --date` (`relative`, `short`, `iso`). Default `blame.date`, else `iso`. |
| `--contents <file>` | Use `<file>`'s text as the final version (on top of `<rev>` or `HEAD`). `-` = stdin. For unsaved edits. |
| `--reverse <start>..<end>` | Search forward: last commit where each line still existed. `<start>` alone = `<start>..HEAD`. |
| `--first-parent` | Follow only first parents: when a line entered the main branch. |
| `-S <revs-file>` | Use commits in `<revs-file>` instead of walking history. |
| `--[no-]progress` | Force progress, even off a terminal. Not with `--porcelain` or `--incremental`. |
| `-M[<num>]` | Credit lines moved or copied within the file to their original commit. `<num>`: min matching alphanumerics (default 20). |
| `-C[<num>]` | `-M` across files changed in the same commit. Twice: also files in the file's creating commit. Three times: any commit. `<num>` default 40. |
| `-w` | Ignore whitespace changes. |
| `--diff-algorithm=<algorithm>` | Diff method: `myers` (default), `minimal`, `patience`, `histogram`. |
| `--[no-]ignore-rev <rev>` | Skip this commit; credit its lines to an earlier one. Repeatable. `blame.markIgnoredLines` marks such lines `?`; `blame.markUnblamableLines` marks unassignable ones `*`. |
| `--[no-]ignore-revs-file <file>` | Ignore commits listed in `<file>` (`fsck.skipList` format). Repeatable. `""` clears earlier lists, including `blame.ignoreRevsFile`. |
| `--[no-]color-lines` | Color annotations repeated from the line above (`color.blame.repeatedLines`, default cyan). |
| `--[no-]color-by-age` | Color annotations by line age (`color.blame.highlightRecent`). |
| `--[no-]score-debug` | Debug: show match scores. |
| `-h` | Show help. |
