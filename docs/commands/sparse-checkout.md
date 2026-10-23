# git sparse-checkout

Experimental: check out only some folders of a big repo. Other files stay tracked: branch switches leave them alone and `git commit -a` doesn't record them deleted.

**Cone mode** (default, recommended): list folders. You get all of each folder, plus files directly in its parents and the top level. `set A/B/C` gives top-level files, files directly in `A/` and `A/B/`, and all of `A/B/C/`. Fast.

**Non-cone mode** (`--no-cone`, deprecated): gitignore-style include patterns (`!` excludes). Slow on big repos, error-prone (quote globs), no sparse index. E.g. all but one file: `git sparse-checkout set --no-cone '/*' '!unwanted'`.

Selection lives in `$GIT_DIR/info/sparse-checkout`; excluded files get skip-worktree in the index. Never adds or removes submodules; use `git submodule init`/`deinit`.

**Usage:** `git sparse-checkout <subcommand> [<flags>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `set` | Enable if needed and replace the selection. |
| `add` | Add folders or patterns. |
| `list` | Show folders (cone) or patterns. |
| `reapply` | Reapply the selection (e.g. after a conflict restored files). Can switch modes. |
| `clean` | Delete leftover files in excluded folders (cone only). |
| `disable` | Turn off and restore all files. |
| `check-rules` | Show which stdin paths rules include. |
| `init` | Deprecated: `set` with no paths. |

### git sparse-checkout set

Sets `core.sparseCheckout`, `core.sparseCheckoutCone`, `index.sparse` as needed, writes the selection, updates files. Enables per-worktree config (`extensions.worktreeConfig`) so other worktrees are unaffected. Cone mode: an excluded folder is deleted, ignored files included, if all its untracked files are ignored; else kept with a warning. No paths: only top-level files remain.

**Usage:** `git sparse-checkout set [--[no-]cone] [--[no-]sparse-index] [--skip-checks] (--stdin | <patterns>...)`

| Flag | What it does |
|------|--------------|
| `<patterns>...` | Folders (cone; like `git ls-tree -d --name-only`, `"`-prefixed names C-quoted) or patterns (`--no-cone`). |
| `--[no-]cone` | Cone mode (default). `--no-cone`: gitignore-style patterns. |
| `--[no-]sparse-index` | Shrink the index to the selection for faster `status`, `add`, etc. Off by default; experimental. Old git and external tools may not understand it; undo with `git sparse-checkout init --no-sparse-index`. |
| `--skip-checks` | Skip possibly-wrong path checks (e.g. cone path is a folder). |
| `--stdin` | Read folders or patterns from stdin, one per line. |

### git sparse-checkout add

Add to an existing sparse selection. `add SOME/DIR` also brings in files directly under `SOME/`.

**Usage:** `git sparse-checkout add [--skip-checks] (--stdin | <patterns>...)`

| Flag | What it does |
|------|--------------|
| `<patterns>...` | Folders (cone) or patterns (non-cone). |
| `--skip-checks` | Skip possibly-wrong path checks. |
| `--stdin` | Read from stdin, one per line. |

### git sparse-checkout list

Print chosen folders (cone) or patterns.

**Usage:** `git sparse-checkout list`

No flags.

### git sparse-checkout reapply

Make the working tree match the selection again. Merges, rebases, tools, or modified/conflicted files can leave files outside it; fix those first.

**Usage:** `git sparse-checkout reapply [--[no-]cone] [--[no-]sparse-index]`

| Flag | What it does |
|------|--------------|
| `--[no-]cone` | Switch to/from cone mode, keeping the selection. |
| `--[no-]sparse-index` | Toggle sparse index. See `set`. |

### git sparse-checkout clean

Cone only: delete files left in excluded tracked folders (e.g. build output). Modified or conflicted files stay; resolve or stage them, `reapply`, then `clean` again. Keeps a sparse index efficient. Needs `-f` unless `clean.requireForce` is false.

**Usage:** `git sparse-checkout clean [-n | --dry-run] [-f | --force] [-v | --verbose]`

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]dry-run` | List folders that would be removed. |
| `-f`, `--[no-]force` | Actually delete. |
| `-v`, `--[no-]verbose` | List every file considered, showing why folders remain. |

### git sparse-checkout disable

Turn off `core.sparseCheckout`, restore all tracked files.

**Usage:** `git sparse-checkout disable`

No flags.

### git sparse-checkout check-rules

Print stdin paths (one per line, `"`-prefixed names C-quoted) that the rules include. Current rules unless `--rules-file`.

**Usage:** `git sparse-checkout check-rules [-z] [--[no-]cone] [--rules-file <file>]`

| Flag | What it does |
|------|--------------|
| `-z` | NUL-separated, unquoted paths in and out (not the rules file). |
| `--[no-]cone` | With `--rules-file`: rules are cone folders (default); `--no-cone`: patterns. |
| `--[no-]rules-file <file>` | Test rules from `<file>` (`set --stdin` format). |

### git sparse-checkout init

Deprecated; use `set`. Same as `set` with no paths; keeps an existing selection file.

**Usage:** `git sparse-checkout init [--[no-]cone] [--[no-]sparse-index]`

| Flag | What it does |
|------|--------------|
| `--[no-]cone` | Cone mode. |
| `--[no-]sparse-index` | Toggle sparse index. See `set`. |
