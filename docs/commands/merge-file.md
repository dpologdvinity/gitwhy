# git merge-file

Three-way merge of single files: apply `<base>`-to-`<other>` changes onto `<current>`, overwriting it. Conflicts get markers. Exit code: conflict count (max 127), 0 if clean, negative on error. No repo needed.

**Usage:** `git merge-file [<flags>] [-L <current-name> [-L <base-name> [-L <other-name>]]] <current> <base> <other>`

## Flags

| Flag | What it does |
|------|--------------|
| `-p`, `--[no-]stdout` | Print result; don't overwrite `<current>`. |
| `--[no-]object-id` | Arguments are blob IDs. Writes result as a blob and prints its ID (unless `-p`). |
| `-L <label>` | Conflict marker label instead of file name. Up to three, in argument order. |
| `-q`, `--[no-]quiet` | Don't warn about conflicts. |
| `--[no-]diff3` | Markers also show base. Default from `merge.conflictStyle`. |
| `--[no-]zdiff3` | Like `--diff3`, but moves lines common to both sides out of the conflict. |
| `--[no-]ours` | Resolve conflicts with `<current>`'s lines. |
| `--[no-]theirs` | Resolve conflicts with `<other>`'s lines. |
| `--[no-]union` | Resolve conflicts keeping both sides. |
| `--diff-algorithm=<algorithm>` | `myers` (default), `minimal`, `patience`, or `histogram` (fewer mismerges on lines like lone braces). |
| `--[no-]marker-size=<n>` | Conflict marker length (default 7). |
