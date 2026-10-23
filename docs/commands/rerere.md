# git rerere

"Reuse recorded resolution": remember conflict fixes and reapply them when the same conflict recurs. Enable with `git config rerere.enabled true`; `git merge`, `git rebase`, `git commit` then run it. Fixes files but doesn't stage them: check `git diff`, then `git add`. Finds conflicts by `<<<<<<<` markers.

**Usage:** `git rerere [<flags>] [clear | forget <pathspec>... | diff | status | remaining | gc]`

No subcommand: record new conflicts and fixes, apply known ones.

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]rerere-autoupdate` | Also stage cleanly resolved files. |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `clear` | Drop in-progress state when abandoning a merge. `git am`/`git rebase` `--skip`/`--abort` do it for you. |
| `forget <pathspec>` | Delete the recorded fix for matching current conflicts. |
| `diff` | Show your resolution so far. Extra args go to system `diff`. |
| `status` | List conflicted paths rerere will record. |
| `remaining` | List conflicted paths rerere didn't resolve (incl. untrackable, like submodules). |
| `gc` | Prune records: unresolved > 15 days (`gc.rerereUnresolved`), resolved > 60 days (`gc.rerereResolved`). |

Subcommands take no flags.
