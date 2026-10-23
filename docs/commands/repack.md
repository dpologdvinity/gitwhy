# git repack

Pack loose objects into a pack file, or combine packs into fewer. Packs store similar objects as deltas. `git gc` runs it.

Common: `git repack -a -d` (one pack, delete old). Add `-f --depth=50 --window=250` for a slow, thorough repack.

**Usage:** `git repack [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `-a` | Pack everything reachable into one pack. Use with `-d`, which then deletes unreachable packed objects. Promisor packs are repacked separately. Dumb-HTTP clients must re-download it all. |
| `-A` | Like `-a`, but with `-d` unreachable objects become loose (left for `git gc` expiry) instead of deleted. |
| `-d` | Delete now-redundant packs and loose copies (`git prune-packed`). |
| `--[no-]cruft` | Like `-a`, but with `-d` unreachable objects go into an age-tracking cruft pack. Not with `-k`. |
| `--[no-]cruft-expiration=<approxidate>` | With `--cruft -d`: delete unreachable objects older than this now. |
| `--max-cruft-size=<n>` | Size limit for new cruft packs (default: `--max-pack-size`). |
| `--combine-cruft-below-size=<n>` | With `--cruft`, no pruning: merge only cruft packs under `<n>` bytes. |
| `--[no-]expire-to=<dir>` | With `--cruft -d`: back up pruned objects as a pack in `<dir>`. |
| `-l`, `--[no-]local` | Skip objects borrowed from alternates. |
| `-f` | Recompute all deltas. Slower, maybe smaller. |
| `-F` | Recompress every object. Slowest. |
| `-q`, `--[no-]quiet` | No progress. |
| `-n` | Skip `git update-server-info` (dumb HTTP/FTP files). |
| `--[no-]window=<n>` | Objects tried as delta bases (default 10). Bigger: smaller, slower. |
| `--[no-]depth=<n>` | Max delta chain (default 50, max 4095). Longer: smaller, slower reads. |
| `--[no-]threads=<n>` | Delta search threads. |
| `--[no-]window-memory=<n>` | Cap delta-search memory per thread. `0`: no cap. Default: `pack.windowMemory`. |
| `--max-pack-size=<n>` | Split into packs of at most `<n>` bytes (min 1 MiB). Blocks bitmaps. Default: `pack.packSizeLimit` or unlimited. |
| `--[no-]filter=<filter-spec>` | Move matching objects (e.g. `blob:limit=1m`) into a separate pack. Best in a bare repo with `-a -d --no-write-bitmap-index`. |
| `--drop-filtered` | (2.56+) Partial clones: delete `--filter` matches to free disk; re-downloaded on demand. Only `--filter=blob:limit=<n>`. Needs `-a` and a promisor remote; implies `-d`. Not with `--filter-to` or `-b`. Refuses mid merge/rebase/am/cherry-pick/revert/bisect; keeps blobs the index uses (non-bare). |
| `--dry-run` | (2.56+) With `--drop-filtered`: list object IDs that would drop; change nothing. |
| `--[no-]filter-to=<dir>` | With `--filter`: write the filtered pack under `<dir>`. If it becomes unreachable, the repo is corrupt. |
| `-b`, `--[no-]write-bitmap-index` | Write a reachability bitmap (faster serving). Needs `-a`, `-A`, or `-m`. Overrides `repack.writeBitmaps`. |
| `--[no-]pack-kept-objects` | Also repack objects in `.keep` packs (may duplicate). Mainly for bitmaps. |
| `--[no-]keep-pack=<pack-name>` | Treat this pack (e.g. `pack-123.pack`) as kept. Repeatable. |
| `--[no-]unpack-unreachable=<when>` | With `-A`: don't loosen unreachable objects older than `<when>`. |
| `-k`, `--[no-]keep-unreachable` | With `-ad`: put unreachable objects in the new pack instead. |
| `-i`, `--[no-]delta-islands` | No deltas across ref groups. For servers hosting forks. |
| `-g<factor>`, `--[no-]geometric=<factor>` | Merge just enough small packs that each has `<factor>`x the objects of the next smaller. Cheap, incremental. |
| `-m`, `--[no-]write-midx[=<mode>]` | Write a [multi-pack index](multi-pack-index.md). `default`: one for all packs. `incremental`: add a layer to a chain. |
| `--[no-]name-hash-version=<n>` | Path grouping for delta search (see [`git pack-objects`](pack-objects.md)). |
| `--[no-]path-walk` | Find delta candidates by path walk (see [`git pack-objects`](pack-objects.md)). |
