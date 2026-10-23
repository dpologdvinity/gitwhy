# git svn

Use git with a Subversion project: clone SVN history, pull new revisions (`fetch`, `rebase`), and send git commits back as SVN revisions (`dcommit`). Handles standard `trunk/branches/tags` (`--stdlayout`) or any layout (`-T`/`-b`/`-t`).

Use `git svn rebase`, not `git pull`/`git merge`, on branches you'll `dcommit`: SVN can't see git merges, and merging can send commits to the wrong SVN branch. Don't amend after `dcommit`. Plain `git clone` of a git-svn repo loses SVN metadata; copy with `rsync`.

Often a separate package (`git-svn` on Debian/Ubuntu). Needs Perl SVN bindings.

**Usage:** `git svn <command> [<options>] [<arguments>]`

Typical: `git svn clone http://svn.example.com/project --stdlayout --prefix svn/`, commit in git, `git svn rebase`, then `git svn dcommit`.

## Flags

All subcommands.

| Flag | What it does |
|------|--------------|
| `-h`, `-H`, `--help` | Usage (per subcommand if given). Also `git svn help`. |
| `-V`, `--version` | Print versions. |
| `-i <GIT_SVN_ID>`, `--id <GIT_SVN_ID>` | Ref to track for a single URL (instead of `GIT_SVN_ID` env). |
| `-R <remote-name>`, `--svn-remote <remote-name>`, `--remote <remote-name>` | `[svn-remote "<name>"]` section to use, for multiple SVN repos. Default `svn`. |
| `--minimize-connections` | Used by `migrate`. |

Many flags also work as config `svn.<flag>` (no dashes, e.g. `svn.authorsfile`) or `svn-remote.<name>.<flag>`.

## Shared flag groups

Subcommands below list which groups they take.

### Connection flags

| Flag | What it does |
|------|--------------|
| `--username=<user>` | SVN user for http, https, `svn://`. For `svn+ssh://`, put it in the URL (`svn+ssh://foo@host/...`). |
| `--config-dir=<dir>` | SVN config folder instead of `~/.subversion`. |
| `--no-auth-cache` | Don't save SVN passwords. |
| `--ignore-paths=<regex>` | Skip SVN paths matching this Perl regex (e.g. `"^doc"`). Must match every fetch. Saved as `svn-remote.<name>.ignore-paths` by `init`/`clone`; both apply. Beats `--include-paths`. |
| `--include-paths=<regex>` | Only SVN paths matching regex. Saved as `svn-remote.<name>.include-paths`. |
| `--ignore-refs=<regex>` | Skip branches/tags whose ref matches. Negative lookahead keeps some: `^refs/remotes/origin/(?!tags/wanted-tag\|wanted-branch).*$`. Saved as `svn-remote.<name>.ignore-refs`. |

### Fetch flags

Plus connection flags.

| Flag | What it does |
|------|--------------|
| `--[no-]follow-parent`, `--[no-]follow` | Link each SVN branch to the history it was copied from (default on). Off is faster, branches unconnected. `svn.followparent`. |
| `-A <file>`, `--authors-file=<file>` | Map SVN users to git authors: `loginname = Joe User <user@example.com>` (`<>` for no email). Stops on unknown names; add and rerun. `svn.authorsfile`. |
| `--authors-prog=<file>` | Program for names not in the authors file; prints `Name <email>`. Relative paths tried from current folder (`init`/`clone`) or repo top (`fetch`), then `$PATH`. `svn.authorsProg`. |
| `-q`, `--quiet` | Less output; twice for less. |
| `--log-window-size=<n>` | Log entries per request (default 100). Higher for huge repos; too high costs memory, may time out. |
| `--no-checkout` | Don't update the working tree after fetch. |
| `--use-log-author` | Author from the first `From:` or `Signed-off-by:` in the SVN message. `svn.useLogAuthor`. |
| `--add-author-from` | When committing to SVN, add `From:` with the git author if no `From:`/`Signed-off-by:`. Pairs with `--use-log-author`. `svn.addAuthorFrom`. |
| `--localtime` | Store local-zone times, not UTC, so `git log` matches `svn log`. Share only within that zone. |
| `--repack[=<n>]`, `--repack-flags=<flags>`, `--repack-args=<flags>`, `--repack-opts=<flags>` | Obsolete; warn, no effect. |
| `--noMetadata`, `--useSvmProps`, `--useSvnsyncProps` | Command-line forms of same-named config (see `init`'s `--no-metadata` etc.). |

### Commit flags

| Flag | What it does |
|------|--------------|
| `-e`, `--edit` | Edit the message before SVN commit. Default off for commits, on for trees. `svn.edit`. |
| `--rmdir` | Delete SVN folders left empty. `svn.rmdir`. |
| `--find-copies-harder` | Passed to `git diff-tree` for copy detection. `svn.findcopiesharder`. |
| `-l<num>` | Rename-detection limit for `git diff-tree`. `svn.l`. |
| `-C<num>`, `--copy-similarity=<num>` | Copy similarity for `git diff-tree` (`-C<num>`). |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `init` | Set up tracking of an SVN URL, no download. |
| `fetch` | Download new SVN revisions. |
| `clone` | `init` + `fetch` into a new folder. |
| `rebase` | Fetch, then rebase unsent commits on top (like `svn update`). |
| `dcommit` | Send each commit to SVN as a revision. |
| `branch` | Create an SVN branch. |
| `tag` | Create an SVN tag. |
| `log` | `svn log`-style history by revision number. |
| `blame` | `svn blame`-style line history. |
| `find-rev` | Convert SVN revisions to git commits and back. |
| `set-tree` | Overwrite SVN with given trees. Use `dcommit`. |
| `create-ignore` | Write `.gitignore` files from SVN ignore properties. |
| `show-ignore` | Print SVN ignores as gitignore patterns. |
| `mkdirs` | Recreate empty folders git can't track. |
| `commit-diff` | Commit the diff of two trees to SVN. |
| `info` | Like `svn info`. |
| `proplist` | List a path's SVN properties. |
| `propget` | Print one SVN property. |
| `propset` | Set an SVN property (applied next commit). |
| `show-externals` | Show `svn:externals`. |
| `gc` | Compress and clean git-svn's log and index files. |
| `reset` | Undo fetches back to a revision, to refetch. |
| `multi-init` | Deprecated: `init -T<trunk> -b<branches> -t<tags>`. |
| `multi-fetch` | Deprecated: `fetch --all`. |
| `migrate` | No-op (migration is automatic). |

### git svn init

**Usage:** `git svn init [<options>] <url> [<directory>]`

Sets up the current folder (or `<directory>`) to track SVN. URL here or as full URLs to `-T`/`-t`/`-b`. Takes connection flags plus:

| Flag | What it does |
|------|--------------|
| `-T<trunk-subdir>`, `--trunk=<trunk-subdir>` | Trunk path under URL, or full URL. |
| `-t<tags-subdir>`, `--tags=<tags-subdir>` | Tags folder. Repeatable. |
| `-b<branches-subdir>`, `--branches=<branches-subdir>` | Branches folder. Repeatable. |
| `-s`, `--stdlayout` | `-T trunk -t tags -b branches`. Explicit `-T`/`-t`/`-b` win. |
| `--prefix=<prefix>` | SVN refs under `refs/remotes/<prefix>`. Default `origin/`. Keep the trailing `/` (required with `-b`). `--prefix ""` = pre-2.0 layout. |
| `--[no-]minimize-url`, `-m` | Connect at the highest allowed SVN level when tracking several folders, to follow moves (default). `--no-minimize-url`: URLs as given, for read-restricted repos. |
| `--no-metadata` | No `git-svn-id:` lines. One-time imports only: no further fetch, rev map rebuild, or `git svn log`. Not recommended; prefer `git filter-repo`. |
| `--use-svm-props` | Map URLs/UUIDs back for SVN::Mirror or svk mirrors. |
| `--use-svnsync-props` | Same, for `svnsync` mirrors. |
| `--rewrite-root=<URL>` | Record this URL instead of the fetch URL (e.g. public `http://` for local `file://`). |
| `--rewrite-uuid=<UUID>` | Record this repo UUID instead. |
| `--shared[=(false\|true\|umask\|group\|all\|world\|everybody)]` | Passed to `git init`. |
| `--template=<template-directory>` | Passed to `git init`. |

Metadata settings (`no-metadata`, `rewrite-root`, `rewrite-uuid`, `use-svnsync-props`, `use-svm-props`) must be set before the first fetch and never changed. One per remote, except `rewrite-root` + `rewrite-uuid` together.

### git svn fetch

**Usage:** `git svn fetch [<options>] [<svn-remote>]`

Download unfetched revisions for the default or named `[svn-remote]`. Takes fetch flags plus:

| Flag | What it does |
|------|--------------|
| `-r <arg>`, `--revision=<arg>` | Only these revisions: `N`, `N1:N2`, `N:HEAD`, `BASE:N`. Partial mirror; skipped history lost. |
| `--all`, `--fetch-all` | Fetch every `[svn-remote]`. |
| `-p`, `--parent` | Fetch only from `HEAD`'s SVN parent. |

### git svn clone

**Usage:** `git svn clone [<options>] <url> [<directory>]`

Makes a folder named after the URL (or `<directory>`), then `init` and `fetch`. Without a layout flag, branches and tags import as plain folders in one linear (possibly huge) history; use `--stdlayout` or clone only `trunk`. Takes all `init` flags, fetch flags, `-r`/`--revision`, plus:

| Flag | What it does |
|------|--------------|
| `--preserve-empty-dirs` | Put a placeholder in each empty SVN folder so git keeps it; removed when unneeded. |
| `--placeholder-filename=<filename>` | Placeholder name. Default `.gitignore`. |

### git svn rebase

**Usage:** `git svn rebase [<options>]`

Fetch `HEAD`'s SVN parent and rebase unsent commits onto it. Needs a clean working tree. Takes fetch flags plus:

| Flag | What it does |
|------|--------------|
| `-l`, `--local` | Don't fetch; rebase onto the last fetch. |
| `--all`, `--fetch-all` | Fetch all branches of the current `[svn-remote]`. |
| `-m`, `-M`, `--merge` | Passed to `git rebase`. |
| `-s <strategy>`, `--strategy=<strategy>` | Merge strategy for `git rebase`. |
| `-p`, `--rebase-merges` | Passed to `git rebase`. |
| `-v`, `--verbose` | More output. |
| `-n`, `--dry-run` | Show the branch and SVN URL to be used. |

### git svn dcommit

**Usage:** `git svn dcommit [<options>] [<branch-or-commit>]`

Send each commit on the current (or named) branch to SVN as a revision, then rebase or reset onto the result. Builds on the SVN commit found by `git log --grep=^git-svn-id: --first-parent -1`, so after a merge the newest SVN commit must be the first parent. Takes fetch and commit flags plus:

| Flag | What it does |
|------|--------------|
| `--no-rebase` | Don't rebase or reset afterward. |
| `--commit-url <URL>` | Commit to this full URL (e.g. `svn+ssh://` when fetching over read-only `http://`). Config `svn-remote.<name>.commiturl` or `svn.commiturl`; whole repo: `svn-remote.<name>.pushurl`. |
| `--mergeinfo=<mergeinfo>` | Set `svn:mergeinfo`, e.g. `"/branches/foo:1-10 /branches/bar:3,5-6,8"`. `svn.pushmergeinfo` fills it automatically where possible. |
| `-i`, `--interactive` | Per patch: `yes`, `no`, `all`, `quit`. `no`/`quit` stop before committing anything. |
| `-n`, `--dry-run` | Print git commands showing the diffs to be sent. |
| `-m`, `-M`, `--merge` | Passed to `git rebase` when reset isn't possible. |
| `-s <strategy>`, `--strategy=<strategy>` | Passed to `git rebase`. |
| `-r <n>`, `--revision=<n>` | SVN revision to use as base. Rarely needed. |
| `--all`, `--fetch-all` | Fetch all branches after each commit. |
| `-v`, `--verbose` | More output. |
| `--set-svn-props=<props>` | Accepted; no effect. |

### git svn branch

**Usage:** `git svn branch [<options>] <name>`

| Flag | What it does |
|------|--------------|
| `-m <msg>`, `--message=<msg>` | Commit message. |
| `-t`, `--tag` | Make a tag instead. |
| `-d<path>`, `--destination=<path>` | Branches (or tags) folder to use when several were given. Must match the left side of a `git config --get-all svn-remote.<name>.branches` (or `.tags`) entry. |
| `--username=<user>` | SVN user to commit as. |
| `--commit-url=<URL>` | URL to write to (e.g. fetch URL read-only). Overrides `commiturl`. |
| `--parents` | Create missing parents (like `svn cp --parents`). |
| `-n`, `--dry-run` | Show the copy's URLs. |

### git svn tag

**Usage:** `git svn tag [<options>] <name>`

`git svn branch -t`. Same flags except `-t`.

### git svn log

**Usage:** `git svn log [<options>] [<path>...]`

History by SVN revision number, local times like `svn`. Unknown flags go to `git log`.

| Flag | What it does |
|------|--------------|
| `-r <n>[:<n>]`, `--revision=<n>[:<n>]` | Revision or range. Numbers only (no `HEAD`, `PREV`, ...). |
| `-v`, `--verbose` | List changed files, like `svn log -v`. |
| `--limit=<n>` | Show `n` revisions. Unlike `--max-count`, skipped/merged commits don't count. |
| `--incremental` | Concatenable output, like `svn log --incremental`. |
| `--show-commit` | Also show git hash. |
| `--oneline` | One line per revision. |
| `--non-recursive` | With `-v`, skip subfolder changes. |
| `-A <file>`, `--authors-file=<file>` | Map SVN users via the authors file. |
| `--color` | Color output. |
| `--pager=<command>` | Accepted; pager still from `GIT_PAGER`/`core.pager`. |

### git svn blame

**Usage:** `git svn blame [<options>] <file>`

SVN revision and author per line as of `HEAD` (local edits ignored), `svn blame` format. Unknown flags go to `git blame`.

| Flag | What it does |
|------|--------------|
| `--git-format` | `git blame` format with SVN revisions. Unsent changes show as revision 0. |

### git svn find-rev

**Usage:** `git svn find-rev (r<n> [<tree-ish>] | <tree-ish>)`

`r<n>` prints the git commit (optionally only on `<tree-ish>`'s branch). A commit prints its SVN revision.

| Flag | What it does |
|------|--------------|
| `-B`, `--before` | Inexact: commit current at that revision. |
| `-A`, `--after` | Inexact: nearest later match. |

### git svn set-tree

**Usage:** `git svn set-tree [<options>] <commit-or-tree>...`

Write trees to SVN as is, overwriting files without patching. Assumes everything's merged and fetch is current. Prefer `dcommit`. Takes fetch and commit flags plus:

| Flag | What it does |
|------|--------------|
| `--stdin`, `-` | Read commits from stdin (first hash per line; `git rev-list --pretty=oneline` works), send oldest first. Also the default with no commits given. |

### git svn create-ignore

**Usage:** `git svn create-ignore [-r <n>]`

Write a `.gitignore` per folder from `svn:ignore` and `svn:global-ignores` and stage them (no commit).

| Flag | What it does |
|------|--------------|
| `-r <n>`, `--revision=<n>` | Properties as of this revision. |

### git svn show-ignore

**Usage:** `git svn show-ignore [-r <n>]`

Print those patterns: `git svn show-ignore >> .git/info/exclude`.

| Flag | What it does |
|------|--------------|
| `-r <n>`, `--revision=<n>` | Properties as of this revision. |

### git svn mkdirs

**Usage:** `git svn mkdirs [-r <n>]`

Recreate empty folders listed in `$GIT_DIR/svn/<refname>/unhandled.log`. `clone`/`rebase` do it already; use after `git checkout` or `git reset`. Disable the automatic version with `svn-remote.<name>.automkdirs=false`.

| Flag | What it does |
|------|--------------|
| `-r <n>`, `--revision=<n>` | Folders as of this revision. |

### git svn commit-diff

**Usage:** `git svn commit-diff -r<revision> [<options>] <old-tree> <new-tree> [<url>]`

Commit the diff of two trees to SVN. Works outside git-svn repos given a URL. Message from `-m`, `-F`, the tag/commit `<new-tree>` names, or editor. Takes commit flags plus:

| Flag | What it does |
|------|--------------|
| `-r <revision>`, `--revision=<revision>` | Base SVN revision. Required. |
| `-m <msg>`, `--message=<msg>` | Message. Disables `--edit`. |
| `-F <filename>`, `--file=<filename>` | Message from file. Disables `--edit`. |

### git svn info

**Usage:** `git svn info [--url] [<path>]`

Like `svn info`. No `-r`.

| Flag | What it does |
|------|--------------|
| `--url` | Only the URL. |

### git svn proplist

**Usage:** `git svn proplist [-r <n>] [<path>]`

| Flag | What it does |
|------|--------------|
| `-r <n>`, `--revision=<n>` | As of this revision. |

### git svn propget

**Usage:** `git svn propget [-r <n>] <property> [<path>]`

| Flag | What it does |
|------|--------------|
| `-r <n>`, `--revision=<n>` | As of this revision. |

### git svn propset

**Usage:** `git svn propset <property> <value> <path>`

Example: `git svn propset svn:keywords "FreeBSD=%H" devel/py-tipper/Makefile`. No flags.

### git svn show-externals

**Usage:** `git svn show-externals [-r <n>]`

| Flag | What it does |
|------|--------------|
| `-r <n>`, `--revision=<n>` | As of this revision. |

### git svn gc

**Usage:** `git svn gc`

Compress `$GIT_DIR/svn/<refname>/unhandled.log`, delete `$GIT_DIR/svn/<refname>/index`. No flags.

### git svn reset

**Usage:** `git svn reset -r <n> [-p]`

Rewind fetched SVN history to refetch, e.g. after `--ignore-paths` or permission changes cause "not found in commit" or "checksum mismatch". Changes only the rev map and SVN tracking ref. Then `git svn fetch` and move your branch with `git rebase --onto` (not merge).

| Flag | What it does |
|------|--------------|
| `-r <n>`, `--revision=<n>` | Newest revision kept. |
| `-p`, `--parent` | Drop revision `n` too, keep its parent. |

### git svn multi-init

Deprecated: `git svn init -T<trunk> -b<branches> -t<tags>`. Takes `init` flags.

### git svn multi-fetch

Deprecated: `git svn fetch --all`. Takes fetch flags and `-r`/`--revision`.

### git svn migrate

No-op for old scripts. Takes connection flags and `--minimize`.

## Config-only settings

| Config | What it does |
|--------|--------------|
| `svn.noMetadata`, `svn-remote.<name>.noMetadata` | `init --no-metadata`. |
| `svn.useSvmProps`, `svn-remote.<name>.useSvmProps` | `init --use-svm-props`. |
| `svn.useSvnsyncProps`, `svn-remote.<name>.useSvnsyncprops` | `init --use-svnsync-props`. |
| `svn-remote.<name>.rewriteRoot`, `.rewriteUUID` | `init --rewrite-root` / `--rewrite-uuid`. |
| `svn-remote.<name>.pushurl` | Write URL when `url` is read-only. `commiturl` wins. |
| `svn.brokenSymlinkWorkaround` | `false` skips slow broken-symlink checks (default `true`). |
| `svn.pathnameencoding` | Re-encode file names (any Perl `Encode` name). |
| `svn-remote.<name>.automkdirs` | `false` stops `clone`/`rebase` recreating empty folders. |
| `svn-remote.<name>.branches`, `.tags`, `.fetch` | Mappings like `branches/*/project-a:refs/remotes/project-a/branches/*`. One `*` per part; local `*` last. `{red,green}` picks a subset. |
