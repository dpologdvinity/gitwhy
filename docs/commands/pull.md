# git pull

`git fetch`, then combine the fetched branch into the current one. No arguments: uses the branch's upstream (e.g. `origin/main`).

How it combines (pick a flag, or set `pull.rebase`, `pull.ff`, `pull.squash`):

1. `--ff-only` (default): move your branch forward; fail if you have commits the remote lacks.
2. `--rebase`: replay your commits on top (`git rebase`).
3. `--no-rebase`: merge commit if needed (`git merge`).
4. `--squash`: stage the remote's changes as one uncommitted change.

Back out of a conflict with `git merge --abort` or `git rebase --abort`.

**Usage:** `git pull [<flags>] [<repository> [<refspec>...]]`

## Flags

| Flag | What it does |
|------|--------------|
| `<repository>` | Remote name or URL (see [URLs](fetch.md#urls)). Default: `branch.<name>.remote`, else `origin`. |
| `<refspec>...` | What to fetch and combine, e.g. `main`. Default: upstream. All given are merged (several = octopus). See [Refspecs](#refspecs). |
| `-q`, `--[no-]quiet` | Less fetch and merge output. |
| `-v`, `--[no-]verbose` | More output. |
| `--[no-]progress` | Force progress output. |
| `--[no-]recurse-submodules[=(yes\|on-demand\|no)]` | Also fetch submodules and update active ones (rebasing or merging their local commits). |

### Combining (merge and rebase)

| Flag | What it does |
|------|--------------|
| `-r`, `--rebase[=(true\|merges\|false\|interactive)]` | Rebase instead of merge. `true` (bare default): replay your commits on upstream, skipping ones from a rewritten upstream. `merges`: keep local merges (`--rebase-merges`). `false`: merge. `interactive`: edit the todo list. Defaults: `pull.rebase`, `branch.<name>.rebase`, `branch.autoSetupRebase`. Rewrites commits; avoid on pushed ones. |
| `--no-rebase` | Merge (`--rebase=false`). |
| `--ff-only` | Fast-forward only; fail if diverged. Default. |
| `--[no-]ff` | `--ff`: fast-forward if possible, else merge commit (default when merging). `--no-ff`: always a merge commit. |
| `--[no-]commit` | Commit the merge (default). `--no-commit`: stop before committing; add `--no-ff` to block fast-forwards. |
| `--[no-]edit` | Edit the merge message. `--no-edit`: accept it. Or set `GIT_MERGE_AUTOEDIT=no`. No `-e` short form. |
| `--[no-]cleanup=<mode>` | Merge message cleanup (modes in [`git commit`](commit.md)). |
| `-S[<key-id>]`, `--[no-]gpg-sign[=<key-id>]` | GPG-sign the merge. Key defaults to committer; attach it (`-SABC`). `--no-` overrides `commit.gpgSign`. |
| `--[no-]log[=<n>]` | Add up to `<n>` merged commit subjects to the message. |
| `--[no-]signoff` | Add a `Signed-off-by:` trailer. |
| `--[no-]stat` | Diffstat at the end. Default: `merge.stat`. |
| `-n` | `--no-stat`. |
| `--[no-]summary` | Deprecated: `--stat`. |
| `--[no-]compact-summary` | Compact summary at the end (new, renamed, mode-changed files). |
| `--[no-]squash` | Stage the merged changes without committing or recording a merge. Next `git commit` makes a plain commit. Not with `--commit`. |
| `--[no-]verify` | Run `pre-merge-commit` and `commit-msg` hooks (default). |
| `-s <strategy>`, `--[no-]strategy=<strategy>` | Merge strategy; repeat to try in order. Default: `ort` (one branch), `octopus` (more). See [Strategies](#strategies). |
| `-X <option>`, `--[no-]strategy-option=<option>` | Strategy option, e.g. `-X theirs`. See [Strategies](#strategies). |
| `--[no-]verify-signatures` | Abort unless the merged tip has a valid trusted GPG signature. |
| `--[no-]autostash` | Stash local changes first, re-apply after. Re-apply can conflict. |
| `--[no-]allow-unrelated-histories` | Allow merging histories with no common ancestor. |

### Fetching

| Flag | What it does |
|------|--------------|
| `--[no-]all` | Fetch all remotes (except `skipFetchAll`). Overrides `fetch.all`. |
| `-a`, `--[no-]append` | Append to `.git/FETCH_HEAD`, don't overwrite. |
| `--[no-]depth=<depth>` | Shallow: `<depth>` commits per branch tip; changes an existing shallow depth. |
| `--[no-]deepen=<n>` | Shallow: deepen by `<n>` commits. |
| `--[no-]shallow-since=<date>` | Shallow: commits after `<date>`. |
| `--[no-]shallow-exclude=<ref>` | Shallow: omit commits reachable from this remote ref. Repeatable. |
| `--unshallow` | Make a shallow clone full. |
| `--[no-]update-shallow` | Accept refs that need `.git/shallow` updated (refused by default). |
| `--[no-]negotiation-restrict=(<commit>\|<glob>)` | Only advertise commits reachable from these as "have". Speeds fetches. Repeatable. |
| `--[no-]negotiation-tip=(<commit>\|<glob>)` | Old name for `--negotiation-restrict`. |
| `--[no-]negotiation-include=(<commit>\|<glob>)` | Always advertise these commits. Repeatable. Default: `remote.<name>.negotiationInclude`. |
| `-f`, `--[no-]force` | Let `<src>:<dst>` refspecs overwrite non-fast-forward. |
| `-k`, `--[no-]keep` | Keep the downloaded pack. |
| `-p`, `--[no-]prune` | Delete remote-tracking refs gone from the remote. Tags only if fetched by explicit refspec. |
| `--no-tags` | Don't auto-fetch tags. Default: `remote.<name>.tagOpt`. |
| `-t`, `--tags` | Fetch all remote tags. |
| `--refmap=<refspec>` | Map command-line refs with this refspec instead of `remote.<name>.fetch`. Repeatable; empty disables. |
| `-j`, `--[no-]jobs[=<n>]` | Fetch `<n>` remotes/submodules at once. `0`: auto. Config: `fetch.parallel`, `submodule.fetchJobs`. |
| `--[no-]set-upstream` | On success, set the fetched branch as upstream. |
| `--[no-]upload-pack=<upload-pack>` | Path to `git-upload-pack` on the remote. |
| `-o <option>`, `--[no-]server-option=<option>` | Send `<option>` to the server (protocol v2). Repeatable. Default: `remote.<name>.serverOption`. |
| `--[no-]show-forced-updates` | Detect force-pushed branches (default). Off for speed via `fetch.showForcedUpdates=false`. |
| `-4`, `--[no-]ipv4` | IPv4 only. |
| `-6`, `--[no-]ipv6` | IPv6 only. |
| `--[no-]dry-run` | Show what would be fetched; change nothing. |

In `git pull`'s manual but rejected by git 2.55's `pull`; use with [`git fetch`](fetch.md):

| Flag | What it does (in `git fetch`) |
|------|--------------|
| `--atomic` | Update all local refs or none. |
| `--negotiate-only` | Fetch nothing; print shared `--negotiation-restrict` ancestors. For `push.negotiate`. |
| `--porcelain` | Script-friendly stdout. |
| `--filter=<filter-spec>` | Partial fetch (e.g. `blob:none`; `auto` = server's suggestion). |
| `--prefetch` | Put refs under `refs/prefetch/` (for `git maintenance`). |

## Refspecs

`[+]<src>[:<dst>]`: fetch remote `<src>`; update local `<dst>` if given. Examples: `main`, `refs/heads/*:refs/remotes/origin/*`, `^refs/heads/wip*` (exclude), `tag v1.0` (= `refs/tags/v1.0:refs/tags/v1.0`). `<src>` may be a full commit ID.

- Without `+` or `--force`: `refs/heads/*` updates must fast-forward; `refs/tags/*` can't change at all. Other refs update freely.
- Use `+` for remote branches that get rewritten.
- Command-line refspecs are all merged. With none, all configured ones are fetched but only one merged: `branch.<name>.merge`, else nothing for a glob, else the first.

## Upstream

Stored as `branch.<name>.remote` and `branch.<name>.merge`. Set by `git clone` (default branch), `git checkout <branch>` (from `origin/<branch>`), `git push -u`, or `push.autoSetupRemote`.

## Strategies

| Strategy (`-s`) | What it does |
|-----------------|--------------|
| `ort` | Default for one branch. Three-way merge with renames; merges multiple bases first. |
| `recursive` | Same as `ort` since 2.50. |
| `resolve` | Simple three-way merge; no renames. |
| `octopus` | Default for several branches. Refuses manual-fix merges. |
| `ours` | Keep your tree exactly; ignore the other side. Not `-X ours`. |
| `subtree` | Like `ort`, matching the other tree to your subfolder first. |

| Option (`-X`, for `ort`) | What it does |
|--------------------------|--------------|
| `ours` / `theirs` | Conflicting hunks take our / their side. Other changes still merge. Binary: whole file. |
| `ignore-space-change`, `ignore-all-space`, `ignore-space-at-eol`, `ignore-cr-at-eol` | Treat that whitespace as unchanged. Whitespace-only changes lose to the other side. |
| `renormalize` / `no-renormalize` | Normalize line endings and filters before comparing. `no-` overrides `merge.renormalize`. |
| `find-renames[=<n>]` | Detect renames (default), with similarity threshold `<n>`. |
| `rename-threshold=<n>` | Deprecated: `find-renames=<n>`. |
| `no-renames` | No rename detection. |
| `diff-algorithm=(histogram\|minimal\|myers\|patience)` | Merge diff algorithm (default `histogram`). |
| `subtree[=<path>]` | Like `subtree` strategy with an explicit folder prefix. |

Three-way merges see only the two tips and base: a change made on both sides but reverted on one still lands.
