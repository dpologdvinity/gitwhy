# git checkout

Switch branches, or restore files from the index or a commit. Clearer split commands: [`git switch`](switch.md) and [`git restore`](restore.md).

- `git checkout <branch>`: switch and update files. Keeps uncommitted changes; fails if they'd be overwritten. If only one remote has `<branch>`, creates a local tracking branch.
- `git checkout <commit>` or `--detach`: detached HEAD (on a commit, not a branch). New commits there are eventually deleted unless you branch (`git checkout -b <name>`) or tag them. Recover with `git reflog`.
- `git checkout [<commit>] [--] <path>...`: overwrite files from the index (discards unstaged edits) or from `<commit>` (also stages).
- No arguments: print tracking info.

`<branch>` can be `-` or `@{-N}` (previous branches). `A...B` means their merge base. Ambiguous names mean the branch; use `--` to force files.

**Usage:** `git checkout [<flags>] <branch>`
or `git checkout [<flags>] [--detach] <commit>`
or `git checkout [<flags>] (-b | -B | --orphan) <new-branch> [<start-point>]`
or `git checkout [<flags>] [<tree-ish>] [--] <pathspec>...`
or `git checkout -p [<tree-ish>] [--] [<pathspec>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `-b <new-branch>` | Create a branch at `<start-point>` (default `HEAD`) and switch to it. |
| `-B <new-branch>` | Same, resetting it if it exists. |
| `--[no-]orphan <new-branch>` | New branch with no history; next commit is a root. Files stay as `<start-point>`'s (`git rm -rf .` to start empty). |
| `-d`, `--[no-]detach` | Check out a commit, not on a branch. Default for non-branch names. |
| `-t`, `--[no-]track[=<direct\|inherit>]` | Set new branch's upstream: `direct` (default) = start point; `inherit` = start point's upstream. Without `-b`, names the branch after the remote one (`origin/hack` makes `hack`). `--no-track` overrides `branch.autoSetupMerge`. |
| `-l` | Create a reflog for the new branch. |
| `--[no-]guess` | Create a tracking branch if one remote has the name. On by default (`checkout.guess`). `checkout.defaultRemote` breaks ties. |
| `-f`, `--[no-]force` | Switching: discard local changes and untracked files in the way. Restoring: skip conflicted files instead of failing. |
| `-m`, `--[no-]merge` | Switching: carry local changes over; on conflict the stash entry is kept. Restoring: recreate conflict markers. |
| `--[no-]conflict=<style>` | `-m` with marker style `merge` (default), `diff3`, or `zdiff3`. Overrides `merge.conflictStyle`. |
| `-2`, `--ours` | Conflicted files: take our side (stage 2). In rebase, the branch rebased onto. |
| `-3`, `--theirs` | Conflicted files: take their side (stage 3). In rebase, your replayed work. |
| `-p`, `--[no-]patch` | Pick hunks to discard (working tree, plus index with a `<tree-ish>`). Always no-overlay. Keys as `git add -p`. |
| `--[no-]auto-advance` | With `-p`: go to the next file after a file's last hunk. |
| `-U`, `--unified=<n>` | With `-p`: context lines (default `diff.context` or 3). |
| `--inter-hunk-context=<n>` | With `-p`: join hunks closer than `<n>` lines (default `diff.interHunkContext` or 0). |
| `--[no-]overlay` | Default overlay never deletes. `--no-overlay`: remove tracked files not in `<tree-ish>`. |
| `--[no-]overwrite-ignore` | Overwrite ignored files when switching. On by default; `--no-` aborts instead. |
| `--[no-]ignore-other-worktrees` | Allow a branch checked out in another worktree. |
| `--[no-]ignore-skip-worktree-bits` | Sparse checkout: also restore paths outside the sparse patterns. |
| `--[no-]recurse-submodules[=<checkout>]` | Also update submodules to recorded commits (detached). Fails on local submodule changes unless `-f`. |
| `-q`, `--[no-]quiet` | No feedback. |
| `--[no-]progress` | Force progress, even off a terminal or with `-q`. |
| `--[no-]pathspec-from-file=<file>` | Read paths from a file (`-` = stdin). |
| `--[no-]pathspec-file-nul` | With `--pathspec-from-file`: NUL-separated, literal paths. |
| `--` | Rest are paths. |
