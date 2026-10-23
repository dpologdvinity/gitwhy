# git pack-refs

Move loose refs (one file per branch/tag) into `.git/packed-refs`, saving space and time with many refs. Later updates still write loose files. Same as `git refs optimize`; `git gc` runs it.

**Usage:** `git pack-refs [<flags>]`

Default: packs tags and already-packed refs, not other branches.

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]all` | Pack all refs (except hidden, broken, symbolic). |
| `--[no-]prune` | Delete loose ref files after packing. Default on. |
| `--[no-]auto` | Pack only if needed. Files backend: loose refs vs. `packed-refs` size. Reftable: compact so each table is at least twice the next newer one. |
| `--[no-]include=<pattern>` | Pack refs matching glob. Repeatable. Drops default tag inclusion. No effect with `--all`. `--no-include` clears. |
| `--[no-]exclude=<pattern>` | Don't pack matching refs; beats `--include`. Repeatable. Doesn't unpack packed refs. `--no-exclude` clears. |
