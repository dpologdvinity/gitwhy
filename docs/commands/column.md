# git column

Lay out stdin lines in columns. Git uses it for branch, tag, and untracked-file lists.

Example: `seq 1 24 | git column --mode=column --padding=5`

**Usage:** `git column [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]command=<name>` | Use config `column.<name>`, else `column.ui`. |
| `--[no-]mode[=<mode>]` | Layout words, comma/space separated (as `column.ui`). When: `always`, `never` (default), `auto` (terminal only). Order: `column` (down first, default), `row` (across), `plain` (one column). Width: `dense` (fit) or `nodense` (equal, default). |
| `--raw-mode=<n>` | `--mode` as a number. Internal. |
| `--[no-]width=<width>` | Total width. Default: terminal width or 80. |
| `--[no-]indent=<string>` | Line prefix. |
| `--[no-]nl=<string>` | Line suffix, newline included. |
| `--[no-]padding=<n>` | Spaces between columns (default 1). |
