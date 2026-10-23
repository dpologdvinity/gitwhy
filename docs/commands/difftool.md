# git difftool

`git diff` that opens each changed file in a visual diff tool (Meld, vimdiff, VS Code, ...). Tool from `diff.tool`, else guessed.

**Usage:** `git difftool [<flags>] [<commit> [<commit>]] [--] [<path>...]`

## Flags

Also takes all [diff options](../shared/diff-options.md) and every `git diff` flag.

| Flag | What it does |
|------|--------------|
| `-t <tool>`, `--[no-]tool=<tool>` | Tool to use (e.g. `meld`). Define custom ones with `difftool.<tool>.cmd` using `$LOCAL` and `$REMOTE`. |
| `--[no-]tool-help` | List usable tools. |
| `-g`, `--[no-]gui` | Use `diff.guitool`. `--no-gui` overrides `difftool.guiDefault`. |
| `-x <command>`, `--[no-]extcmd=<command>` | Run `<command> $LOCAL $REMOTE` instead. |
| `-d`, `--[no-]dir-diff` | Open the tool once on two temp folders of all changed files. Never prompts. |
| `--[no-]symlinks` | With `--dir-diff`, symlink working files so edits stick. Default on, except Windows. |
| `-y`, `--no-prompt` | Don't ask before each file. |
| `--prompt` | Ask before each file (default); overrides config. |
| `--[no-]trust-exit-code` | Stop on tool error and return its exit code. Default: ignore. |
| `--rotate-to=<file>` | Start at `<file>`; earlier files go last. |
| `--skip-to=<file>` | Start at `<file>`; skip earlier files. |
| `--no-index` | Compare two paths outside git. |
| `--index` | Undo `--no-index`. |
