# git merge-tree

Merge two branches without touching the working tree, index, or refs; print the result tree ID. Same logic as `git merge`. For scripts; pair with `git commit-tree` and `git update-ref` for a real merge commit.

**Usage:** `git merge-tree [--write-tree] [<flags>] <branch1> <branch2>`

Old form (deprecated): `git merge-tree [--trivial-merge] <base-tree> <branch1> <branch2>`. Trivial merge only (no content merging or renames); diff-like output.

**Output:** clean: the tree ID. Conflicts: tree ID (files may have markers), one `<mode> <object> <stage> <filename>` line per conflicted stage, blank line, messages. With `--stdin`: each result starts with status (`1` clean, `0` conflicts) + NUL and ends with NUL.

**Exit code:** 0 clean, 1 conflicts, other = error. With `--stdin`: 0 either way.

Trust the exit code: some conflicts list no files and many leave no markers.

## Flags

| Flag | What it does |
|------|--------------|
| `--write-tree` | Real merge (default). |
| `--trivial-merge` | Old trivial mode. Deprecated. |
| `--[no-]messages` | Print info messages at the end. Default: only on conflict. |
| `--quiet` | No output, exit code only. Faster: stops at the first conflict. |
| `-z` | NUL-terminated, unquoted names; messages as `<count>NUL<path>NUL...<type>NUL<message>NUL`. |
| `--name-only` | List conflicted file names only, once each. |
| `--allow-unrelated-histories` | Merge branches with no shared history. |
| `--stdin` | One merge per input line: `[<base-commit> -- ]<branch1> <branch2>`. Implies `-z`. |
| `--[no-]merge-base=<tree-ish>` | Use this merge base; branches may be trees. Single base only: may differ from `git merge` (even lose changes). Not with `--stdin`. |
| `-X`, `--[no-]strategy-option=<option>` | Strategy option. See [merge strategy options](merge.md#strategy-options--x). |
