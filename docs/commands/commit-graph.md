# git commit-graph

Write or verify the commit-graph: a history cache that speeds up `log`, `merge-base`, and path history. `git gc` and `git fetch` usually maintain it. Disabled with replace objects or grafts.

**Usage:** `git commit-graph (write | verify) [<flags>]`

## Flags

For both subcommands.

| Flag | What it does |
|------|--------------|
| `--[no-]object-dir <dir>` | Objects folder to use (graph in `<dir>/info`). Must be a known object folder. |
| `--[no-]progress` | Show progress. Default: when stderr is a terminal. |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `write` | Write a commit-graph. Warns and does nothing if `core.commitGraph` is off. |
| `verify` | Check the graph for corruption. |

### git commit-graph write

Default input: commits in all packs. `--reachable`, `--stdin-packs`, `--stdin-commits` are mutually exclusive.

| Flag | What it does |
|------|--------------|
| `--reachable` | Commits reachable from all refs. |
| `--stdin-packs` | Only pack indexes named on stdin. |
| `--stdin-commits` | Walk from commit IDs on stdin. Non-commits skipped; bad IDs are errors. |
| `--append` | Keep commits already in the graph. |
| `--[no-]changed-paths` | Store changed-path Bloom filters (faster `git log -- <path>`). Slow on big repos. Stays on in later writes until `--no-changed-paths`. |
| `--[no-]max-new-filters=<n>` | Compute at most `<n>` new filters (`-1` = no limit), new layer only. Overrides `commitGraph.maxNewFilters`. |
| `--split[=<strategy>]` | Write a chain of graph files; new commits go in a tip file merged per the rules below. `no-merge`: never merge. `replace`: rewrite as one file. |
| `--size-multiple=<X>` | With `--split`: merge into the previous tip if `X` times the new commit count exceeds it. Default 2. |
| `--max-commits=<M>` | With `--split`: merge into the previous tip if the new one would exceed `M` commits. |
| `--expire-time=<datetime>` | With `--split`: delete unused graph files older than this. Default: now. |

### git commit-graph verify

| Flag | What it does |
|------|--------------|
| `--shallow` | Only check the tip file of a split chain. |
