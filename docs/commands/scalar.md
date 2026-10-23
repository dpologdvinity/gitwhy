# scalar

Keep very large repos fast: tuned settings, background maintenance, partial downloads. Run as `scalar`, not `git scalar`. Ships with git on most systems.

Works on an **enlistment**: a folder whose `src/` is the working tree (untracked build output can live outside it). A registered repo not in `src` is its own enlistment. All subcommands except `clone`, `list`, and `reconfigure --all` need an enlistment.

**Usage:** `scalar [-C <directory>] [-c <key>=<value>] <subcommand> [<options>]`

## Flags

Before the subcommand.

| Flag | What it does |
|------|--------------|
| `-C <directory>` | `cd` first. Same as `git -C`. |
| `-c <key>=<value>` | One-run config. Same as `git -c`. |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `clone` | Clone into a new enlistment with scalar settings. |
| `list` | List enlistments. |
| `register` | Add an existing repo and start background maintenance. |
| `unregister` | Remove a repo and stop its maintenance. |
| `run` | Run one maintenance task now. |
| `reconfigure` | Reapply scalar settings. |
| `diagnose` | Zip logs and stats for a bug report. |
| `delete` | Delete an enlistment from disk (unpushed work too) and unregister it. |
| `help` | Open this manual page. |
| `version` | Print version. |

### scalar clone

**Usage:** `scalar clone [<options>] <url> [<enlistment>]`

Downloads commits and trees only; file contents on demand. Working tree in `<enlistment>/src`. Sparse checkout starts with top-level files only; widen with `git sparse-checkout set <dirs>` or `disable`. Peek with `git ls-tree HEAD:<dir>`.

| Flag | What it does |
|------|--------------|
| `-b <name>`, `--branch <name>` | Check out this branch, not the remote default. |
| `--[no-]single-branch` | Only fetch that one branch, now and later. |
| `--[no-]src` | Use `<enlistment>/src` (default). `--no-src`: `<enlistment>` itself. |
| `--[no-]tags` | Fetch tags (default). `--no-tags` skips them in future fetches too. |
| `--[no-]full-clone` | No sparse checkout. |
| `--[no-]maintenance` | Enable background maintenance (default). |

### scalar list

**Usage:** `scalar list`

No flags. Runs anywhere.

### scalar register

**Usage:** `scalar register [--[no-]maintenance] [<enlistment>]`

Default: current enlistment. Inside `src/`, the parent becomes the enlistment.

| Flag | What it does |
|------|--------------|
| `--[no-]maintenance` | Enable background maintenance (default). `--no-` doesn't disable maintenance set up otherwise. |

### scalar unregister

**Usage:** `scalar unregister [<enlistment>]`

No flags.

### scalar run

**Usage:** `scalar run <task> [<enlistment>]`

Rarely needed; maintenance runs these on schedule. No flags. Tasks:

| Task | What it does |
|------|--------------|
| `all` | All tasks below. |
| `config` | Apply recommended settings. |
| `commit-graph` | `git maintenance` `commit-graph`. |
| `fetch` | Background object download (`git maintenance` `prefetch`). |
| `loose-objects` | `git maintenance` `loose-objects`. |
| `pack-files` | Combine small packs (`git maintenance` `incremental-repack`). |

### scalar reconfigure

**Usage:** `scalar reconfigure [--maintenance=(enable|disable|keep)] [--all | <enlistment>]`

| Flag | What it does |
|------|--------------|
| `-a`, `--all` | All enlistments in `scalar.repo`. Run after upgrades. |
| `--maintenance=<mode>` | `enable` (default), `disable`, or `keep` background maintenance. |

### scalar diagnose

**Usage:** `scalar diagnose [<enlistment>]`

No flags. Writes a `.zip` in a folder next to `src`.

### scalar delete

**Usage:** `scalar delete <enlistment>`

No flags. Deletes the whole enlistment folder, including uncommitted and unpushed work.

### scalar help

**Usage:** `scalar help`

No flags. Same as `git help scalar`.

### scalar version

**Usage:** `scalar version [-v | --verbose] [--build-options]`

| Flag | What it does |
|------|--------------|
| `-v`, `--verbose` | Accepted; no change. |
| `--build-options` | Also print build info. |

## Settings scalar applies

`clone`, `register`, and `reconfigure` set these unless already set:

`am.keepCR=true`, `commitGraph.changedPaths=true`, `commitGraph.generationVersion=1`, `core.autoCRLF=false`, `core.logAllRefUpdates=true`, `core.safeCRLF=false`, `credential.https://dev.azure.com.useHttpPath=true`, `feature.experimental=false`, `feature.manyFiles=false`, `fetch.showForcedUpdates=false`, `fetch.unpackLimit=1`, `fetch.writeCommitGraph=false`, `gc.auto=0`, `gui.GCWarning=false`, `index.skipHash=true`, `index.threads=true`, `index.version=4`, `log.excludeDecoration=refs/prefetch/*`, `merge.renames=true`, `merge.stat=false`, `pack.useBitmaps=false`, `pack.usePathWalk=true`, `receive.autoGC=false`, `status.aheadBehind=false`. Also `core.untrackedCache` (true except on Windows), `http.sslBackend=schannel` (Windows), and `core.fsmonitor=true` where supported.
