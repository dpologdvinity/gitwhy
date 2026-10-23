# git clean

Delete untracked files, from the current folder down. Ignored files are kept unless `-x` or `-X`. Deletion is permanent; try `-n` first.

**Usage:** `git clean [-d] [-f] [-i] [-n] [-q] [-e <pattern>] [-x | -X] [--] [<pathspec>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `-f`, `--[no-]force` | Actually delete. Required unless `clean.requireForce=false`. `-ff` also deletes untracked nested git repos. |
| `-n`, `--[no-]dry-run` | List what would be deleted. |
| `-i`, `--[no-]interactive` | Pick what to delete from a menu (filter by pattern, select by number like `2-5 7,9`, ask each). No `-f` needed. |
| `-d` | Also delete untracked folders. Not needed with a `<pathspec>`. |
| `-q`, `--[no-]quiet` | Only report errors. |
| `-e <pattern>`, `--exclude=<pattern>` | Treat matches as ignored: kept, but deleted by `-X`. Repeatable. Works with `-x`. |
| `-x` | Also delete ignored files. `-e` still protects. |
| `-X` | Delete only ignored files. |
| `<pathspec>...` | Only clean matching paths. |
