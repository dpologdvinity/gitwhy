# gitwhy: every git command, in plain English

One file per command, with every flag. Global options (`-C`, `-c`, `--no-pager`, ...): [`git`](commands/git.md).
Shared option sets: [diff options](shared/diff-options.md), [revision options](shared/revision-options.md).

## Everyday commands

| Command | What it does |
|---------|--------------|
| [`add`](commands/add.md) | Stage files' current content in the index for the next commit. |
| [`am`](commands/am.md) | Turn patch emails (mbox, Maildir, or stdin) into commits on the current branch, keeping each email's author, date, and message. |
| [`archive`](commands/archive.md) | Write one commit's or tree's tracked files as a tar or zip (stdout by default). |
| [`backfill`](commands/backfill.md) | Experimental: download blobs a blobless partial clone (`--filter=blob:none`) skipped, in large batches grouped by path. |
| [`bisect`](commands/bisect.md) | Binary-search history for the commit that introduced a bug. |
| [`branch`](commands/branch.md) | List, create, rename, copy, or delete branches. |
| [`bundle`](commands/bundle.md) | Pack commits and refs into one file for offline transfer (USB, email). |
| [`checkout`](commands/checkout.md) | Switch branches, or restore files from the index or a commit. |
| [`cherry-pick`](commands/cherry-pick.md) | Copy existing commits onto the current branch as new commits, e.g. |
| [`citool`](commands/citool.md) | GUI to stage files, write a message, and make one commit. |
| [`clean`](commands/clean.md) | Delete untracked files, from the current folder down. |
| [`clone`](commands/clone.md) | Copy a repo and its history into a new folder, with remote `origin` pointing back. |
| [`commit`](commands/commit.md) | Record the index as a new commit (child of `HEAD`) with a message and move the current branch to it. |
| [`describe`](commands/describe.md) | Name a commit by its nearest earlier tag: `v1.0.4-14-g2414721` = 14 commits after `v1.0.4`, at commit `2414721`. |
| [`diff`](commands/diff.md) | Show changes between working files, the index, commits, blobs, or any two files on disk. |
| [`fetch`](commands/fetch.md) | Download commits, branches, and tags from another repo without touching your branches or files. |
| [`format-patch`](commands/format-patch.md) | Turn commits into email-ready patch files, one per commit (`0001-fix-typo.patch`, ...). |
| [`gc`](commands/gc.md) | Clean up and compact the repo: pack objects, prune old unreachable ones, pack refs, expire reflogs, clean rerere and stale worktrees, update commit-graph. |
| [`gitk`](commands/gitk.md) | GUI history browser: commit graph, commit details and changes, files at each commit. |
| [`grep`](commands/grep.md) | Search tracked files for a pattern. |
| [`gui`](commands/gui.md) | GUI for committing: stage, write messages, amend, branch, merge, fetch, push. |
| [`history`](commands/history.md) | Experimental: edit one past commit and auto-replay the commits after it. |
| [`init`](commands/init.md) | Create a new repo (the `.git` folder) in the current folder or `<directory>`. |
| [`log`](commands/log.md) | Show commit history, newest first. |
| [`maintenance`](commands/maintenance.md) | Run repo upkeep jobs (packing, commit-graph, prefetch, ...) now or on an hourly/daily/weekly background schedule. |
| [`merge`](commands/merge.md) | Join other branches' changes into the current branch. |
| [`mv`](commands/mv.md) | Move or rename a file, directory, or symlink and stage it. |
| [`notes`](commands/notes.md) | Attach text to commits (or other objects) without changing their IDs. |
| [`pull`](commands/pull.md) | `git fetch`, then combine the fetched branch into the current one. |
| [`push`](commands/push.md) | Upload commits to another repo and update its branches or tags. |
| [`range-diff`](commands/range-diff.md) | Compare two versions of a commit series (e.g. |
| [`rebase`](commands/rebase.md) | Replay your commits, as new commits, on top of another starting point, giving straight-line history instead of a merge. |
| [`reset`](commands/reset.md) | Two jobs: move the current branch to another commit (optionally resetting index and files), or unstage files (opposite of `git add`). |
| [`restore`](commands/restore.md) | Restore files: discard unstaged edits (from the index), unstage (from `HEAD`), or take a file from any commit. |
| [`revert`](commands/revert.md) | Undo commits by adding new commits that reverse them. |
| [`rm`](commands/rm.md) | Delete tracked files from the working tree and index, staging the removal. |
| [`scalar`](commands/scalar.md) | Keep very large repos fast: tuned settings, background maintenance, partial downloads. |
| [`shortlog`](commands/shortlog.md) | Summarize history by author: commit count and subjects per person. |
| [`show`](commands/show.md) | Show objects. |
| [`sparse-checkout`](commands/sparse-checkout.md) | Experimental: check out only some folders of a big repo. |
| [`stash`](commands/stash.md) | Set uncommitted changes aside, leaving files matching `HEAD`; reapply later on any commit. |
| [`status`](commands/status.md) | Show staged changes, unstaged changes, and untracked files, plus the branch, ahead/behind counts, and any merge/rebase in progress. |
| [`submodule`](commands/submodule.md) | Manage submodules: other repos inside this one (the "superproject"), pinned to a commit. |
| [`switch`](commands/switch.md) | Switch branches (updates files and index), optionally creating one. |
| [`tag`](commands/tag.md) | Create, list, delete, or verify tags (fixed names for commits, usually releases). |
| [`worktree`](commands/worktree.md) | Have several working folders for one repo, each on a different branch. |

## Change settings and repo data

| Command | What it does |
|---------|--------------|
| [`config`](commands/config.md) | Read and change settings. |
| [`fast-export`](commands/fast-export.md) | Dump history (commits, files, tags, branches) as a text stream for [`git fast-import`](commands/fast-import.md). |
| [`fast-import`](commands/fast-import.md) | Build history from a text stream on stdin: commits, files, tags, branches. |
| [`filter-branch`](commands/filter-branch.md) | Not recommended: use [`git filter-repo`](commands/filter-repo.md). |
| [`mergetool`](commands/mergetool.md) | Open each conflicted file in a merge tool (vimdiff, meld, kdiff3, ...) to resolve it. |
| [`pack-refs`](commands/pack-refs.md) | Move loose refs (one file per branch/tag) into `.git/packed-refs`, saving space and time with many refs. |
| [`prune`](commands/prune.md) | Delete unreachable loose objects (not reachable from any ref or reflog), loose objects already packed, and stale `.git/shallow` entries. |
| [`reflog`](commands/reflog.md) | Show and manage reflogs: local logs of where `HEAD` and each branch pointed. |
| [`refs`](commands/refs.md) | Low-level ref tools: list, check, pack, convert storage format, and (2.56+) create, update, delete, rename. |
| [`remote`](commands/remote.md) | Manage remotes: named links to other copies of the repo (e.g. |
| [`repack`](commands/repack.md) | Pack loose objects into a pack file, or combine packs into fewer. |
| [`replace`](commands/replace.md) | Make git use one object in place of another, without rewriting history. |

## Inspect the repo

| Command | What it does |
|---------|--------------|
| [`annotate`](commands/annotate.md) | Show each line of a file with the commit that last changed it. |
| [`blame`](commands/blame.md) | Show the commit, author, and date that last changed each line of a file. |
| [`bugreport`](commands/bugreport.md) | Create a bug report file with blanks to fill in plus auto-collected info (`git version --build-options`, `uname`, compiler, enabled hooks, `$SHELL`). |
| [`count-objects`](commands/count-objects.md) | Count loose (unpacked) objects and their disk use. |
| [`diagnose`](commands/diagnose.md) | Zip up machine, git, and repo details for debugging. |
| [`difftool`](commands/difftool.md) | `git diff` that opens each changed file in a visual diff tool (Meld, vimdiff, VS Code, ...). |
| [`fsck`](commands/fsck.md) | Check the object store for corruption, missing objects, and unreachable objects. |
| [`gitweb`](commands/gitweb.md) | Web UI for git repos (branches, commits, files, diffs, blame, search, feeds). |
| [`help`](commands/help.md) | Show help. |
| [`instaweb`](commands/instaweb.md) | Browse the current repo in a browser via a local gitweb server. |
| [`merge-tree`](commands/merge-tree.md) | Merge two branches without touching the working tree, index, or refs; print the result tree ID. |
| [`rerere`](commands/rerere.md) | "Reuse recorded resolution": remember conflict fixes and reapply them when the same conflict recurs. |
| [`show-branch`](commands/show-branch.md) | Show branches side by side as a text grid of which commits each has, back to where they meet. |
| [`verify-commit`](commands/verify-commit.md) | Check commits' GPG signatures (from `git commit -S`). |
| [`verify-tag`](commands/verify-tag.md) | Check signatures (GPG, SSH, X.509) on signed tags. |
| [`version`](commands/version.md) | Print the git version. |
| [`whatchanged`](commands/whatchanged.md) | Deprecated, to be removed: use `git log --raw --no-merges`. |

## Work with other systems (SVN, Perforce, email, CVS)

| Command | What it does |
|---------|--------------|
| [`archimport`](commands/archimport.md) | Import a GNU Arch project into git via `tla` (must be installed). |
| [`cvsexportcommit`](commands/cvsexportcommit.md) | Apply one git commit to a CVS checkout. |
| [`cvsimport`](commands/cvsimport.md) | Import a CVS repo into git, fresh or on top of an earlier import. |
| [`cvsserver`](commands/cvsserver.md) | Act as a CVS server so CVS clients (`cvs`, Eclipse, TortoiseCVS) can check out and commit. |
| [`imap-send`](commands/imap-send.md) | Upload patches from an mbox (e.g. |
| [`p4`](commands/p4.md) | Use git with a Perforce (p4) project: clone a depot path, pull p4 changes, submit git commits as p4 changelists. |
| [`quiltimport`](commands/quiltimport.md) | Apply a quilt patch series as commits on the current branch, one per patch, in order. |
| [`request-pull`](commands/request-pull.md) | Print a pull request message: branch description, commit summary since `<start>`, and where to pull from. |
| [`send-email`](commands/send-email.md) | Email patches: patch files, folders of them, or a commit range (run through [`git format-patch`](commands/format-patch.md), whose flags also work). |
| [`svn`](commands/svn.md) | Use git with a Subversion project: clone SVN history, pull new revisions (`fetch`, `rebase`), and send git commits back as SVN revisions (`dcommit`). |

## Low-level: change objects and refs

| Command | What it does |
|---------|--------------|
| [`apply`](commands/apply.md) | Apply a patch to files, optionally the index. |
| [`checkout-index`](commands/checkout-index.md) | Plumbing: copy files from the index to the working tree. |
| [`commit-graph`](commands/commit-graph.md) | Write or verify the commit-graph: a history cache that speeds up `log`, `merge-base`, and path history. |
| [`commit-tree`](commands/commit-tree.md) | Low-level: make a commit object from a tree and print its ID. |
| [`hash-object`](commands/hash-object.md) | Print the object ID for a file's contents; optionally store it. |
| [`index-pack`](commands/index-pack.md) | Plumbing: build the `.idx` index (and optional `.rev`) for a packfile. |
| [`merge-file`](commands/merge-file.md) | Three-way merge of single files: apply `<base>`-to-`<other>` changes onto `<current>`, overwriting it. |
| [`merge-index`](commands/merge-index.md) | Plumbing: run `<merge-program>` (usually `git merge-one-file`) on each unmerged index file. |
| [`mktag`](commands/mktag.md) | Low-level: make an annotated tag object from stdin and print its ID. |
| [`mktree`](commands/mktree.md) | Plumbing: build a tree object from `git ls-tree`-format lines on stdin; print its ID. |
| [`multi-pack-index`](commands/multi-pack-index.md) | Plumbing: manage the multi-pack-index (MIDX), one index over many packs for faster lookups. |
| [`pack-objects`](commands/pack-objects.md) | Plumbing: build a packfile (compressed objects, many stored as deltas). |
| [`prune-packed`](commands/prune-packed.md) | Delete loose objects already in a pack. |
| [`read-tree`](commands/read-tree.md) | Low-level: load a tree into the index. |
| [`replay`](commands/replay.md) | Experimental: copy commits onto a new base (like `rebase`/`cherry-pick`) without touching working files or index. |
| [`symbolic-ref`](commands/symbolic-ref.md) | Read, set, or delete a symbolic ref (a ref pointing to another ref, like `HEAD` holding `ref: refs/heads/<branch>`). |
| [`unpack-objects`](commands/unpack-objects.md) | Read a pack from stdin and write each object as a loose file. |
| [`update-index`](commands/update-index.md) | Low-level: edit the index directly (stage files, set "assume unchanged" and "skip worktree" bits). |
| [`update-ref`](commands/update-ref.md) | Low-level: safely set or delete a ref (branch, tag, `HEAD`, ...), optionally checking its old value first. |
| [`write-tree`](commands/write-tree.md) | Low-level: save the index (staging area) as a tree object and print its ID. |

## Low-level: read objects and refs

| Command | What it does |
|---------|--------------|
| [`cat-file`](commands/cat-file.md) | Plumbing: print an object's contents, type, or size. |
| [`cherry`](commands/cherry.md) | Show which of your commits are already upstream. |
| [`diff-files`](commands/diff-files.md) | Plumbing: compare working files with the index. |
| [`diff-index`](commands/diff-index.md) | Plumbing: compare a tree (usually `HEAD`) with working files or the index. |
| [`diff-pairs`](commands/diff-pairs.md) | Plumbing: read NUL-separated raw file pairs on stdin (e.g. |
| [`diff-tree`](commands/diff-tree.md) | Plumbing: compare two trees or commits; one commit is compared with its parents. |
| [`for-each-ref`](commands/for-each-ref.md) | List refs with custom format, sorting, and filters. |
| [`for-each-repo`](commands/for-each-repo.md) | Experimental: run a git command in each repo listed in a config variable. |
| [`format-rev`](commands/format-rev.md) | Format commits from stdin with a `git log`-style format, one output per input, in order. |
| [`get-tar-commit-id`](commands/get-tar-commit-id.md) | Print the commit ID stored in a `git archive` tar read from stdin. |
| [`last-modified`](commands/last-modified.md) | Experimental: show the commit that last changed each file and folder (like GitHub's file listing). |
| [`ls-files`](commands/ls-files.md) | List tracked files, and optionally untracked, ignored, modified, or deleted ones. |
| [`ls-remote`](commands/ls-remote.md) | List a remote's refs and their commit IDs. |
| [`ls-tree`](commands/ls-tree.md) | List a tree's entries (mode, type, ID, name), like `ls` for a past version. |
| [`merge-base`](commands/merge-base.md) | Find the best common ancestor of commits (where histories split; a merge's starting point). |
| [`name-rev`](commands/name-rev.md) | Name a commit relative to a later branch or tag, e.g. |
| [`pack-redundant`](commands/pack-redundant.md) | Deprecated, slated for removal; use [`git gc`](commands/gc.md) or [`git repack`](commands/repack.md). |
| [`repo`](commands/repo.md) | Experimental: print repo facts (bare, shallow, hash type, paths) and size stats in script-friendly formats. |
| [`rev-list`](commands/rev-list.md) | List commit IDs reachable from given commits, newest first. |
| [`rev-parse`](commands/rev-parse.md) | Low-level, for scripts. |
| [`show-index`](commands/show-index.md) | Plumbing: read a pack index (`.idx`) from stdin and print each object's pack offset, ID, and (v2+) CRC32. |
| [`show-ref`](commands/show-ref.md) | List refs with their IDs (`<id> <ref>` per line), or check a ref exists. |
| [`unpack-file`](commands/unpack-file.md) | Write a blob to a temp file (`.merge_file_XXXXXX`) in the current directory and print its name. |
| [`var`](commands/var.md) | Print a computed git setting (editor, pager, author identity, ...). |
| [`verify-pack`](commands/verify-pack.md) | Check pack files for corruption. |

## Low-level: transfer data between repos

| Command | What it does |
|---------|--------------|
| [`daemon`](commands/daemon.md) | Tiny `git://` server (TCP 9418). |
| [`fetch-pack`](commands/fetch-pack.md) | Plumbing: download missing objects by talking to a remote `git upload-pack`. |
| [`http-backend`](commands/http-backend.md) | Server side: CGI program a web server runs to serve repos over HTTP(S): fetch/clone (smart and dumb), push (smart), protocol v2. |
| [`send-pack`](commands/send-pack.md) | Plumbing push: send objects and ref updates to a remote `git receive-pack`. |
| [`update-server-info`](commands/update-server-info.md) | Write `info/refs` and `objects/info/packs` so "dumb" HTTP/FTP servers (plain file serving) can be cloned from. |

## Internal helpers (git runs these for you)

| Command | What it does |
|---------|--------------|
| [`check-attr`](commands/check-attr.md) | Show which `.gitattributes` settings apply to files: `unspecified`, `unset`, `set`, or a value. |
| [`check-ignore`](commands/check-ignore.md) | Debug ignore rules: print the given paths that `.gitignore`, `.git/info/exclude`, or `core.excludesFile` ignores. |
| [`check-mailmap`](commands/check-mailmap.md) | Print each contact's canonical `Name <email>` per `.mailmap` (which maps old or misspelled identities to one), or the input unchanged if unmapped. |
| [`check-ref-format`](commands/check-ref-format.md) | Check that a name is a valid ref name; exit non-zero if not. |
| [`column`](commands/column.md) | Lay out stdin lines in columns. |
| [`credential`](commands/credential.md) | Get, save, or forget credentials the way git does (config, helpers like keychains, or prompting). |
| [`credential-cache`](commands/credential-cache.md) | Credential helper that keeps passwords in memory for a while. |
| [`credential-store`](commands/credential-store.md) | Credential helper that saves logins to a plain-text file. |
| [`fmt-merge-msg`](commands/fmt-merge-msg.md) | Mostly internal: print the default merge message (`Merge branch 'x' of ...`) from a `FETCH_HEAD`-format list on stdin. |
| [`hook`](commands/hook.md) | Run or list hooks (scripts git runs at events like `pre-commit`). |
| [`http-fetch`](commands/http-fetch.md) | Low-level: download objects over "dumb" HTTP. |
| [`http-push`](commands/http-push.md) | Push over "dumb" HTTP/WebDAV (no git server). |
| [`interpret-trailers`](commands/interpret-trailers.md) | Add or read trailers: `Key: value` lines ending a commit message (`Signed-off-by: Alice <a@x.com>`, `Co-authored-by: ...`). |
| [`mailinfo`](commands/mailinfo.md) | Low-level: split one email from stdin into commit message `<msg>` and patch `<patch>`; print author name, email, subject. |
| [`mailsplit`](commands/mailsplit.md) | Split an mbox or Maildir into one file per email (`0001`, `0002`, ...). |
| [`merge-octopus`](commands/merge-octopus.md) | Internal: the "octopus" merge strategy, run by `git merge -s octopus` (default when merging 3+ branches). |
| [`merge-one-file`](commands/merge-one-file.md) | Internal: the default program [`git merge-index`](commands/merge-index.md) runs to merge one file (adds, deletes, mode changes, three-way content merge with conflict markers). |
| [`merge-ours`](commands/merge-ours.md) | Internal command run by `git merge -s ours`. |
| [`merge-recursive`](commands/merge-recursive.md) | Internal: `git merge` runs it for the `recursive` strategy, which since git 2.50 means `ort` (the default). |
| [`merge-recursive-ours`](commands/merge-recursive-ours.md) | Internal: [`git merge-recursive`](commands/merge-recursive.md) with `--ours`. |
| [`merge-recursive-theirs`](commands/merge-recursive-theirs.md) | Internal command: [`git merge-recursive`](commands/merge-recursive.md) with `--theirs` (`git merge -X theirs`). |
| [`merge-resolve`](commands/merge-resolve.md) | Internal: the "resolve" strategy, run by `git merge -s resolve`. |
| [`merge-subtree`](commands/merge-subtree.md) | Internal: the "subtree" merge strategy, run by `git merge -s subtree`. |
| [`patch-id`](commands/patch-id.md) | Print an ID per patch read from stdin, based only on the changes (line numbers ignored). |
| [`receive-pack`](commands/receive-pack.md) | Internal: server side of a push. |
| [`remote-ext`](commands/remote-ext.md) | Internal: handles `ext::` URLs by running your command and talking to the git server over its stdin/stdout. |
| [`remote-fd`](commands/remote-fd.md) | Internal: handles `fd::<infd>[,<outfd>][/<anything>]` URLs, talking to a git server over already-open file descriptors. |
| [`remote-ftp`](commands/remote-ftp.md) | Internal helper git runs for `ftp://` remotes (fetch, clone, `ls-remote` only). |
| [`remote-ftps`](commands/remote-ftps.md) | Internal: handles `ftps://` remote URLs for `git fetch`, `git clone`, `git ls-remote` (fetch only). |
| [`remote-http`](commands/remote-http.md) | Internal: handles `http://` remote URLs for `git fetch`, `git clone`, `git push`, `git ls-remote`. |
| [`remote-https`](commands/remote-https.md) | Internal helper git runs for `https://` remotes during fetch, clone, push, and ls-remote. |
| [`sh-i18n`](commands/sh-i18n.md) | Not a command: a shell library git's scripts source for translated messages. |
| [`sh-setup`](commands/sh-setup.md) | Not a command: a shell library git's scripts source to find the repo and get helper functions. |
| [`shell`](commands/shell.md) | Restricted login shell for SSH accounts that only push and fetch. |
| [`stripspace`](commands/stripspace.md) | Clean stdin text like git cleans commit messages: strip trailing whitespace, squeeze blank-line runs, drop leading/trailing blank lines, add a final newline. |
| [`subtree`](commands/subtree.md) | Keep another project in a subdirectory, with or without its history, and move changes both ways: merge upstream updates in, or split the folder's changes out as a separate history to push. |
| [`upload-archive`](commands/upload-archive.md) | Internal: server side of `git archive --remote`. |
| [`upload-pack`](commands/upload-pack.md) | Internal: server side of fetch and clone. |
| [`url-parse`](commands/url-parse.md) | Split git URLs (including scp-style `user@host:path`) into parts, or check they parse. |

## Separate tools

| Command | What it does |
|---------|--------------|
| [`filter-repo`](commands/filter-repo.md) | Rewrite all history fast: remove files or secrets, move into a subfolder, fix authors, rename tags, and more. |
| [`lfs`](commands/lfs.md) | Git LFS (Large File Storage), a separate add-on: big files are stored as small "pointers" in the repo, with content on an LFS server and in `.git/lfs/objects`. |

## Old names and aliases

| Command | What it does |
|---------|--------------|
| [`fsck-objects`](commands/fsck-objects.md) | Old name for [`git fsck`](commands/fsck.md); use that. |
| [`init-db`](commands/init-db.md) | Old name for [`git init`](commands/init.md); use that. |
| [`pickaxe`](commands/pickaxe.md) | Old alias for [`git blame`](commands/blame.md); use that. |
| [`stage`](commands/stage.md) | Alias for `git add`; same flags. |
