# git log

Show commit history, newest first.

**Usage:** `git log [<flags>] [<revision-range>] [[--] <path>...]`

`<revision-range>` defaults to `HEAD`. `main..feature` = commits on `feature` not on `main`. `<path>` limits to commits touching those files.

## Flags

Also takes all [revision options](../shared/revision-options.md) except object traversal and all [diff options](../shared/diff-options.md) (`-p`, `--stat`, ...). No diff by default.

| Flag | What it does |
|------|--------------|
| `--follow` | Follow one file's history through renames. |
| `--decorate[=short\|full\|auto\|no]`, `--no-decorate` | Show ref names next to commits. `short` drops `refs/heads/` etc.; `full` keeps it; `auto` (default) short on a terminal only. Default from `log.decorate`. |
| `--[no-]decorate-refs=<pattern>` | Only decorate with matching refs. Repeatable. |
| `--[no-]decorate-refs-exclude=<pattern>` | Never decorate with matching refs. Repeatable. |
| `--clear-decorations` | Reset `--decorate-refs*`; decorate with every ref, not just branches, tags, remotes, stash, `HEAD`. |
| `--[no-]source` | Show which command-line ref reached each commit. |
| `--[no-]mailmap`, `--[no-]use-mailmap` | Rewrite identities via `.mailmap`. |
| `--full-diff` | With paths: show each matching commit's full diff, not just those paths. |
| `--log-size` | Print `log size <bytes>` per message. |
| `-L <start>,<end>:<file>`, `-L :<funcname>:<file>` | Trace a line range (or function regex) in `<file>` over time, patch per commit. Bounds: line numbers, `/regex/`, or `+<n>`/`-<n>` for `<end>`. Repeatable. No other paths; one starting revision max. |
| `-q`, `--[no-]quiet` | No diff output. |

## Merge commit diffs

Merges show no diff unless you pick one (or use `--first-parent`, which defaults to `first-parent`).

| Flag | What it does |
|------|--------------|
| `-m` | Diff merges in `log.diffMerges` format (default `separate`). Needs `-p`. |
| `-c` | Combined diff against all parents. `--diff-merges=combined -p`. |
| `--cc` | Like `-c`, hiding hunks taken from one side. `--diff-merges=dense-combined -p`. |
| `--dd` | Diff every commit against its first parent. `--diff-merges=first-parent -p`. |
| `--remerge-diff` | Diff a redone merge against the real one, showing conflict resolutions. `--diff-merges=remerge -p`. |
| `--diff-merges=<format>` | `off`/`none`, `on`/`m`, `first-parent`/`1`, `separate` (each parent), `combined`/`c`, `dense-combined`/`cc`, `remerge`/`r`. |
| `--no-diff-merges` | `--diff-merges=off`. |
| `--combined-all-paths` | Combined diffs: show each parent's file name (renames). |
