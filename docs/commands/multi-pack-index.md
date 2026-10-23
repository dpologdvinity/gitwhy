# git multi-pack-index

Plumbing: manage the multi-pack-index (MIDX), one index over many packs for faster lookups. `git repack`, `git gc`, and `git maintenance` usually handle it.

**Usage:** `git multi-pack-index [<flags>] <subcommand> [<subcommand-flags>]`

## Flags

Go before the subcommand.

| Flag | What it does |
|------|--------------|
| `--[no-]object-dir=<dir>` | Use objects in `<dir>` (must be an alternate). Reads `<dir>/packs/`. |
| `--[no-]progress` | Force progress on/off. Default: on for a terminal. For `write`, `verify`, `expire`, `repack`. |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `write` | Write a new MIDX. |
| `compact` | Merge a range of incremental MIDX layers into one. |
| `verify` | Check the MIDX. |
| `expire` | Delete covered packs no longer used, then rewrite the MIDX. |
| `repack` | Combine small packs into one and update the MIDX. |

### git multi-pack-index write

| Flag | What it does |
|------|--------------|
| `--preferred-pack=<pack>` | Prefer this pack's copy of duplicate objects (else oldest wins). Must hold an object. |
| `--[no-]bitmap` | Also write a bitmap (faster clone/fetch). |
| `--[no-]stdin-packs` | Only index packs named on stdin (`.idx` basenames). |
| `--refs-snapshot=<path>` | With `--bitmap`: file of ref-tip IDs taken before repacking. `+` prefix marks a preferred tip. |
| `--[no-]incremental` | Write a new layer with only packs not yet in the MIDX. |
| `--[no-]write-chain-file` | With `--incremental`: `--no-write-chain-file` skips the chain file and prints the layer checksum. |
| `--base=<checksum>` | Layer to build on; `none` for root. Needs `--no-write-chain-file`. |

### git multi-pack-index compact

**Usage:** `git multi-pack-index compact [<flags>] <from> <to>`

Merge layers `<from>` through `<to>` (checksums) into one. Writes a version-2 MIDX, unreadable by git before 2.54.

| Flag | What it does |
|------|--------------|
| `--[no-]incremental` | Write into the MIDX chain, not stand-alone. |
| `--[no-]bitmap` | Also write a bitmap. |
| `--[no-]write-chain-file` | With `--incremental`: `--no-write-chain-file` skips the chain file and prints the checksum. |
| `--base=<checksum>` | Base layer instead of `<from>`'s parent; `none` for none. |

### git multi-pack-index verify

No extra flags.

### git multi-pack-index expire

No extra flags. Keeps `.keep` and cruft packs. Not for incremental MIDX.

### git multi-pack-index repack

Not for incremental MIDX. Run `expire` after to delete old packs.

| Flag | What it does |
|------|--------------|
| `--batch-size=<size>` | `0`: repack everything into one pack. Else pick old small packs up to `<size>` total; no-op if only one. Skips `.keep` packs unless `repack.packKeptObjects`. |
