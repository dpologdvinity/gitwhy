# git verify-commit

Check commits' GPG signatures (from `git commit -S`). Fails if any is missing or bad.

**Usage:** `git verify-commit [<flags>] <commit>...`

## Flags

| Flag | What it does |
|------|--------------|
| `-v`, `--[no-]verbose` | Print the raw commit first. |
| `--[no-]raw` | Print gpg's raw status to stderr instead of the summary. |
