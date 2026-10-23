# git maintenance

Run repo upkeep jobs (packing, commit-graph, prefetch, ...) now or on an hourly/daily/weekly background schedule.

**Usage:** `git maintenance <subcommand> [<flags>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `run` | Run tasks now. |
| `start` | `register`, then schedule `git maintenance run --schedule=<frequency>` hourly, daily, weekly for all registered repos. |
| `stop` | Remove the schedule. Repos stay registered. |
| `register` | Add this repo to global `maintenance.repo` and set safe task schedules. |
| `unregister` | Remove this repo from `maintenance.repo`. Scheduler keeps running. |
| `is-needed` | Exit 0 if maintenance is needed, 1 if not. Runs nothing. |

## Tasks

Names for `--task`. Default: tasks with `maintenance.<task>.enabled` (only `gc` by default), or those from `maintenance.strategy`.

| Task | What it does |
|------|--------------|
| `gc` | Full `git gc`: repack into one pack, delete stale data. Slow on big repos. Don't combine with `loose-objects`. |
| `commit-graph` | Update the commit-graph, safe alongside other git processes. |
| `prefetch` | Fetch all remotes into `refs/prefetch/` to speed later fetches. Doesn't move remote-tracking refs or tags. Skips `remote.<name>.skipFetchAll` remotes. |
| `loose-objects` | Delete already-packed loose objects, then pack a batch (`maintenance.loose-objects.batchSize`, default 50,000). |
| `incremental-repack` | Multi-pack-index: delete unused packs, merge small ones. |
| `geometric-repack` | `git repack --geometric`, keeping helper files current. |
| `pack-refs` | Pack loose refs into one file. |
| `reflog-expire` | Delete old reflog entries. |
| `rerere-gc` | Delete stale rerere resolutions. |
| `worktree-prune` | Delete records of missing worktrees. |

### git maintenance run

| Flag | What it does |
|------|--------------|
| `--[no-]auto` | Only tasks past their thresholds (e.g. `gc.auto`, `gc.autoPackLimit`). Not with `--schedule`. |
| `--[no-]schedule=<frequency>` | Only tasks whose `maintenance.<task>.schedule` is `hourly`, `daily`, or `weekly` and whose `maintenance.<task>.lastRun` is old enough. |
| `--[no-]detach` | Run in the background. Default: `maintenance.autoDetach`, then `gc.autoDetach`. |
| `--[no-]quiet` | No progress. |
| `--task=<task>` | Only this task. Repeatable; runs in given order. |

### git maintenance start

| Flag | What it does |
|------|--------------|
| `--scheduler=<scheduler>` | `auto` (default; Linux: `systemd-timer` if available, else `crontab`), `crontab`, `systemd-timer`, `launchctl` (macOS), `schtasks` (Windows). |

### git maintenance stop

No flags.

### git maintenance register

Also sets `maintenance.strategy=incremental` if unset (hourly `commit-graph`, `prefetch`; daily `loose-objects`, `incremental-repack`; no `gc`) and `maintenance.auto=false` in this repo (kept after `unregister`).

| Flag | What it does |
|------|--------------|
| `--[no-]config-file=<file>` | Use this config file instead of global. |

### git maintenance unregister

| Flag | What it does |
|------|--------------|
| `--[no-]config-file=<file>` | Use this config file instead of global. |
| `-f`, `--[no-]force` | Succeed even if not registered. |

### git maintenance is-needed

| Flag | What it does |
|------|--------------|
| `--[no-]auto` | Check `--auto` thresholds. |
| `--task=<task>` | Check only this task. Repeatable. |
| `--schedule` | In the usage line but rejected by 2.55. |
