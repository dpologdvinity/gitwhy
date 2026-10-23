# git backfill

Experimental: download blobs a blobless partial clone (`--filter=blob:none`) skipped, in large batches grouped by path. Faster than fetching one by one (e.g. during `git blame`).

**Usage:** `git backfill [<flags>] [<revision-range>]`

`<revision-range>` limits which commits' blobs to fetch. Default `HEAD`.

## Flags

Also takes commit-limiting [revision options](../shared/revision-options.md) (`--first-parent`, `--since`, pathspecs, ...). Of `--filter=<spec>`, only `sparse:<oid>` is useful.

| Flag | What it does |
|------|--------------|
| `--min-batch-size=<n>` | Minimum objects per request. Default 50,000. |
| `--[no-]sparse` | Only blobs inside the sparse-checkout. On by default when sparse-checkout is enabled. |
| `--[no-]include-edges` | Also fetch boundary-commit blobs (the `A` in `A..B`), needed by `git log -p A..B` or `git replay`. On by default. |
