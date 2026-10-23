# git filter-branch

Not recommended: use [`git filter-repo`](filter-repo.md). Very slow and easy to get subtly wrong; warns every run (silence with `FILTER_BRANCH_SQUELCH_WARNING=1`).

Rewrite history by running shell "filters" on every commit of the given branches (files, messages, authors, parents). Everything else is kept. All rewritten commits get new IDs, so collaborators must re-sync. Makes `git replace` replacements and grafts permanent.

Only named positive refs are rewritten (`a..b` rewrites `b`). Old refs are saved under `refs/original/`; check results before deleting them. Exit 2 if nothing to rewrite.

Filters run in the order below via shell `eval` (except `--commit-filter`). `GIT_COMMIT` is the old ID; `GIT_AUTHOR_NAME`/`_EMAIL`/`_DATE` and `GIT_COMMITTER_*` set the new identity. A failing filter aborts everything. `map <old-id>` prints the rewritten ID (or the old one if not rewritten yet).

**Usage:** `git filter-branch [<filters>] [--prune-empty] [--original <namespace>] [-d <dir>] [-f] [--state-branch <branch>] [--] [<rev-list-options>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<rev-list-options>...` | What to rewrite, as [`git rev-list`](rev-list.md) takes (see [commit selection options](../shared/revision-options.md)), e.g. `-- --all`. Put `--` before flags like `--all`. Tips excluded by path limits move to the nearest kept ancestor. |
| `--setup <command>` | Run once first; its definitions are available to all filters except `--commit-filter`. |
| `--subdirectory-filter <directory>` | Keep only history of `<directory>` and make it the root. |
| `--env-filter <command>` | Change only the environment (author/committer variables). |
| `--tree-filter <command>` | Run in each checked-out commit; resulting files become the commit (ignore rules don't apply). Slow. Use `rm -f`. |
| `--index-filter <command>` | Edit the index without checkout. Much faster. Typical: `git rm --cached --ignore-unmatch <file>`. |
| `--parent-filter <command>` | Parent list on stdin (`""`, `-p <id>`, `-p <id> -p <id>...`); print the new one. |
| `--msg-filter <command>` | Old message on stdin, new on stdout. |
| `--commit-filter <command>` | Replaces `git commit-tree` (args `<tree> [-p <parent>]...`, message on stdin); print new ID(s), all used as parents of children. `skip_commit "$@"` drops the commit (changes carry into the next); `git_commit_non_empty_tree "$@"` skips no-op commits. |
| `--tag-name-filter <command>` | Rename tags on rewritten commits: old name on stdin, new on stdout. `cat` keeps names (overwrites old tags; back up). Signatures removed; tags of tags point at the commit. |
| `--prune-empty` | Drop non-merge commits that end up empty. Not with `--commit-filter` (use `git_commit_non_empty_tree`). |
| `--original <namespace>` | Where old refs go. Default `refs/original`. |
| `-d <directory>` | Temp folder. Default `.git-rewrite/`. tmpfs speeds it up. |
| `-f`, `--force` | Run even if the temp folder or `refs/original/` exists. |
| `--state-branch <branch>` | Load and save the old-to-new ID map here for incremental runs. Created if missing. |

## Examples

- Remove a file from all history: `git filter-branch --index-filter 'git rm --cached --ignore-unmatch secret.txt' -- --all`
- Make `foodir/` the root: `git filter-branch --subdirectory-filter foodir -- --all`
- Graft history: `git replace --graft <commit> <new-parent>` then `git filter-branch <new-parent>..HEAD`

## Making the repo actually smaller

Removed files stay on disk until dropped. First remove every name the file had (`git log --name-only --follow --all -- <file>`) and rewrite all refs (`--tag-name-filter cat -- --all`). Then either:

- Safe: `git clone file:///path/to/repo` (a plain path hardlinks, keeping everything).
- In place (destructive; back up first): delete `refs/original/*` (`git for-each-ref --format="%(refname)" refs/original/ | xargs -n 1 git update-ref -d`), then `git reflog expire --expire=now --all`, then `git gc --prune=now`.
