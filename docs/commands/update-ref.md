# git update-ref

Low-level: safely set or delete a ref (branch, tag, `HEAD`, ...), optionally checking its old value first.

- `git update-ref <ref> <new-id>`: point `<ref>` at `<new-id>`. `HEAD` updates the current branch. No fast-forward check: commits only on the old value drop off the branch (reflog keeps them).
- `git update-ref <ref> <new-id> <old-id>`: only if `<ref>` is `<old-id>`. 40 zeros or `""` requires the ref not exist.
- `git update-ref -d <ref> [<old-id>]`: delete (only if `<old-id>`, if given). No merged check; its reflog goes too.
- `--stdin`: apply many commands together.

For symbolic refs use [`git symbolic-ref`](symbolic-ref.md) or the `symref-*` commands below.

**Usage:** `git update-ref [<flags>] <ref> <new-id> [<old-id>]`, `git update-ref [<flags>] -d <ref> [<old-id>]`, `git update-ref [<flags>] --stdin [-z] [--batch-updates]`

## Flags

| Flag | What it does |
|------|--------------|
| `-m <reason>` | Reflog message. |
| `-d` | Delete the ref. |
| `--no-deref` | Overwrite a symbolic `<ref>` itself, not its target. |
| `--deref` | Follow symbolic refs (default). |
| `--[no-]create-reflog` | Always create a reflog. |
| `--[no-]stdin` | Read commands from stdin (below). All or nothing. |
| `-z` | With `--stdin`: NUL-separated, unquoted. |
| `-0`, `--[no-]batch-updates` | With `--stdin`: apply valid updates, report others as `rejected <old> <new> <reason>`. System errors (disk, memory) still fail all. |

## --stdin commands

One per line. Quote fields with spaces C-style. Zero ID: 40 zeros or empty; omit optional fields and their space. Any object name form works. Repeating a ref is an error.

| Command | What it does |
|---------|--------------|
| `update <ref> <new-id> [<old-id>]` | Set `<ref>`, checking `<old-id>` if given. Zero `<new-id>`: must not exist after. Zero `<old-id>`: must not exist before. |
| `create <ref> <new-id>` | Create; must not exist. `<new-id>` non-zero. |
| `delete <ref> [<old-id>]` | Delete, checking non-zero `<old-id>` if given. |
| `verify <ref> [<old-id>]` | Check only. Zero or missing `<old-id>`: must not exist. |
| `symref-update <ref> <new-target> [ref <old-target> \| oid <old-id>]` | Point symbolic `<ref>` at `<new-target>`, checking old target or ID if given. |
| `symref-create <ref> <new-target>` | Create symbolic ref; must not exist. |
| `symref-delete <ref> [<old-target>]` | Delete symbolic ref, checking target if given. |
| `symref-verify <ref> [<old-target>]` | Check symbolic ref's target; missing target: must not exist. Needs no-deref. |
| `option no-deref` | Don't follow symbolic refs for the next command. |
| `start` | Begin transaction. Input ending without `commit` aborts. |
| `prepare` | Lock queued refs; abort if any fails. |
| `commit` | Apply queued updates, end transaction. |
| `abort` | Cancel transaction, release locks. |

With `-z`, NUL replaces spaces between fields (after the command word) and ends the line, e.g. `update <ref>NUL<new-id>NUL<old-id>NUL`. Zero = 40 zeros; missing = empty field.

Each ref changes atomically, but readers may briefly see only some changes.

## Reflog

If `core.logAllRefUpdates` is true (for `refs/heads/`, `refs/remotes/`, `refs/notes/`, `HEAD`, etc.) or `$GIT_DIR/logs/<ref>` exists, each change appends `<old-id> <new-id> <committer> [TAB <reason>]`. Fails if the log can't be written or no committer identity is set.
