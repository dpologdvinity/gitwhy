# git hook

Run or list hooks (scripts git runs at events like `pre-commit`). Hooks come from `.git/hooks/<event>` and from config, so one script can serve many repos:

```
[hook "linter"]
  event = pre-commit
  event = pre-push
  command = ~/bin/linter --cpp20
```

Config hooks run in the order their `event` lines are read; the `.git/hooks` script runs last. Unknown config events are ignored (wrapper tools can define their own). Event list: `githooks(5)`.

**Usage:** `git hook <subcommand> [<flags>] <hook-name>`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `run` | Run all hooks for an event. |
| `list` | List hooks for an event. Warns and exits 1 if none. |

### git hook run

`git hook run [<flags>] <hook-name> [-- <hook-args>]`. Hook arguments go after `--`.

| Flag | What it does |
|------|--------------|
| `--allow-unknown-hook-name` | Accept unknown event names (default rejects, to catch typos). |
| `--ignore-missing` | Exit 0 quietly if no hook exists. |
| `--to-stdin=<path>` | Feed this file to the hook's stdin. |
| `-j`, `--jobs=<n>` | Run up to `<n>` hooks at once. Default `hook.jobs`, else 1. Overrides per-hook `parallel`. Always serial: `applypatch-msg`, `pre-commit`, `prepare-commit-msg`, `commit-msg`, `post-commit`, `post-checkout`, `push-to-checkout`. |

### git hook list

`git hook list [<flags>] <hook-name>`

| Flag | What it does |
|------|--------------|
| `--allow-unknown-hook-name` | Accept unknown event names. |
| `-z` | NUL-terminated lines. |
| `--show-scope` | Prefix config hooks with scope (`local`, `global`, `system`) and a tab. |
