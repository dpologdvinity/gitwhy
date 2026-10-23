# git http-push

Push over "dumb" HTTP/WebDAV (no git server). `git push` runs it when needed. Disabled with libcurl < 7.16.

**Usage:** `git http-push [--all] [--dry-run] [--force] [--verbose] [-d | -D] <URL> <ref> [<ref>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `--all` | Check every object in the branch's history exists remotely, not just new ones. |
| `--force` | Allow non-fast-forward updates. Remote can lose commits. |
| `--dry-run` | Do everything but send. |
| `--verbose` | List objects walked and sent. |
| `-d`, `-D` | Delete `<ref>` remotely (not remote `HEAD`). `-d` requires remote `HEAD` and the branch locally, with the branch an ancestor of remote `HEAD`; `-D` skips checks. |
| `<URL>` | Remote repo. |
| `<ref>...` | Refs to update: `<name>` (= `<name>:<name>`) or `<src>:<dst>`. `<src>` must match one local ref. New `<dst>` needs `refs/` prefix unless equal to `<src>`. `+` prefix skips fast-forward check for that ref. |
