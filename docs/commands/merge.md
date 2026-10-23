# git merge

Join other branches' changes into the current branch. If you have no new commits, the branch just moves forward ("fast-forward"); otherwise a merge commit is made. Old tip saved as `ORIG_HEAD`.

**Usage:** `git merge [<flags>] [<commit>...]`, `git merge (--continue | --abort | --quit)`

- `<commit>...`: usually branch names. Two or more: one "octopus" merge. None: the upstream. `FETCH_HEAD`: what the last `git fetch` marked.
- Commit or stash first. Refuses with staged changes or uncommitted changes in files it must update; `--abort` can't always restore them.
- Merging an annotated tag not stored in `refs/tags/` (e.g. fetched by URL) always makes a merge commit (avoid with `--ff-only` or `<tag>^0`).

**On conflict:** branch doesn't move; `MERGE_HEAD` points at the other branch; clean files are staged; conflicted files get `<<<<<<<`/`=======`/`>>>>>>>` markers (`merge.conflictStyle=diff3` or `zdiff3` also shows the base after `|||||||`); the index holds stages 1 (base), 2 (yours), 3 (theirs), e.g. `git show :2:<file>`. With `ort`, `git diff AUTO_MERGE` shows your fixes so far. Fix, `git add`, then `git merge --continue` (or `git commit`), or `git merge --abort`. Helpers: `git mergetool`, `git diff`, `git log --merge -p <path>`.

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]commit` | Commit the result (default). `--no-commit`: stop before committing. Fast-forwards still happen; add `--no-ff`. |
| `-e`, `--[no-]edit` | Edit the merge message. `--no-edit`: use the generated one. `GIT_MERGE_AUTOEDIT=no` for old scripts. |
| `--[no-]cleanup=<mode>` | Message cleanup (see [`commit`](commit.md)). `scissors` also adds a scissors line to `MERGE_MSG` on conflict. |
| `--[no-]ff` | `--ff` (default): fast-forward if possible. `--no-ff`: always make a merge commit. Defaults to `--no-ff` for annotated tags outside `refs/tags/`. |
| `--ff-only` | Fast-forward or fail. |
| `-S`, `--[no-]gpg-sign[=<key-id>]` | GPG-sign. Key defaults to committer; attach it (`-SABC`). `--no-gpg-sign` overrides `commit.gpgSign`. |
| `--[no-]log[=<n>]` | Add one-line summaries of up to `<n>` merged commits (default 20). |
| `--[no-]signoff` | Add `Signed-off-by:`. |
| `--[no-]stat` | Show a diffstat (default, `merge.stat`). |
| `-n` | `--no-stat`. |
| `--[no-]summary` | Deprecated synonym for `--stat`. |
| `--[no-]compact-summary` | Compact summary with created/renamed/mode notes. |
| `--[no-]squash` | Stage the combined changes without committing or recording a merge; next `git commit` is an ordinary commit. Not with `--commit`. |
| `--[no-]verify` | Run `pre-merge-commit` and `commit-msg` hooks (default). |
| `-s`, `--[no-]strategy=<strategy>` | Strategy (below). Repeat to try in order. Default `ort`, or `octopus` for 2+ branches. |
| `-X`, `--[no-]strategy-option=<option>` | Strategy option (below). |
| `--[no-]verify-signatures` | Abort unless the merged tip is signed with a trusted key. |
| `-q`, `--[no-]quiet` | Less output. Implies `--no-progress`. |
| `-v`, `--[no-]verbose` | More output. |
| `--[no-]progress` | Force progress on/off. Default: on a terminal. |
| `--[no-]autostash` | Stash local changes (`MERGE_AUTOSTASH`) and reapply after. Can conflict. |
| `--[no-]allow-unrelated-histories` | Allow merging histories with no common ancestor. |
| `-m`, `--[no-]message=<msg>` | Merge message. `--log` summaries appended. |
| `-F`, `--file=<file>` | Message from a file. `--log` summaries appended. |
| `--[no-]into-name=<branch>` | Default message as if merging into `<branch>`. |
| `--[no-]rerere-autoupdate` | Let [`rerere`](rerere.md) stage what it resolves. |
| `--[no-]overwrite-ignore` | Overwrite ignored files in the way (default). `--no-`: abort. |
| `--[no-]abort` | Cancel a conflicted merge, restoring pre-merge state and autostash. Like `git reset --merge`. |
| `--[no-]quit` | Forget the merge, leaving index and files as is. Autostash goes to the stash list. |
| `--[no-]continue` | Finish after resolving (runs `git commit`). |

## Merge strategies (`-s`)

| Strategy | What it does |
|----------|--------------|
| `ort` | Default for two heads. Three-way; multiple bases are merged into a virtual base. Detects renames, not copies. Submodules: fast-forward if one contains the other, else conflict with a suggestion. Takes `-X` options below. |
| `recursive` | Alias for `ort` (since 2.50). |
| `resolve` | Simple three-way merge of two heads. No renames. Tries to detect criss-cross ambiguities. |
| `octopus` | Default for 3+ heads. Refuses merges needing manual fixes. |
| `ours` | Keep the current tree exactly; record others as merged. Not `-X ours`. |
| `subtree` | `ort` that first shifts a tree that is a subdirectory of the other to line up. |

Three-way strategies compare only tips and base: a change made on both sides and reverted on one stays in.

## Strategy options (`-X`)

For `ort` (and `recursive`, `subtree`).

| Option | What it does |
|--------|--------------|
| `ours` | Conflicting hunks: take yours. Their clean changes still apply. Binary: yours. Not the `ours` strategy. |
| `theirs` | Conflicting hunks: take theirs. |
| `ignore-space-change` | Treat whitespace-amount-only changes as unchanged. Their whitespace-only change: yours wins; your whitespace-only change vs. their real one: theirs wins. Same for the three below. |
| `ignore-all-space` | Same, ignoring all whitespace. |
| `ignore-space-at-eol` | Same, ignoring line-end whitespace. |
| `ignore-cr-at-eol` | Same, ignoring line-end CR. |
| `renormalize` | Run all three versions through checkout/check-in filters first. For differing line-ending rules. |
| `no-renormalize` | Disable `renormalize`; overrides `merge.renormalize`. |
| `find-renames[=<n>]` | Detect renames (default), optional similarity threshold. Overrides `merge.renames`. |
| `rename-threshold=<n>` | Old name for `find-renames=<n>`. |
| `no-renames` | No rename detection. Overrides `merge.renames`. |
| `diff-algorithm=<algo>` | `histogram` (ort default), `minimal`, `myers`, `patience`. |
| `histogram` | Old name for `diff-algorithm=histogram`. |
| `patience` | Old name for `diff-algorithm=patience`. |
| `subtree[=<path>]` | Like `subtree` strategy, adding or stripping `<path>` instead of guessing. |
