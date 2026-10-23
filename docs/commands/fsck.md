# git fsck

Check the object store for corruption, missing objects, and unreachable objects. Read-only unless `--lost-found`.

**Usage:** `git fsck [<flags>] [<object>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<object>...` | Start reachability checks here. Default: index, all refs, all reflogs. |
| `--[no-]unreachable` | List objects no ref, reflog, or index entry reaches. |
| `--[no-]dangling` | List objects nothing points to. On by default; `--no-dangling` is faster. |
| `--[no-]root` | List root commits. |
| `--[no-]tags` | List tags. |
| `--[no-]cache` | Also start from index entries. |
| `--[no-]reflogs` | Reflog-only commits count as reachable (default). `--no-reflogs` finds commits branches used to point to. |
| `--[no-]full` | Also check packs and alternate stores (default). `--no-full`: loose objects only. |
| `--[no-]connectivity-only` | Only check referenced objects exist; skip contents. Faster, misses corrupt files. |
| `--[no-]strict` | Extra checks, e.g. group-writable modes from old git. |
| `-v`, `--[no-]verbose` | More detail. Disables progress. |
| `--[no-]lost-found` | Save dangling objects to `.git/lost-found/commit/` or `.git/lost-found/other/` (blobs as contents). Recovers lost work. |
| `--[no-]name-objects` | Show readable paths for reachable objects, like `HEAD@{1234567890}~25^2:src/`. |
| `--[no-]progress` | Progress even if stderr isn't a terminal. Default on in a terminal. |
| `--[no-]references` | Also run `git refs verify` (default). |

## Output

| Line | Meaning |
|------|---------|
| `unreachable <type> <id>` | Nothing you started from reaches it. Deletable if you gave all start points. |
| `missing <type> <id>` | Referenced but not stored. |
| `dangling <type> <id>` | Nothing points to it. Dangling commit may be lost work. |
| `hash mismatch <id>` | Contents don't match ID. Serious corruption. |

Other problems print as `<msgId>: <text>` (e.g. `missingEmail`). Config `fsck.<msgId>` = `error`, `warn`, or `ignore` sets severity (not FATAL ones). `fsck.skipList` skips known-bad objects. Message IDs: [git-scm.com/docs/git-fsck](https://git-scm.com/docs/git-fsck#_fsck_messages).

Also runs `git commit-graph verify` if `core.commitGraph` is true.

## Environment variables

| Variable | What it does |
|----------|--------------|
| `GIT_OBJECT_DIRECTORY` | Object store (usually `.git/objects`). |
| `GIT_INDEX_FILE` | Index file. |
| `GIT_ALTERNATE_OBJECT_DIRECTORIES` | Extra object stores. |
