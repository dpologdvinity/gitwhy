# git count-objects

Count loose (unpacked) objects and their disk use. Helps decide when to repack.

**Usage:** `git count-objects [-v] [-H | --human-readable]`

## Flags

| Flag | What it does |
|------|--------------|
| `-v`, `--[no-]verbose` | Full report: `count`, `size` (loose objects), `in-pack`, `packs`, `size-pack`, `prune-packable` (loose objects also packed; remove with `git prune-packed`), `garbage`, `size-garbage` (stray files), `alternate` (borrowed object stores). Sizes in KiB. |
| `-H`, `--[no-]human-readable` | Sizes in KiB/MiB/GiB. |
