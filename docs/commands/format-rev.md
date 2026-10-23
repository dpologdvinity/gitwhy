# git format-rev

Format commits from stdin with a `git log`-style format, one output per input, in order. Flushes per record for line-by-line scripting. Experimental.

Example: `git last-modified | git format-rev --stdin-mode=text --format=%an` swaps each commit ID for its author.

**Usage:** `git format-rev --stdin-mode=<mode> --format=<pretty> [--[no-]notes=<ref>] [-z] [--[no-]null-output] [--[no-]null-input]`

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]stdin-mode=<mode>` | `revs` (or `rev`): each line is a commit name (e.g. `HEAD~2`); prints it formatted plus a line end. `text`: replace full commit IDs in free text, adding no line end. |
| `--[no-]format=<pretty>` | Format, as in `git log --format` (e.g. `%an`, `reference`). |
| `--notes=<ref>`, `--no-notes` | Notes ref for `%N`. |
| `-z`, `--null` | NUL ends input and output records. Not negatable. |
| `--[no-]null-input` | NUL ends input records. |
| `--[no-]null-output` | NUL ends output records (e.g. with `%n`). |
