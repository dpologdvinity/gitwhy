# git notes

Attach text to commits (or other objects) without changing their IDs. `git log` shows them under `Notes:`. Stored in a separate ref (`refs/notes/commits` by default) with its own history. Push and fetch `refs/notes/*` to share.

**Usage:** `git notes [--ref <notes-ref>] [<subcommand>] [<flags>] [<object>]`

`<object>` defaults to `HEAD`. No subcommand: `list`.

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]ref=<notes-ref>` | Use this notes ref (`foo` = `refs/notes/foo`). Overrides `GIT_NOTES_REF` and `core.notesRef`. Goes before the subcommand. |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `list` | Print an object's note, or all as `<note-blob> <object>` lines. |
| `add` | Add a note. Fails if one exists (unless `-f`); with no message flags, edits the existing one. |
| `copy` | Copy a note to another object. |
| `append` | Append to a note, or create it. |
| `edit` | Edit a note in your editor. |
| `show` | Print a note. |
| `merge` | Merge another notes ref into the current one. |
| `remove` | Delete notes. |
| `prune` | Delete notes on missing objects. |
| `get-ref` | Print the notes ref in use. |

### git notes list

No flags. `git notes list [<object>]`.

### git notes add

`git notes add [<flags>] [<object>]`. Multiple `-m`/`-F` become paragraphs.

| Flag | What it does |
|------|--------------|
| `-m`, `--message=<msg>` | Note text. Repeatable. |
| `-F`, `--file=<file>` | Text from a file (`-` = stdin). Not binary-safe; use `-C`. |
| `-C`, `--reuse-message=<object>` | Use a blob as-is (implies `--no-stripspace`). |
| `-c`, `--reedit-message=<object>` | Like `-C`, then edit. |
| `-e`, `--[no-]edit` | Edit `-m`/`-F` text before saving. |
| `-f`, `--[no-]force` | Replace an existing note. |
| `--[no-]allow-empty` | Keep an empty note instead of deleting it. |
| `--[no-]separator[=<text>]` | Text between paragraphs. Default: blank line. `--no-`: nothing. |
| `--[no-]stripspace` | Tidy whitespace and blank lines. Default on, except with `-C`. |

### git notes copy

`git notes copy [-f] (<from> [<to>] | --stdin)`. Fails if `<to>` has a note or `<from>` has none.

| Flag | What it does |
|------|--------------|
| `-f`, `--[no-]force` | Overwrite `<to>`'s note. |
| `--[no-]stdin` | Read `<from> <to>` lines from stdin (`post-rewrite` format). No command-line objects. |
| `--[no-]for-rewrite=<command>` | With `--stdin`: copy as `amend` or `rebase` would, per `notes.rewrite.*`. For hooks. |

### git notes append

`git notes append [<flags>] [<object>]`. Like `add` minus `-f`; text goes after a separator.

| Flag | What it does |
|------|--------------|
| `-m`, `--message=<msg>` | Text to append. Repeatable. |
| `-F`, `--file=<file>` | Text from a file (`-` = stdin). |
| `-C`, `--reuse-message=<object>` | Append a blob as-is. |
| `-c`, `--reedit-message=<object>` | Like `-C`, then edit. |
| `-e`, `--[no-]edit` | Edit before appending. |
| `--[no-]allow-empty` | Keep the note even if empty. |
| `--[no-]separator[=<text>]` | Text between old note and new paragraphs. |
| `--[no-]stripspace` | Tidy whitespace (see `add`). |

### git notes edit

`git notes edit [--allow-empty] [<object>]`

| Flag | What it does |
|------|--------------|
| `--[no-]allow-empty` | Keep the note even if emptied. |
| `--[no-]stripspace` | Tidy whitespace (see `add`). |
| `-m`, `-F`, `-c`, `-C`, `-e`, `--[no-]separator` | Deprecated here; use `git notes add -f`. |

### git notes show

No flags. `git notes show [<object>]`.

### git notes merge

`git notes merge [-v | -q] [-s <strategy>] <notes-ref>`. Merge `<notes-ref>` into the current one. Manual conflicts go to `.git/NOTES_MERGE_WORKTREE`.

| Flag | What it does |
|------|--------------|
| `-s`, `--[no-]strategy=<strategy>` | `manual` (default), `ours`, `theirs`, `union`, `cat_sort_uniq` (join, sort, dedupe). Overrides `notes.mergeStrategy`. |
| `--commit` | Finish a conflicted merge after fixing files. |
| `--abort` | Cancel a conflicted merge. |
| `-v`, `--[no-]verbose` | More output. |
| `-q`, `--[no-]quiet` | Less output. |

### git notes remove

`git notes remove [<flags>] [<object>...]`

| Flag | What it does |
|------|--------------|
| `--[no-]ignore-missing` | Don't error on objects without notes. |
| `--[no-]stdin` | Also read objects from stdin. |

### git notes prune

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]dry-run` | List only. |
| `-v`, `--[no-]verbose` | List each removed note's object. |

### git notes get-ref

No flags.
