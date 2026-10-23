# git citool

GUI to stage files, write a message, and make one commit. Exits nonzero if closed without committing. Same as `git gui citool`; see [git gui](gui.md).

Often a separate package (`git-gui` on Debian/Ubuntu). Needs Tcl/Tk.

**Usage:** `git citool`

## Flags

| Flag | What it does |
|------|--------------|
| `--amend` | Amend the last commit. |
| `--nocommit` | Don't commit; exit 0 when done. Still refuses with unresolved conflicts (handy as a conflict-resolving GUI). |
