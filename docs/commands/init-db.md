# git init-db

Old name for [`git init`](init.md); use that. Same flags.

**Usage:** `git init-db [<flags>] [<directory>]`

## Flags

| Flag | What it does |
|------|--------------|
| `-q`, `--[no-]quiet` | Errors and warnings only. |
| `--[no-]bare` | Bare repo: no working files. |
| `--[no-]template=<dir>` | Copy hooks etc. from `<dir>` into `.git`. |
| `--[no-]separate-git-dir=<git-dir>` | Put git data at `<git-dir>`; leave a `.git` file pointing there. |
| `-b <name>`, `--[no-]initial-branch=<name>` | First branch name. |
| `--[no-]object-format=<hash>` | `sha1` or `sha256`. |
| `--[no-]ref-format=<format>` | `files` or `reftable`. |
| `--shared[=<permissions>]` | Share with a group or everyone: `group`, `all`, `umask`, or octal like `0660`. |
