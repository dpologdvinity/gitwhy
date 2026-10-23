# git gc

Clean up and compact the repo: pack objects, prune old unreachable ones, pack refs, expire reflogs, clean rerere and stale worktrees, update commit-graph. Git runs `git gc --auto` itself, so rarely needed by hand.

Objects are kept if referenced by branches, tags, remote-tracking branches, the index, reflogs, or any `refs/*`. Notes don't keep their target alive.

**Usage:** `git gc [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]aggressive` | Recompute all deltas harder for a smaller repo. Much slower. Uses `gc.aggressiveDepth` (50), `gc.aggressiveWindow` (250). |
| `--[no-]auto` | Only work if needed (see `gc.auto`, `gc.autoPackLimit`). Runs `pre-auto-gc` hook. |
| `--[no-]detach` | Run in background. Overrides `gc.autoDetach`. |
| `--[no-]prune[=<date>]` | Delete unreachable loose objects older than `<date>` (default 2 weeks, or `gc.pruneExpire`). On by default. `--prune=now` risks corruption if another git process is writing. |
| `--[no-]cruft` | Pack unreachable objects into a cruft pack instead of loose files. On by default. |
| `--max-cruft-size=<n>` | With `--cruft`: max bytes per new cruft pack. Overrides `gc.maxCruftSize`. |
| `--[no-]expire-to=<dir>` | With `--cruft`: back up pruned objects to a cruft pack in `<dir>`. |
| `-q`, `--[no-]quiet` | No progress. |
| `--[no-]force` | Run even if another `gc` may be running. |
| `--[no-]keep-largest-pack` | Repack all into one pack, except the largest pack, `.keep` packs, and cruft packs. Ignores `gc.bigPackThreshold`. |
