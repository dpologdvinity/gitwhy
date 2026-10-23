# git verify-pack

Check pack files for corruption. Give the `.idx`; checks it and its `.pack`. Exits non-zero on problems.

**Usage:** `git verify-pack [-v] [-s] [--] <pack>.idx...`

## Flags

| Flag | What it does |
|------|--------------|
| `<pack>.idx...` | Pack indexes to check. |
| `-v`, `--[no-]verbose` | Also list objects (`<id> <type> <size> <size-in-pack> <offset>`, plus `<depth> <base-id>` for deltas) and a delta-chain histogram. |
| `-s`, `--[no-]stat-only` | Skip the check; print only the histogram. |
| `--[no-]object-format=<hash>` | `sha1` or `sha256`. Needed outside a repo. |
| `--` | Rest are file names. |
