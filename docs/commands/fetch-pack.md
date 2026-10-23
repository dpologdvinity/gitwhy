# git fetch-pack

Plumbing: download missing objects by talking to a remote `git upload-pack`. Use [`git fetch`](fetch.md) instead; it wraps this and updates refs.

Prints fetched refs as `<id> <ref>` lines.

**Usage:** `git fetch-pack [<flags>] <repository> [<refs>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<repository>` | Remote URL or path (`[<host>:]<directory>` works). |
| `<refs>...` | Remote refs to fetch (e.g. `refs/heads/main`). None given and no `--all`/`--stdin`: fetches nothing. Raw IDs work if the server allows (`uploadpack.allowTipSHA1InWant`, `allowReachableSHA1InWant`, `allowAnySHA1InWant`). |
| `--all` | Fetch every remote ref. |
| `--stdin` | Also read refs from stdin, one per line. With `--stateless-rpc`: pkt-lines ending in flush. |
| `-q`, `--quiet` | Less output (passes `-q` to `git unpack-objects`). |
| `-k`, `--keep` | Keep received data as a pack, not loose objects. Twice: also lock the pack against repacking. |
| `--thin` | Ask for a thin pack (deltas against objects you have). Less data. |
| `--include-tag` | Also get annotated tags pointing at fetched objects, if supported. |
| `--upload-pack=<path>`, `--exec=<path>` | Path to `git-upload-pack` on the remote. |
| `--depth=<n>` | Shallow: at most `<n>` commits per tip. `2147483647` = unlimited. |
| `--shallow-since=<date>` | Shallow: only commits after `<date>`. |
| `--shallow-exclude=<ref>` | Shallow: exclude commits reachable from this remote ref. Repeatable. |
| `--deepen-relative` | Count `--depth` from the current shallow boundary. |
| `--refetch` | Refetch all matching objects, ignoring what you have. Use after changing a partial-clone `--filter`. |
| `--filter=<filter-spec>` | Partial clone filter (e.g. `blob:none`). See `git rev-list --filter`. |
| `--no-filter` | Cancel an earlier `--filter`. |
| `--no-progress` | No progress. |
| `--check-self-contained-and-connected` | Print `connectivity-ok` if the pack is complete and connected. |
| `-v` | More output. |
| `--diag-url` | Debug: show how the URL parses, then exit. |
| `--stateless-rpc` | Internal: one request/response, for smart HTTP. |
| `--lock-pack` | Internal: lock the pack and print the lock file name. |
| `--cloning` | Internal: part of a clone. |
| `--update-shallow` | Internal: accept refs that update `.git/shallow`. |
| `--from-promisor` | Internal: mark the pack as from a promisor remote. |
