# git cherry-pick

Copy existing commits onto the current branch as new commits, e.g. to bring one fix over without merging. Needs a clean working tree (unless `-n`).

On conflict it stops with `CHERRY_PICK_HEAD` set and conflict markers in files. Fix, `git add`, then `git cherry-pick --continue`.

To undo commits, use `git revert` instead.

**Usage:** `git cherry-pick [<flags>] <commit>...`
or `git cherry-pick (--continue | --skip | --abort | --quit)`

## Flags

Also takes [commit selection options](../shared/revision-options.md) (e.g. `--stdin`, `--reverse`). Each `<commit>` is picked alone; a range `A..B` picks every commit in it, oldest first.

| Flag | What it does |
|------|--------------|
| `<commit>...` | Commits to copy, e.g. `abc123`, `main~4 main~2`, `..main` (in `main`, not `HEAD`). |
| `-e`, `--[no-]edit` | Edit each message first. |
| `-n`, `--no-commit` | Apply to files and index without committing. Good for squashing picks. No `CHERRY_PICK_HEAD`. |
| `--commit` | Commit as usual. Cancels `-n`. |
| `-x` | Add `(cherry picked from commit <id>)` to the message. For backports between public branches. |
| `-r` | Does nothing (legacy). |
| `-m`, `--[no-]mainline <parent-number>` | Required to pick a merge: which parent (from 1) is the main line. |
| `--[no-]cleanup=<mode>` | Message cleanup, as in `git commit`: `strip`, `whitespace`, `verbatim`, `scissors`, `default`. |
| `-s`, `--[no-]signoff` | Add a `Signed-off-by` trailer. |
| `-S[<keyid>]`, `--[no-]gpg-sign[=<keyid>]` | GPG-sign. Key defaults to committer; no space after `-S`. `--no-` overrides `commit.gpgSign`. |
| `--[no-]ff` | If `HEAD` is the pick's parent, fast-forward to it instead of copying. |
| `--[no-]allow-empty` | Keep originally empty commits. For ones that become empty, use `--empty=keep`. |
| `--[no-]allow-empty-message` | Allow commits with empty messages. |
| `--empty=<mode>` | Commit whose changes are already present: `stop` (default), `drop`, or `keep` (implies `--allow-empty`). |
| `--[no-]keep-redundant-commits` | Deprecated: `--empty=keep`. |
| `--[no-]strategy=<strategy>` | Merge strategy (see `git merge`). |
| `-X <option>`, `--[no-]strategy-option=<option>` | Strategy option, e.g. `-Xpatience`, `-Xtheirs`. |
| `--[no-]rerere-autoupdate` | Stage conflict fixes rerere reuses. `--no-`: leave them to check. |
| `--continue` | Resume after fixing conflicts. |
| `--skip` | Skip current commit; go on. |
| `--abort` | Cancel; return to the start (keeps earlier local edits). |
| `--quit` | Forget the operation; keep current state. |
