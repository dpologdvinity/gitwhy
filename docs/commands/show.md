# git show

Show objects. Commits: message and diff (merges as compact combined diff). Tags: message and tagged object. Trees: file names. Blobs: contents.

**Usage:** `git show [<flags>] [<object>...]`

`<object>` defaults to `HEAD`. Examples: `HEAD~2`, `v1.0`, `main:README.md` (file as of `main`).

## Flags

Also takes commit formatting [revision options](../shared/revision-options.md#commit-formatting) (`--pretty`, `--format`, `--oneline`, `--abbrev-commit`, `--no-abbrev-commit`, `--encoding`, `--expand-tabs`, `--notes`, `--no-notes`, `--show-notes-by-default`, `--show-notes`, `--show-signature`, `--date`, ...) and all [diff options](../shared/diff-options.md) (`-p`, `--stat`, `-s`/`--no-patch`, ...). Most `git log` flags work too.

| Flag | What it does |
|------|--------------|
| `-q`, `--[no-]quiet` | Hide the diff. |
| `--[no-]source` | Show which command-line ref reached each commit. |
| `--[no-]mailmap`, `--[no-]use-mailmap` | Rewrite names and emails via `.mailmap`. |
| `--decorate[=short\|full\|auto\|no]`, `--no-decorate` | Show branch and tag names. |
| `--[no-]decorate-refs=<pattern>` | Decorate only with matching refs. |
| `--[no-]decorate-refs-exclude=<pattern>` | Never decorate with matching refs. |
| `--clear-decorations` | Drop earlier `--decorate-refs*`; use all refs. |
| `-L <start>,<end>:<file>`, `-L :<funcname>:<file>` | Only changes to that range or function in `<file>`. |

## Merge commit diffs

| Flag | What it does |
|------|--------------|
| `-m` | Merge diffs in the `log.diffMerges` format (default: against each parent separately). |
| `-c` | Combined diff against all parents. `--diff-merges=combined -p`. |
| `--cc` | Compact combined diff, hiding one-side hunks. Default. |
| `--dd` | Diff against first parent. `--diff-merges=first-parent -p`. |
| `--remerge-diff` | Diff a re-done merge against the real one: shows conflict resolutions. |
| `--diff-merges=<format>` | `off`/`none`, `on`/`m`, `first-parent`/`1`, `separate`, `combined`/`c`, `dense-combined`/`cc`, `remerge`/`r`. |
| `--no-diff-merges` | `--diff-merges=off`. |
| `--combined-all-paths` | Combined diffs: each parent's file name. |
