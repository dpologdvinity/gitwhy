# git patch-id

Print an ID per patch read from stdin, based only on the changes (line numbers ignored). Same ID means almost surely the same change; finds duplicate commits. On `git diff-tree --patch` output, prints `<patch-id> <commit-id>`.

**Usage:** `git rev-list <range> | git diff-tree --patch --stdin | git patch-id [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `--stable` | File order doesn't affect the ID. Whitespace ignored. Not comparable with git 1.9 and older. Default if `patchid.stable` is true. |
| `--unstable` | File order matters (git 1.9 compatible). Whitespace ignored. Default. |
| `--verbatim` | Don't ignore whitespace. Implies `--stable`; conflicts with `--unstable`. Default if `patchid.verbatim` is true. |
