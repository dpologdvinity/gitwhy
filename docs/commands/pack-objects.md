# git pack-objects

Plumbing: build a packfile (compressed objects, many stored as deltas). Reads object IDs (or revisions with `--revs`) from stdin. Writes `<base-name>-<hash>.pack` and `.idx` and prints the hash, or streams to stdout. Run for you by `repack`, `gc`, `push`, and fetch servers.

**Usage:** `git pack-objects [<flags>] (--stdout | <base-name>) < <object-list>`

## Flags

### Input and output

| Flag | What it does |
|------|--------------|
| `<base-name>` | Write `<base-name>-<hash>.pack` and `.idx`. |
| `--[no-]stdout` | Write the pack to stdout. |
| `--[no-]revs` | Stdin is revisions (`rev-list --objects` style, plus `--not` and `--shallow <id>` lines); pack all they reach. |
| `--unpacked` | Implies `--revs`. Only unpacked objects. |
| `--all` | Implies `--revs`. Add everything reachable from any ref. |
| `--reflog` | Implies `--revs`. Add objects reachable from reflogs. |
| `--indexed-objects` | Implies `--revs`. Add objects the index uses. |
| `--[no-]stdin-packs[=<mode>]` | Stdin lists packs (`pack-1234.pack`): pack their objects, minus `^`-prefixed packs. `follow`: also add reachable objects from unlisted packs; `!` packs excluded but not assumed complete. Not with `--revs` (except `--unpacked`). |
| `--[no-]include-tag` | Add annotated tags of packed objects. |
| `--[no-]non-empty` | Don't write an empty pack. |
| `--max-pack-size=<n>` | Split into packs of at most `<n>` bytes (`k`/`m`/`g`; min 1 MiB). Default: unlimited or `pack.packSizeLimit`. Can bloat and slow the repo. |
| `--index-version=<version>[,<offset>]` | For tests: force `.idx` version and 64-bit entries past `<offset>`. |

### Which objects to skip

| Flag | What it does |
|------|--------------|
| `--[no-]incremental` | Skip objects already packed. |
| `--[no-]local` | Skip objects from alternates. |
| `--[no-]honor-pack-keep` | Skip objects in local `.keep` packs. |
| `--[no-]keep-pack=<name>` | Skip objects in this pack (file name only). Repeatable. |
| `--[no-]filter=<filter-spec>` | Omit objects (e.g. `blob:none`), as `git rev-list --filter`. `--no-filter` cancels. |
| `--missing=<action>` | Debug: `error` (default; partial clones fetch first), `allow-any` (skip, no fetch), `allow-promisor` (skip only promisor-expected ones). |
| `--[no-]exclude-promisor-objects` | Skip promisor-remote objects. |
| `--[no-]exclude-promisor-objects-best-effort` | Same, tolerating missing objects; implies `--missing=allow-any`. |
| `--[no-]keep-true-parents` | Also pack parents hidden by grafts. |
| `--[no-]uri-protocol=<protocol>` | Server: skip blobs `uploadpack.blobPackfileUri` serves over this protocol. |

### Unreachable objects

| Flag | What it does |
|------|--------------|
| `--[no-]keep-unreachable` | Implies `--revs`. Keep unreachable objects from repacked packs. |
| `--[no-]pack-loose-unreachable` | Implies `--revs`. Pack unreachable loose objects (deleting the loose copies). |
| `--[no-]unpack-unreachable[=<time>]` | Implies `--revs`. Write unreachable objects newer than `<time>` as loose files instead. |
| `--[no-]cruft` | Write a cruft pack (with `.mtimes`) of unreachable objects. Stdin: kept packs and deleted packs (`-` prefix). Used by `git repack --cruft`. Not with `--stdin-packs`, the flags above, or `--revs`. |
| `--[no-]cruft-expiration[=<time>]` | With `--cruft`: drop objects older than `<time>` unless reached by newer kept ones. Default: drop nothing. |

### Delta compression

| Flag | What it does |
|------|--------------|
| `--[no-]window=<n>` | Delta candidates per object (default 10). Bigger: smaller pack, slower. |
| `--[no-]depth=<n>` | Max delta chain (default 50, max 4095). Deeper reads slower. |
| `--window-memory=<n>` | Cap window memory (`k`/`m`/`g`; `0` = unlimited). Default `pack.windowMemory`. |
| `--[no-]reuse-delta` | Reuse existing deltas (default). `--no-` recomputes them. |
| `--[no-]reuse-object` | Reuse compressed data (default). `--no-` recompresses all; implies `--no-reuse-delta`. |
| `--[no-]compression=<n>` | zlib level. Default `pack.compression`, then `core.compression`, then `-1`. |
| `--[no-]delta-base-offset` | Delta bases as offsets: 3-5% smaller, unreadable by ancient git. On in `gc`, `repack`, `bundle`. |
| `--[no-]threads=<n>` | Delta search threads; `0` = per CPU. Each needs its own window memory. |
| `--[no-]thin` | With `--stdout`: omit bases the receiver has. Needs `git index-pack --fix-thin` to be valid. |
| `--[no-]shallow` | Optimize for shallow clients. With `--thin`: smaller, slower. |
| `--[no-]sparse` | With `--revs`: only walk paths with new objects. Faster small pushes; may add a few extra objects. Default `pack.useSparse` (true). |
| `--[no-]delta-islands` | Never delta across "islands" (`pack.island` ref groups), so shared-storage forks reuse deltas. Use a large `--window`. |
| `--[no-]name-hash-version=<n>` | Path hash for grouping. `1` (default) weights path end; required for bitmaps. `2` mixes in folders, avoiding same-name collisions. |
| `--[no-]path-walk` | Group by path, then compress across paths. Before 2.56: not with `--delta-islands`, ignores `--use-bitmap-index`; 2.56+ works with both, using bitmaps where enough. Filters: `blob:none`, `blob:limit=<n>`, `tree:0`, `object:type=<type>`, `sparse:<oid>`, `combine:`. |

### Bitmaps

| Flag | What it does |
|------|--------------|
| `--[no-]use-bitmap-index` | Use reachability bitmaps to count faster. |
| `--[no-]write-bitmap-index` | Write a reachability bitmap. |

### Progress

| Flag | What it does |
|------|--------------|
| `-q`, `--[no-]quiet` | No progress. |
| `--[no-]progress` | Force progress off a terminal. |
| `--[no-]all-progress` | With `--stdout`: also show write progress. |
| `--[no-]all-progress-implied` | Enable `--all-progress` whenever progress shows, without forcing it. |

No deltas for files over `core.bigFileThreshold` or with `-delta` attribute.
