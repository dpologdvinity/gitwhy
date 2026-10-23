# git whatchanged

Deprecated, to be removed: use `git log --raw --no-merges`. Shows history with changed files per commit, no merges.

**Usage:** `git whatchanged --i-still-use-this [<flags>] [<revision-range>] [[--] <path>...]`

## Flags

Takes every [`git log`](log.md) flag, all [revision options](../shared/revision-options.md) and all [diff options](../shared/diff-options.md).

| Flag | What it does |
|------|--------------|
| `--i-still-use-this` | Required, or it refuses to run. |
| `-q`, `--[no-]quiet` | No diff output. |
| `--[no-]source` | Show which command-line ref led to each commit. |
| `--[no-]mailmap`, `--[no-]use-mailmap` | Map names and emails via `.mailmap`. |
| `--decorate[=short\|full\|auto\|no]`, `--no-decorate` | Show branch and tag names by commits. |
| `--[no-]decorate-refs=<pattern>` | Only decorate with matching refs. |
| `--[no-]decorate-refs-exclude=<pattern>` | Never decorate with matching refs. |
| `--clear-decorations` | Forget earlier `--decorate-refs*`; decorate with all refs. |
| `-L <start>,<end>:<file>`, `-L :<funcname>:<file>` | Trace a line range or function in `<file>`. |
