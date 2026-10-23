# git fsck-objects

Old name for [`git fsck`](fsck.md); use that. Same flags.

**Usage:** `git fsck-objects [<flags>] [<object>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<object>...` | Start reachability checks here. Default: index, refs, reflogs. |
| `--[no-]unreachable` | List objects no ref, reflog, or index entry reaches. |
| `--[no-]dangling` | List objects nothing points to (default). `--no-` is faster. |
| `--[no-]root` | List root commits. |
| `--[no-]tags` | List tags. |
| `--[no-]cache` | Also start from index entries. |
| `--[no-]reflogs` | Count reflog-only commits as reachable (default). `--no-` finds old branch tips. |
| `--[no-]full` | Also check packs and alternates (default). `--no-`: loose objects only. |
| `--[no-]connectivity-only` | Only check referenced objects exist. Faster; misses corrupt contents. |
| `--[no-]strict` | Extra checks, e.g. old group-writable modes. |
| `-v`, `--[no-]verbose` | More detail. Disables progress. |
| `--[no-]lost-found` | Save dangling objects to `.git/lost-found/commit/` or `other/`. Recovers lost work. |
| `--[no-]name-objects` | Show readable names like `HEAD@{1234567890}~25^2:src/`. |
| `--[no-]progress` | Force progress output. |
| `--[no-]references` | Also check refs with `git refs verify` (default). |
