# git symbolic-ref

Read, set, or delete a symbolic ref (a ref pointing to another ref, like `HEAD` holding `ref: refs/heads/<branch>`).

- One argument: print its target (`git symbolic-ref HEAD` = current branch).
- Two arguments: point `<name>` at `<ref>`.
- `--delete <name>`: remove it.

Exit: 0 ok, 1 if not symbolic (e.g. detached HEAD), 128 other errors.

**Usage:** `git symbolic-ref [-m <reason>] <name> <ref>`, `git symbolic-ref [-q] [--short] [--no-recurse] <name>`, `git symbolic-ref --delete [-q] <name>`

## Flags

| Flag | What it does |
|------|--------------|
| `-d`, `--[no-]delete` | Delete `<name>`. |
| `-q`, `--[no-]quiet` | No error message if not symbolic; still exits non-zero. |
| `--[no-]short` | Short name (`main`, not `refs/heads/main`). |
| `--[no-]recurse` | Follow chains of symbolic refs to the end (default). `--no-`: one step. |
| `-m <reason>` | Reflog reason. Setting only. |
