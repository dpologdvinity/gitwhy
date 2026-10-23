# git pickaxe

Old alias for [`git blame`](blame.md); use that. Same flags. (Unrelated to `git log -S`/`-G` "pickaxe" search.)

**Usage:** `git pickaxe [<flags>] [<rev-opts>] [<rev>] [--] <file>`

## Flags

Also takes [revision options](../shared/revision-options.md) (`<rev-opts>`), e.g. `--reverse <start>..<end>`, `--first-parent`, `--encoding=<enc>`, `--date=<format>`.

| Flag | What it does |
|------|--------------|
| `--[no-]incremental` | Stream results in machine format. |
| `-b` | Blank IDs for boundary commits. |
| `--[no-]root` | Don't mark root commits as boundaries. |
| `--[no-]show-stats` | Print work stats at the end. |
| `--[no-]progress` | Force progress. Not with `--porcelain` or `--incremental`. |
| `--[no-]score-debug` | Show `-C`/`-M` scores. |
| `-f`, `--[no-]show-name` | Show original file name. Default: only after renames. |
| `-n`, `--[no-]show-number` | Show original line number. |
| `-p`, `--[no-]porcelain` | Machine-readable output. |
| `--[no-]line-porcelain` | Like `--porcelain`, commit info on every line. |
| `-c` | `git annotate` format. |
| `-t` | Raw timestamps. |
| `-l` | Full commit IDs. |
| `-s` | Hide author and timestamp. |
| `-e`, `--[no-]show-email` | Email instead of name. |
| `-w` | Ignore whitespace changes. |
| `--diff-algorithm=<algorithm>` | `default`/`myers`, `minimal`, `patience`, `histogram`. |
| `--[no-]ignore-rev=<rev>` | Skip this commit; blame the one before. Repeatable. |
| `--[no-]ignore-revs-file=<file>` | Skip commits listed in `<file>`. Repeatable; `""` clears. |
| `--[no-]color-lines` | Color repeated-commit annotations differently. |
| `--[no-]color-by-age` | Color by line age. |
| `-S <file>` | Revision list from `<file>`. |
| `--[no-]contents=<file>` | Blame these contents as the final version (`-` = stdin). |
| `-C[<score>]` | Find lines from other files in the same commit. Twice: the creating commit's files. Thrice: any commit. Default 40 chars. |
| `-M[<score>]` | Find lines moved within the file. Default 20 chars. |
| `-L <range>` | Only lines `<start>,<end>` or `:<funcname>`. Repeatable. |
| `--[no-]abbrev[=<n>]` | At least `<n>` hex digits per ID. |
