# git show-index

Plumbing: read a pack index (`.idx`) from stdin and print each object's pack offset, ID, and (v2+) CRC32.

**Usage:** `git show-index [--object-format=<hash>] < <pack-idx-file>`

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]object-format=<hash>` | Index hash: `sha1` or `sha256`. Default: the repo's, or `sha1` outside a repo. |
