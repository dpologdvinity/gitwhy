# git unpack-objects

Read a pack from stdin and write each object as a loose file. Skips objects already present.

**Usage:** `git unpack-objects [<flags>] < <file.pack>`

## Flags

| Flag | What it does |
|------|--------------|
| `-n` | Dry run: check the pack, write nothing. |
| `-q` | No progress. |
| `-r` | On a corrupt pack, save as many objects as possible. |
| `--strict` | Refuse objects with broken content or links. |
| `--max-input-size=<size>` | Fail if the pack exceeds `<size>`. |
