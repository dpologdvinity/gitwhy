# git shortlog

Summarize history by author: commit count and subjects per person. For release notes.

**Usage:** `git shortlog [<flags>] [<revision-range>] [[--] <path>...]` or `git log --pretty=short | git shortlog [<flags>]`

No revision with piped input (or no current branch): summarizes stdin. `<revision-range>` defaults to `HEAD`.

## Flags

Also takes commit limiting and history simplification [revision options](../shared/revision-options.md) (`--since`, `--author`, `--no-merges`, `--all`, ...).

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]numbered` | Sort by commit count, not name. |
| `-s`, `--[no-]summary` | Counts only, no subjects. |
| `-e`, `--[no-]email` | Show emails. |
| `--format[=<format>]` | Per-commit [format string](../shared/revision-options.md#placeholders) instead of subject (`--format='[%h] %s'`). |
| `--date=<format>` | [Date format](../shared/revision-options.md#date-formats) for placeholders. |
| `--[no-]group=<type>` | Group by `author` (default), `committer`, `trailer:<field>` (others skipped), or `format:<format>`. Repeatable. |
| `-c`, `--[no-]committer` | `--group=committer`. |
| `-w[<width>[,<indent1>[,<indent2>]]]` | Wrap at `<width>` (76); indent first line `<indent1>` (6), rest `<indent2>` (9). `0` = indent only. |

Names merge via `.mailmap` (see `git check-mailmap`).
