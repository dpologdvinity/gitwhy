# git verify-tag

Check signatures (GPG, SSH, X.509) on signed tags. Exits non-zero if bad or missing.

**Usage:** `git verify-tag [-v | --verbose] [--format=<format>] [--raw] <tag>...`

## Flags

| Flag | What it does |
|------|--------------|
| `-v`, `--[no-]verbose` | Print the tag contents first. |
| `--[no-]raw` | Print the signing tool's raw status to stderr. |
| `--[no-]format=<format>` | Print each tag with `%(fieldname)` placeholders, as in `git for-each-ref --format`. |
| `<tag>...` | Tags to check. |
