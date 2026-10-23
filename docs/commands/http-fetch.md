# git http-fetch

Low-level: download objects over "dumb" HTTP. Git runs it for you; use [`git fetch`](fetch.md).

**Usage:** `git http-fetch [<flags>] [--stdin | --packfile=<hash> | <commit>] <URL>`

## Flags

| Flag | What it does |
|------|--------------|
| `<commit>` | Commit ID or ref name under `<URL>/refs/` to download. |
| `<URL>` | Remote repo address. |
| `-a`, `-c`, `-t` | Ignored. |
| `-d` | Ignored. |
| `-v` | Print what's downloaded. |
| `-w <filename>` | Write the commit ID to `$GIT_DIR/refs/<filename>` when done. |
| `--stdin` | Read targets from stdin: `<commit>` per line, optional tab and `-w` file name. |
| `--recover` | Verify everything reachable is present. Use after an interrupted fetch. |
| `--packfile=<hash>` | Internal: download the pack at `<URL>` and run `git index-pack` on it. `<hash>` only names the temp file. Needs `--index-pack-args`. |
| `--index-pack-args=<args>` | Internal: `index-pack` command for `--packfile`, URL-encoded words separated by spaces. |
