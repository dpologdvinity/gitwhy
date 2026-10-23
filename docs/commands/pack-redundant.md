# git pack-redundant

Deprecated, slated for removal; use [`git gc`](gc.md) or [`git repack`](repack.md). Finds pack files whose objects all exist in other packs. Fails without `--i-still-use-this`.

Prints redundant pack names (pipe to `xargs rm` from the repo root). Object IDs on stdin count as unneeded, e.g. to drop packs of only unreachable objects:

`git fsck --full --unreachable | cut -d ' ' -f3 | git pack-redundant --all | xargs rm`

**Usage:** `git pack-redundant [--verbose] [--alt-odb] (--all | <pack-filename>...)`

## Flags

| Flag | What it does |
|------|--------------|
| `<pack-filename>...` | Packs to check. |
| `--all` | Check every pack; ignore given names. |
| `--alt-odb` | Objects in alternate object stores needn't also be in a local pack. |
| `--verbose` | Print stats to stderr. Slower. |
| `--i-still-use-this` | Required to run. |
