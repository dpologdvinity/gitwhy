# git prune-packed

Delete loose objects already in a pack. `git gc` and `git repack -d` run it for you.

**Usage:** `git prune-packed [-n | --dry-run] [-q | --quiet]`

## Flags

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]dry-run` | List what would be deleted. |
| `-q`, `--[no-]quiet` | No progress. |
