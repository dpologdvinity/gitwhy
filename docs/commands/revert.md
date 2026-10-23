# git revert

Undo commits by adding new commits that reverse them. Keeps history, so safe on shared branches. Needs a clean working tree (unless `-n`).

To discard uncommitted edits use `git reset --hard` or `git restore`; for old file versions, `git restore --source`.

**Usage:** `git revert [--[no-]edit] [-n] [-m <parent-number>] [-s] [-S[<keyid>]] <commit>...`
or `git revert (--continue | --skip | --abort | --quit)`

## Flags

| Flag | What it does |
|------|--------------|
| `<commit>...` | Commits to undo. Ranges like `A..B` work. |
| `-e`, `--[no-]edit` | Edit each message. Default in a terminal. |
| `-n`, `--no-commit` | Apply reversals to files and index without committing; index needn't match `HEAD`. Combines several reverts into one commit. |
| `--commit` | Commit. Cancels `-n`. |
| `-m`, `--[no-]mainline <parent-number>` | Required for merges: parent (from 1) to revert to. Re-merging that branch later won't bring those changes back. |
| `--[no-]cleanup=<mode>` | Message cleanup: `strip`, `whitespace`, `verbatim`, `scissors`, `default` (as `git commit`). `scissors` adds a scissors line to `MERGE_MSG` on conflict. |
| `-s`, `--[no-]signoff` | Add `Signed-off-by`. |
| `-S[<keyid>]`, `--[no-]gpg-sign[=<keyid>]` | GPG-sign. Key defaults to committer; no space after `-S`. `--no-gpg-sign` overrides `commit.gpgSign`. |
| `--[no-]strategy=<strategy>` | Merge strategy (see `git merge`). |
| `-X <option>`, `--[no-]strategy-option=<option>` | Strategy option. |
| `--[no-]rerere-autoupdate` | Stage rerere-replayed conflict fixes. |
| `--[no-]reference` | Body says `This reverts commit <abbrev> (<subject>, <date>).` instead of full ID. Default from `revert.reference`. |
| `--continue` | Resume after fixing conflicts. |
| `--skip` | Skip current commit. |
| `--abort` | Cancel and restore the start state. |
| `--quit` | Forget the operation; keep current state. |
