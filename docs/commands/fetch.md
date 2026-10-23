# git fetch

Download commits, branches, and tags from another repo without touching your branches or files. Updates remote-tracking branches like `origin/main` so you can inspect before `git merge` or `git rebase`. `git pull` = fetch plus merge/rebase.

By default also gets tags pointing into fetched history. Writes results to `.git/FETCH_HEAD`.

Output: one line per ref, `<flag> <summary> <from> -> <to> [<reason>]`. Flags: ` ` fast-forward, `+` forced update, `-` pruned, `t` tag updated, `*` new ref, `!` rejected/failed, `=` up to date (only with `-v`).

**Usage:** `git fetch [<flags>] [<repository> [<refspec>...]]`
or `git fetch [<flags>] <group>`
or `git fetch --multiple [<flags>] [(<repository> | <group>)...]`
or `git fetch --all [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `<repository>` | Remote name (`origin`) or URL. Default: current branch's remote, else `origin`. |
| `<group>` | Remotes listed in config `remotes.<group>`. |
| `<refspec>...` | What to fetch and where to store it: `[+]<src>[:<dst>]`, e.g. `main`, `refs/heads/*:refs/remotes/origin/*` (one `*` each side). `+` allows non-fast-forward. `^<src>` excludes. `tag <name>` = `refs/tags/<name>:refs/tags/<name>`. Default: `remote.<name>.fetch`. Non-fast-forward to `refs/heads/*`, or any change to `refs/tags/*`, needs `+` or `--force`; other namespaces update freely. |
| `--[no-]all` | Fetch every remote (except `remote.<name>.skipFetchAll`). Overrides `fetch.all`. |
| `-m`, `--[no-]multiple` | Allow several remotes/groups as arguments. No refspecs. |
| `-a`, `--[no-]append` | Append to `.git/FETCH_HEAD`, don't overwrite. |
| `--[no-]atomic` | Update all local refs or none. |
| `--[no-]depth=<depth>` | Shallow: only `<depth>` commits per tip. In a shallow repo, deepen or cut to that. Doesn't fetch tags for deepened commits. |
| `--[no-]deepen=<n>` | `--depth`, counted from the current shallow boundary. |
| `--[no-]shallow-since=<date>` | Shallow: keep history after `<date>`. |
| `--[no-]shallow-exclude=<ref>` | Shallow: omit commits reachable from this remote branch or tag. Repeatable. |
| `--unshallow` | Make a shallow repo full (as much as the source has). |
| `--[no-]update-shallow` | Accept refs from a shallow source that need `.git/shallow` updated (refused by default). |
| `--[no-]negotiation-restrict=<commit-or-glob>`, `--[no-]negotiation-tip=<commit-or-glob>` | Negotiate using only commits reachable from these, not all local refs. Faster. Repeatable; globs match ref names. `--negotiation-tip` = old name. |
| `--[no-]negotiation-include=<commit-or-glob>` | Always tell the server you have these commits. Applied after `--negotiation-restrict`. Default `remote.<name>.negotiationInclude`. |
| `--[no-]negotiate-only` | Fetch nothing; print `--negotiation-restrict` ancestors the server also has. Used by `push.negotiate`. Not with `--recurse-submodules=yes\|on-demand`. |
| `--[no-]dry-run` | Show what would happen; change nothing. |
| `--[no-]porcelain` | Script-friendly output on stdout: `<flag> <old-id> <new-id> <local-ref>`. Not with `--recurse-submodules=yes\|on-demand`. Overrides `fetch.output`. |
| `--[no-]filter=<filter-spec>` | Partial clone: server omits some objects, fetched later on demand. E.g. `blob:none`, `blob:limit=1m`. `auto`: combine filters advertised for accepted promisor remotes. See `git rev-list --filter`. |
| `--[no-]write-fetch-head` | Write `.git/FETCH_HEAD` (default). Never written with `--dry-run`. |
| `-f`, `--[no-]force` | Let `<src>:<dst>` refspecs force non-fast-forward updates of local refs. |
| `-k`, `--[no-]keep` | Keep the downloaded pack. |
| `--[no-]auto-maintenance`, `--[no-]auto-gc` | Run `git maintenance run --auto` at the end (default). |
| `--[no-]write-commit-graph` | Write the commit-graph file after fetching. Overrides `fetch.writeCommitGraph`. |
| `--[no-]prefetch` | Store fetched refs under `refs/prefetch/` (for `git maintenance` prefetch). |
| `-p`, `--[no-]prune` | First delete remote-tracking refs whose remote branch is gone. Tags pruned only if an explicit refspec fetches them. Default `fetch.prune` / `remote.<name>.prune`. |
| `-P`, `--[no-]prune-tags` | With `--prune`, also delete local tags not on the remote (refspec `refs/tags/*:refs/tags/*`; changed tags still need `--force`). Deletes your local-only tags. Default `fetch.pruneTags`. |
| `-t`, `--[no-]tags` | Also fetch all tags (`refs/tags/*`). Doesn't make tags prunable alone. |
| `-n`, `--no-tags` | Don't auto-fetch tags. Per-remote default: `remote.<name>.tagOpt`. |
| `--refetch` | Re-download everything like a fresh clone, no negotiation. Use after changing a partial-clone filter. Maintenance removes duplicates. |
| `--refmap=<refspec>` | Map command-line refs to remote-tracking refs with this refspec, not `remote.<name>.fetch`. Repeatable. Empty: don't update remote-tracking refs. |
| `--[no-]recurse-submodules[=<when>]` | Also fetch submodules: `on-demand` (default: only those with newly referenced commits), `yes` (all populated), `no`. No value = `yes`. Default `fetch.recurseSubmodules`. Only cloned submodules. |
| `-j`, `--[no-]jobs=<n>` | Fetch up to `<n>` remotes (with `--multiple`) or submodules in parallel. `0` = sensible default. Default 1. Config: `fetch.parallel`, `submodule.fetchJobs`. |
| `--[no-]set-upstream` | On success, set fetched remote branch as current branch's upstream. |
| `--submodule-prefix=<path>` | Internal: path prefix in messages while recursing submodules. |
| `--recurse-submodules-default=<yes\|on-demand>` | Internal: fallback for `--recurse-submodules`; any other setting wins. |
| `-u`, `--[no-]update-head-ok` | Internal (for `git pull`): allow updating the current branch's ref. |
| `--[no-]upload-pack=<path>` | Path to `git-upload-pack` on the remote (passed to `git fetch-pack` as `--exec`). |
| `-q`, `--[no-]quiet` | Less output; no progress. |
| `-v`, `--[no-]verbose` | More output, including up-to-date refs. |
| `--[no-]progress` | Show progress even when not on a terminal. |
| `-o`, `--[no-]server-option=<option>` | Send string to server (protocol v2; no NUL or newline). Repeatable. Default `remote.<name>.serverOption`. |
| `--[no-]show-forced-updates` | Check for force-updated branches (default). `--no-` skips it for speed. Overrides `fetch.showForcedUpdates`. |
| `-4`, `--ipv4` | Use only IPv4. |
| `-6`, `--ipv6` | Use only IPv6. |
| `--[no-]stdin` | Also read refspecs from stdin, one per line (not `tag <name>`). |

## Pruning

`git fetch --prune origin` (or `git remote prune origin`, no fetch) deletes remote-tracking refs for branches deleted on the remote. Always: `fetch.prune=true`. Pruning follows the remote's refspecs: with `refs/tags/*:refs/tags/*` (or `--prune-tags`), local tags missing from the remote are deleted too, including local-only or other remotes' tags.

## URLs

`<repository>` can be:

- `ssh://[user@]host[:port]/path` or `[user@]host:path` (SSH)
- `https://host[:port]/path` (also `http`; `ftp[s]` slow, deprecated)
- `git://host[:port]/path` (no auth, no encryption)
- `/path/to/repo` or `file:///path/to/repo` (local)
- `<transport>::<address>` (a remote helper, e.g. `ext::`)

`url.<base>.insteadOf` config rewrites URL prefixes.
