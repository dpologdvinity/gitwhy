# git mv

Move or rename a file, directory, or symlink and stage it.

**Usage:** `git mv [<flags>] <source> <destination>` or `git mv [<flags>] <source>... <directory>`

## Flags

| Flag | What it does |
|------|--------------|
| `-f`, `--[no-]force` | Overwrite an existing destination. |
| `-k` | Skip moves that would fail instead of stopping. |
| `-n`, `--[no-]dry-run` | Show what would move. |
| `-v`, `--[no-]verbose` | Print each moved file. |
| `--[no-]sparse` | Allow moving files outside the sparse-checkout area. |

## Submodules

Moving a submodule also updates its `.git` file, `core.worktree`, and `.gitmodules` path (and stages `.gitmodules`, unless `-n`).
