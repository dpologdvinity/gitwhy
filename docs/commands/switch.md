# git switch

Switch branches (updates files and index), optionally creating one. Uncommitted changes come along; refuses if they'd be lost (see `-m`, `--discard-changes`).

**Usage:** `git switch [<flags>] [--no-guess] <branch>`
or `git switch [<flags>] --detach [<start-point>]`
or `git switch [<flags>] (-c | -C) <new-branch> [<start-point>]`
or `git switch [<flags>] --orphan <new-branch>`

## Flags

| Flag | What it does |
|------|--------------|
| `<branch>` | Target branch. `-` = previous (`@{-1}`); `@{-<N>}` = Nth previous. |
| `<start-point>` | Commit for the new branch or detach (default `HEAD`). `A...B` = their single merge base. |
| `-c`, `--[no-]create <new-branch>` | Create and switch. Not created if the switch fails. |
| `-C`, `--[no-]force-create <new-branch>` | Like `-c`; resets an existing branch to `<start-point>`. |
| `-d`, `--[no-]detach` | Switch to a commit, on no branch (detached HEAD). |
| `--[no-]guess` | If only one remote has `<branch>`, create it tracking `<remote>/<branch>`. `checkout.defaultRemote` breaks ties. On by default (`checkout.guess`). |
| `--[no-]discard-changes` | Switch, throwing away uncommitted changes and untracked files in the way (submodules too with `--recurse-submodules`). |
| `-f`, `--[no-]force` | Same as `--discard-changes`. |
| `-m`, `--[no-]merge` | Carry clashing changes over via stash. On conflict the stash is kept: fix, then `git stash drop`. |
| `--[no-]conflict=<style>` | `-m` with style `merge` (default), `diff3`, `zdiff3`. Overrides `merge.conflictStyle`. |
| `-q`, `--[no-]quiet` | No feedback. |
| `--[no-]progress` | Force progress, even off a terminal or with `--quiet`. |
| `-t`, `--[no-]track[=(direct\|inherit)]` | Set the new branch's upstream (see `git branch --track`). Implies `-c`, naming it after the remote branch (`origin/hack` gives `hack`). `--no-track` overrides `branch.autoSetupMerge`. |
| `--[no-]orphan <new-branch>` | New branch with no history. Removes all tracked files. |
| `--[no-]ignore-other-worktrees` | Allow a branch checked out in another worktree. |
| `--[no-]overwrite-ignore` | Overwrite ignored files in the way. On by default; `--no-` stops instead. |
| `--[no-]recurse-submodules[=<checkout>]` | Also update active submodules (detached `HEAD`). Off by default. |
