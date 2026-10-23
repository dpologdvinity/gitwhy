# git init

Create a new repo (the `.git` folder) in the current folder or `<directory>`. Safe to rerun: changes nothing existing, only adds new template files (or moves the repo with `--separate-git-dir`).

Typical start: `git init`, `git add .`, `git commit`.

**Usage:** `git init [-q] [--bare] [--template=<dir>] [--separate-git-dir=<git-dir>] [--object-format=<format>] [--ref-format=<format>] [-b <branch>] [--shared[=<permissions>]] [<directory>]`

## Flags

| Flag | What it does |
|------|--------------|
| `<directory>` | Where to create it (made if missing). Default: current folder. |
| `-q`, `--[no-]quiet` | Only errors and warnings. |
| `--[no-]bare` | No working files, just git data. For server repos you push to. |
| `-b`, `--[no-]initial-branch=<name>` | First branch name. Default `init.defaultBranch`, else `master` (`main` from git 3.0). |
| `--[no-]object-format=<hash>` | `sha1` (default) or `sha256`. They can't share history yet. |
| `--[no-]ref-format=<format>` | `files` (default) or `reftable` (compact binary, faster with many refs). |
| `--[no-]template=<dir>` | Copy files from here into `.git` (skips dotfiles). Fallbacks: `$GIT_TEMPLATE_DIR`, `init.templateDir`, built-in templates. |
| `--[no-]separate-git-dir=<git-dir>` | Store git data at `<git-dir>`, leave a `.git` file pointing there. On an existing repo, moves it. |
| `--shared[=<permissions>]` | Let a group push. Sets `core.sharedRepository`: `group`/`true` (default), `all`/`world`/`everybody` (also world-readable), `umask`/`false` (normal), or octal like `0660`. Also refuses force-pushes (`receive.denyNonFastForwards`). |

## Environment variables

| Variable | What it does |
|----------|--------------|
| `GIT_DIR` | Create the repo here instead of `./.git`. With `--bare` and no `GIT_DIR`: current folder. |
| `GIT_OBJECT_DIRECTORY` | Object store location instead of `$GIT_DIR/objects`. |
| `GIT_TEMPLATE_DIR` | Template folder if no `--template`. |
| `GIT_DEFAULT_HASH` | Default `--object-format`. |
| `GIT_DEFAULT_REF_FORMAT` | Default `--ref-format`. |
