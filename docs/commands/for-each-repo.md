# git for-each-repo

Experimental: run a git command in each repo listed in a config variable. Used by `git maintenance`.

Example: `git for-each-repo --config=maintenance.repo maintenance run` runs `git -C <repo> maintenance run` per path.

**Usage:** `git for-each-repo --config=<config> [--keep-going] [--] <arguments>`

## Flags

| Flag | What it does |
|------|--------------|
| `--config=<config>` | Config variable listing absolute repo paths (multi-valued). Read from system, global, and local config. |
| `--[no-]keep-going` | Continue past failures; exit 1 if any failed. Default: stop at first failure with its exit code. |
| `<arguments>` | Git command to run, without `git`. Prefix `--` if it could look like a flag here. |
