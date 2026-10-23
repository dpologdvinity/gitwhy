# git sh-setup

Not a command: a shell library git's scripts source to find the repo and get helper functions. Sets (doesn't export) `GIT_DIR` and `GIT_OBJECT_DIRECTORY`.

Before sourcing, set `USAGE` (optionally `LONG_USAGE`) for `usage`, and `SUBDIRECTORY_OK` to allow running from a subfolder.

**Usage (inside a script):** `. "$(git --exec-path)/git-sh-setup"`

## Functions

No flags. Defines:

| Function | What it does |
|----------|--------------|
| `die` | Print message to stderr and exit. |
| `usage` | `die` with the usage message. |
| `set_reflog_action` | Set `GIT_REFLOG_ACTION` unless already set. |
| `git_editor` | Open a file in the user's editor (`GIT_EDITOR`, `core.editor`, `VISUAL`, `EDITOR`). Fails if none set and terminal is dumb. |
| `is_bare_repository` | Print `true` or `false`. |
| `cd_to_toplevel` | `cd` to the working tree top. |
| `require_work_tree` | Exit unless inside the working tree. |
| `require_work_tree_exists` | Exit unless the repo has a working tree. Use before `cd_to_toplevel`. |
| `require_clean_work_tree <action> [<hint>]` | Exit with `Cannot <action>: <reason>. <hint>` if tracked files or index have uncommitted changes. |
| `get_author_ident_from_commit` | Print shell code to `eval` setting `GIT_AUTHOR_NAME`, `GIT_AUTHOR_EMAIL`, `GIT_AUTHOR_DATE` from a commit. |
| `create_virtual_base` | Shrink the first file to lines shared with the second. Fake base for a 3-way merge. |
