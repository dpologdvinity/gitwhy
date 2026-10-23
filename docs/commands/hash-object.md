# git hash-object

Print the object ID for a file's contents; optionally store it. File may be outside the repo.

**Usage:** `git hash-object [-t <type>] [-w] [--path=<file> | --no-filters] [--stdin [--literally]] [--] <file>...`
or `git hash-object [-t <type>] [-w] --stdin-paths [--no-filters]`

## Flags

| Flag | What it does |
|------|--------------|
| `-t <type>` | `blob` (default), `commit`, `tree`, or `tag`. |
| `-w` | Write the object to the database. |
| `--[no-]stdin` | Read content from stdin. |
| `--[no-]stdin-paths` | Read file names from stdin, one per line. |
| `--[no-]path=<file>` | Apply `<file>`'s filters (attributes, line endings). For temp files or stdin. |
| `--no-filters` | Hash raw bytes, no filters. Default for stdin without `--path`. |
| `--filters` | Undo `--no-filters`. |
| `--[no-]literally` | With `--stdin`: allow invalid objects. For testing. |
