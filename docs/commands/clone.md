# git clone

Copy a repo and its history into a new folder, with remote `origin` pointing back. Remote branches become `origin/*`; the remote's default branch is checked out and tracked.

**Usage:** `git clone [<flags>] [--] <repository> [<directory>]`

Examples: `git clone https://github.com/git/git.git`; `git clone --depth 1 <url>` (latest commit only); `git clone --filter=blob:none <url>` (file contents on demand).

## Flags

| Flag | What it does |
|------|--------------|
| `<repository>` | URL, local path, or bundle file (see [URLs](fetch.md#urls)). |
| `<directory>` | New folder (must be new or empty). Default: repo name (`repo` for `/path/to/repo.git`). |
| `-l`, `--[no-]local` | Local source: copy `.git` directly, hard-linking objects. Default for plain paths; ignored for URLs (incl. `file://`). `--no-local` forces the protocol (needed for another user's repo). |
| `--no-hardlinks` | Local clone: copy objects, don't hard-link. For backups. |
| `--hardlinks` | Undo `--no-hardlinks`. |
| `-s`, `--[no-]shared` | Local clone: borrow objects from the source via alternates, copying none. Risky: breaks if the source prunes objects. `git repack -a` ends borrowing. |
| `--[no-]reference=<repo>` | Borrow objects from local `<repo>` to download less. Same risk as `--shared`; see `--dissociate`. |
| `--[no-]reference-if-able=<repo>` | Like `--reference`; warn instead of fail if `<repo>` is missing. |
| `--[no-]dissociate` | Borrow only during the clone, then copy what's needed and stop. |
| `-q`, `--[no-]quiet` | No progress. |
| `-v`, `--[no-]verbose` | More output. |
| `--[no-]progress` | Force progress output. |
| `--[no-]server-option=<option>` | Send `<option>` to the server (protocol v2). Repeatable. Default: `remote.<name>.serverOption`. |
| `-n`, `--no-checkout` | Don't check out files. |
| `--checkout` | Undo `--no-checkout`. |
| `--[no-]reject-shallow` | Fail if the source is shallow. Default: `clone.rejectShallow`. |
| `--[no-]bare` | Bare repo: no working files; `<directory>` is the git data. Remote branches become local branches, no tracking. |
| `--[no-]sparse` | Sparse checkout of top-level files only. Add folders with [`git sparse-checkout`](sparse-checkout.md). |
| `--[no-]filter=<filter-spec>` | Partial clone. `blob:none`: no file contents until needed. `blob:limit=<size>`: skip files of `<size>`+. `auto`: use the server's suggested filter (see `promisor.acceptFromServer`). More: `git rev-list --filter`. |
| `--[no-]also-filter-submodules` | Apply `--filter` to submodules. Needs `--filter` and `--recurse-submodules`. Default: `clone.filterSubmodules`. |
| `--[no-]mirror` | Like `--bare`, but copies *all* refs and sets `git remote update` to overwrite them. |
| `-o <name>`, `--[no-]origin=<name>` | Remote name instead of `origin`. Overrides `clone.defaultRemoteName`. |
| `-b <name>`, `--[no-]branch=<name>` | Check out `<name>` instead of the default branch. A tag gives a detached `HEAD`. |
| `--[no-]revision=<rev>` | Fetch only history up to `<rev>` (ref or commit ID); detach `HEAD` there; no branches. Not with `--branch` or `--mirror`. |
| `-u <upload-pack>`, `--[no-]upload-pack=<upload-pack>` | SSH: path to `git-upload-pack` on the server. |
| `--[no-]template=<template-directory>` | Copy hooks etc. from here into `.git`. See [`git init`](init.md). |
| `-c <key>=<value>`, `--[no-]config=<key>=<value>` | Set config in the new repo before fetching. Repeat a key to add values. Not for `remote.<name>.mirror`/`tagOpt`; use `--mirror`/`--no-tags`. |
| `--[no-]depth=<depth>` | Shallow: last `<depth>` commits. Implies `--single-branch` unless `--no-single-branch`. |
| `--[no-]shallow-since=<date>` | Shallow: commits after `<date>`. |
| `--[no-]shallow-exclude=<ref>` | Shallow: omit commits reachable from this remote ref. Repeatable. |
| `--[no-]single-branch` | Fetch one branch only (`--branch` or default), now and in later fetches. |
| `--[no-]tags` | Fetch tags (default). `--no-tags` skips them in later fetches too (sets `remote.<name>.tagOpt`). |
| `--[no-]recurse-submodules[=<pathspec>]` | Also clone submodules (all or matching; repeatable). Sets `submodule.active`. Ignored with `-n`, `--bare`, `--mirror`. |
| `--[no-]recursive[=<pathspec>]` | Old name for `--recurse-submodules`. |
| `--[no-]shallow-submodules` | Clone submodules at depth 1. |
| `--[no-]remote-submodules` | Submodules at their remote-tracking branch, not the recorded commit. |
| `--[no-]separate-git-dir=<git-dir>` | Put git data at `<git-dir>`; leave a `.git` file pointing there. |
| `--[no-]ref-format=<ref-format>` | Ref storage: `files` (default) or `reftable`. |
| `-j <n>`, `--[no-]jobs=<n>` | Clone `<n>` submodules at once. Default: `submodule.fetchJobs`. |
| `-4`, `--ipv4` | IPv4 only. |
| `-6`, `--ipv6` | IPv6 only. |
| `--[no-]bundle-uri=<uri>` | Seed from a bundle at `<uri>`, then fetch the rest. Not with `--depth`, `--shallow-since`, `--shallow-exclude`. |
| `--` | Rest is `<repository>` and `<directory>`. |

## More examples

| Command | What it does |
|---------|--------------|
| `git clone -l -s -n . ../copy` | Local copy borrowing objects, no checkout. |
| `git clone --reference /git/linux.git <url> my-linux` | Clone, borrowing objects you have locally. |
| `git clone --bare -l /home/proj/.git /pub/scm/proj.git` | Bare repo to publish from. |
| `git clone --no-local /home/otheruser/proj.git /pub/scm/proj.git` | Clone another user's local repo. |
