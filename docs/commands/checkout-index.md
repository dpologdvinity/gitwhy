# git checkout-index

Plumbing: copy files from the index to the working tree. Won't overwrite existing files unless `-f`. No file names and no `-a`: does nothing.

**Usage:** `git checkout-index [<flags>] [--] [<file>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `-u`, `--[no-]index` | Update index stat data for written files. |
| `-q`, `--[no-]quiet` | No warnings for existing files or files not in the index. |
| `-f`, `--[no-]force` | Overwrite existing files. |
| `-a`, `--[no-]all` | Check out every index file except skip-worktree ones. Not with file names. |
| `-n`, `--no-create` | Only refresh existing files; create none. |
| `--create` | Undo `--no-create`. |
| `--[no-]prefix=<string>` | Prepend `<string>` to written paths. `out/` exports into a directory; no trailing `/` makes a filename prefix. |
| `--stage=<1\|2\|3\|all>` | For conflicted files, write that stage (1 base, 2 ours, 3 theirs). `all` writes all three; implies `--temp`. |
| `--[no-]temp` | Write to temp files, print `tempname<TAB>path` (with `--stage=all`: three names, `.` for missing stage). Index not updated. For merge tools. |
| `--[no-]ignore-skip-worktree-bits` | Include skip-worktree files. |
| `--[no-]stdin` | Read paths from stdin, one per line. |
| `-z` | NUL-separated `--stdin` paths and `--temp` output. |
| `--` | Rest are file names. |
