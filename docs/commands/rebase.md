# git rebase

Replay your commits, as new commits, on top of another starting point, giving straight-line history instead of a merge. `-i` lets you reorder, edit, combine, or drop them.

**Usage:** `git rebase [<flags>] [--onto <newbase> | --keep-base] [<upstream> [<branch>]]`, `git rebase [<flags>] [--onto <newbase>] --root [<branch>]`, `git rebase (--continue | --skip | --abort | --quit | --edit-todo | --show-current-patch)`

- `<upstream>`: your commits not in it get replayed. Default: the branch's upstream (with `--fork-point`). Any commit works.
- `<branch>`: switch to it first. Default: current branch.
- Commits whose changes are already in `<upstream>` are skipped. `HEAD` detaches at `<upstream>` (or `--onto`), each commit is cherry-picked, then the branch points at the result.
- Old tip saved in `ORIG_HEAD` and `<branch>@{1}`.
- **On conflict:** fix, `git add`, `git rebase --continue`. Or `--skip` the commit, or `--abort`. "Ours" = new base plus replayed commits; "theirs" = your commit being replayed (reversed from merge).
- Don't rebase commits others built on. If upstream rebased under you: `git rebase <upstream>` if patches are unchanged, else `git rebase --onto <upstream> <old-upstream-tip>` (e.g. `<upstream>@{1}`).
- `--onto` examples: `git rebase --onto main next topic` moves `topic` from `next` to `main`. `git rebase --onto X~5 X~3 X` deletes `X~4` and `X~3`.
- The `pre-rebase` hook can refuse.

## Mode flags

Each used alone.

| Flag | What it does |
|------|--------------|
| `--continue` | Resume after fixing a conflict. |
| `--skip` | Drop the current commit and resume. |
| `--abort` | Restore the branch and `HEAD` to the start. |
| `--quit` | Stop, leaving `HEAD`, index, and files as they are. Autostash goes to the stash list. |
| `--edit-todo` | Edit the remaining interactive todo list. |
| `--show-current-patch` | Show the commit being applied (`git show REBASE_HEAD`). |

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]onto=<newbase>` | Replay onto this instead of `<upstream>`. `A...B` = their merge base (a missing side = `HEAD`). |
| `--[no-]keep-base` | Replay onto the merge base of `<upstream>` and `<branch>`, so the base doesn't move. Implies `--reapply-cherry-picks --no-fork-point`. Not with `--onto` or `--root`. |
| `--apply` | Old apply backend (`git am`). Drops empty commits, no directory-rename detection, can misplace hunks with repeated context. |
| `-m`, `--merge` | Merge backend (default). |
| `--empty=(drop\|keep\|stop)` | Commits made empty by upstream: `drop` (default), `keep` (default with `--exec` without `-i`), `stop` (default with `-i`; old name `ask`). |
| `--[no-]keep-empty` | Keep originally empty commits (default). |
| `--[no-]reapply-cherry-picks` | Replay commits even if upstream has the same change. Default: drop them with a warning (reads all upstream commits; this flag is faster). On with `--keep-base`. |
| `--allow-empty-message` | No-op. |
| `-s`, `--[no-]strategy=<strategy>` | Merge strategy (default `ort`). Implies `--merge`. `ours` discards all your changes. See [strategies](merge.md#merge-strategies--s). |
| `-X`, `--[no-]strategy-option=<option>` | Strategy option. Implies `--merge`. `ours`/`theirs` are swapped. See [strategy options](merge.md#strategy-options--x). |
| `--[no-]rerere-autoupdate` | Let [`rerere`](rerere.md) stage what it resolves. |
| `-S`, `--[no-]gpg-sign[=<key-id>]` | GPG-sign new commits. Key defaults to committer; attach it (`-SABC`). `--no-gpg-sign` overrides `commit.gpgSign`. |
| `-q`, `--[no-]quiet` | Less output. Implies `--no-stat`. |
| `-v`, `--[no-]verbose` | More output. Implies `--stat`. |
| `--stat` | Diffstat of upstream changes since last rebase. See `rebase.stat`. |
| `-n`, `--no-stat` | No diffstat. |
| `--[no-]verify` | Run `pre-rebase` (default). |
| `-C <n>` | Require `<n>` matching context lines. Implies `--apply`. |
| `-f`, `--[no-]force-rebase`, `--no-ff` | Recreate every commit, even unchanged ones. Helps re-merge after reverting a merge. `--ff` is the opposite. |
| `--[no-]fork-point` | Use `<upstream>`'s reflog to find the real start, so commits upstream dropped aren't replayed. Default on only when `<upstream>` is omitted. See `rebase.forkPoint`. |
| `--[no-]ignore-whitespace` | Ignore whitespace-only differences. Apply backend: in context. Merge backend: whitespace-only hunks may be lost. |
| `--[no-]whitespace=<action>` | Passed to [`git apply`](apply.md). Implies `--apply`. |
| `--[no-]committer-date-is-author-date` | Committer date = author date. Implies `--force-rebase`. Can make dates go backwards. |
| `--[no-]reset-author-date`, `--ignore-date` | Author date = now. Implies `--force-rebase`. |
| `--[no-]signoff` | Add `Signed-off-by:` to rebased commits (with `-i`: picked, edited, reworded only). |
| `--[no-]trailer=<trailer>` | Add a trailer (e.g. `Reviewed-by: X`) to each rebased commit. Repeatable. Implies `--force-rebase`. |
| `-i`, `--interactive` | Edit the todo list first (below). Line format: `rebase.instructionFormat`. |
| `-r`, `--[no-]rebase-merges[=<mode>]` | Recreate merge commits instead of flattening (`label`/`reset`/`merge` lines). Merge conflict fixes must be redone. `no-rebase-cousins` (default): commits not descended from `<upstream>` keep their base; `rebase-cousins` moves them too. Merges use `ort`. `--no-rebase-merges` overrides `rebase.rebaseMerges`. |
| `-x`, `--[no-]exec=<cmd>` | Run `<cmd>` after each commit (e.g. `-x "make test"`); failure stops the rebase. Repeatable. With `--autosquash`, after each squash group. No `-i` needed. |
| `--[no-]root` | Replay all commits from the first, so roots can be edited. |
| `--[no-]autosquash` | Move `fixup! <title>`, `squash! <title>`, `amend! <title>` commits after their target, marked `fixup`, `squash`, `fixup -C`. Create with `git commit --fixup`/`--squash`. Config: `rebase.autoSquash`. |
| `--[no-]autostash` | Stash local changes and reapply after. Can conflict. |
| `--[no-]reschedule-failed-exec` | Re-queue failed `exec` lines to rerun on `--continue`. Fixed at start. |
| `--[no-]update-refs` | Also move other branches pointing at rebased commits (not ones checked out in a worktree). For stacked branches. Config: `rebase.updateRefs`. |

`--apply`, `--whitespace`, `-C` can't be combined with: `--merge`, `--strategy`, `--strategy-option`, `--autosquash`, `--rebase-merges`, `--interactive`, `--exec`, `--no-keep-empty`, `--empty`, `--[no-]reapply-cherry-picks` (without `--keep-base`), `--update-refs`, `--root` (without `--onto`), `--trailer`. Also incompatible: `--keep-base` with `--onto` or `--root`; `--fork-point` with `--root`.

## Interactive todo list

`git rebase -i <commit>` lists commits after `<commit>`, oldest first: `pick deadbee Commit title`. Change commands, reorder, or delete lines, then save. Only the command and ID matter. Merges omitted unless `-r`.

| Command | What it does |
|---------|--------------|
| `p`, `pick <commit>` | Use as-is. |
| `r`, `reword <commit>` | Use, edit message. |
| `e`, `edit <commit>` | Use, then stop to amend, split, or add commits. `git rebase --continue` after. |
| `s`, `squash <commit>` | Meld into the previous commit; edit combined message. First commit's author kept. |
| `f`, `fixup [-C \| -c] <commit>` | Meld into the previous commit, dropping this message. `-C`: use only this message. `-c`: same, in the editor. |
| `x`, `exec <command>` | Run a shell command at repo root. Non-zero exit stops. |
| `b`, `break` | Stop here. `--continue` resumes. |
| `d`, `drop <commit>` | Remove the commit. |
| `l`, `label <label>` | Name current `HEAD` (`refs/rewritten/<label>`, deleted after). |
| `t`, `reset <label>` | Move `HEAD`, index, files to a label or commit. Won't overwrite untracked files. |
| `m`, `merge [-C <commit> \| -c <commit>] <label> [# <oneline>]` | Merge `<label>` into `HEAD` (more labels = octopus). `-C`: reuse that merge's message; `-c`: and edit it; neither: `<oneline>` as message. |
| `u`, `update-ref <ref>` | Move `<ref>` to current `HEAD` at the end. |
| `noop` | Do nothing. |

`#` lines are comments. `rebase.missingCommitsCheck` (`ignore`/`warn`/`error`) handles deleted lines; `drop` avoids the warning.

**Split a commit:** mark `edit`; at the stop, `git reset HEAD^`, commit in pieces, then `git rebase --continue`.
