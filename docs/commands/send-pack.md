# git send-pack

Plumbing push: send objects and ref updates to a remote `git receive-pack`. Use [`git push`](push.md) instead.

**Usage:** `git send-pack [<flags>] [<host>:]<directory> (--all | <ref>...)`

`<host>:` connects over ssh. Refs sent:
- `--all`: every local branch. No `<ref>` allowed.
- Neither: branches on both sides.
- `<ref>`: `<src>:<dst>`, or `<name>` for `<name>:<name>`. `<src>` must match one local ref; `<dst>` at most one remote ref, else must start with `refs/` (or equal `<src>`). Prefix `+` to force that ref.

Without force, only fast-forward updates are allowed.

## Flags

| Flag | What it does |
|------|--------------|
| `-v`, `--[no-]verbose` | More output. |
| `-q`, `--[no-]quiet` | Less output. |
| `--[no-]receive-pack=<path>` | Path to `git-receive-pack` on the remote. |
| `--[no-]exec=<path>` | Same as `--receive-pack`. |
| `--[no-]remote=<remote>` | Remote name, for looking up its remote-tracking refs (e.g. `--force-with-lease`). |
| `--[no-]all` | Push every local branch. |
| `-n`, `--[no-]dry-run` | Do everything but send. |
| `--[no-]mirror` | Make remote refs match local: push all, delete remote refs missing locally. |
| `-f`, `--[no-]force` | Skip the fast-forward check. Remote can lose commits. |
| `--[no-]signed[=(true\|false\|if-asked)]` | GPG-sign the push. `true`: fail if unsupported. `if-asked`: sign only if supported. Fails if gpg fails. |
| `--[no-]push-option=<string>` | Send string to server hooks. Errors if unsupported. Repeatable. |
| `--[no-]progress` | Show progress even off a terminal. |
| `--[no-]thin` | Send a thin pack (deltas against objects the remote has). |
| `--[no-]atomic` | All refs update or none. |
| `--[no-]stateless-rpc` | Internal: stateless protocol for smart HTTP. |
| `--[no-]stdin` | Also read refs from stdin, one per line. With `--stateless-rpc`: pkt-line, ending in flush. |
| `--[no-]helper-status` | Internal: per-ref status in remote-helper format. |
| `--[no-]force-with-lease[=<refname>[:<expect>]]` | Force only if the remote ref still has the expected value (default: your remote-tracking ref). See [`push`](push.md). |
| `--[no-]force-if-includes` | With `--force-with-lease`: also require the remote-tracking ref's commits be integrated locally. |
