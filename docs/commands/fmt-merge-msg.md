# git fmt-merge-msg

Mostly internal: print the default merge message (`Merge branch 'x' of ...`) from a `FETCH_HEAD`-format list on stdin. Used by scripts calling `git merge`.

Example: `git fetch origin main && git fmt-merge-msg --log <.git/FETCH_HEAD`

**Usage:** `git fmt-merge-msg [-m <message>] [--into-name <branch>] [--log[=<n>] | --no-log] [-F <file>]`

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]log[=<n>]` | Also list merged commit subjects, up to `<n>` per branch (default 20). Overrides `merge.log`. |
| `--[no-]summary` | Deprecated: `--[no-]log`. |
| `-m`, `--[no-]message=<message>` | Use as the title instead of "Merge branch ...". For use with `--log`. |
| `--[no-]into-name=<branch>` | Word it as merging into `<branch>`. |
| `-F`, `--[no-]file=<file>` | Read the list from this file instead of stdin. |
