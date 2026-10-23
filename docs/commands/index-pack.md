# git index-pack

Plumbing: build the `.idx` index (and optional `.rev`) for a packfile. Fetch and push run it. Prints the pack hash; with `--stdin` prefixed `pack<TAB>` or `keep<TAB>`.

**Usage:** `git index-pack [-v] [-o <index-file>] [--[no-]rev-index] <pack-file>`
or `git index-pack --stdin [--fix-thin] [--keep] [-v] [-o <index-file>] [--[no-]rev-index] [<pack-file>]`

## Flags

| Flag | What it does |
|------|--------------|
| `-v` | Verbose, with progress. |
| `-o <index-file>` | Index path. Default: pack name with `.idx` (pack must end in `.pack`). |
| `--[no-]rev-index` | Also write a `.rev`; with `--verify`, check it. Overrides `pack.writeReverseIndex`. |
| `--verify` | Check an existing pack against its index. |
| `--stdin` | Read the pack from stdin; save to `<pack-file>` or `objects/pack/`. Pair with `--keep`. |
| `--fix-thin` | With `--stdin`: add missing delta bases to a thin pack. |
| `--keep` | Create a `.keep` file first so `git repack` won't delete the pack. |
| `--keep=<msg>` | `--keep` with `<msg>` written in the file. |
| `--promisor[=<msg>]` | Create a `.promisor` file (partial clone), optionally with `<msg>`. Not with `<pack-file>`. |
| `--strict[=<msg-id>=<severity>,...]` | Fail on broken objects or links. List adjusts `fsck` checks, e.g. `missingEmail=ignore`. |
| `--fsck-objects[=<msg-id>=<severity>,...]` | Fail on broken objects; allow broken links. |
| `--check-self-contained-and-connected` | Internal. Fail on broken links. |
| `--threads=<n>` | Delta threads. `0` = auto (max 3). More uses more memory. |
| `--max-input-size=<size>` | Fail if the pack exceeds `<size>`. |
| `--object-format=<hash>` | `sha1` or `sha256`. Default: repo's. Not with `--stdin`. |
| `--progress-title=<title>` | Internal. Progress bar title. |
| `--index-version=<version>[,<offset>]` | Tests only. Force index version and 64-bit offset threshold. |
