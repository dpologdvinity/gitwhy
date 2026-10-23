# git branch

List, create, rename, copy, or delete branches. Creating doesn't switch; use `git switch <name>` (or `git switch -c <name>` to create and switch).

**Usage:**
- List: `git branch [<flags>] [-r | -a] [--list] [<pattern>...]`. Current branch marked `*` (green), other worktrees' branches `+` (cyan). Glob patterns need `--list`, else they're taken as a new branch name.
- Create: `git branch [<flags>] <branch-name> [<start-point>]`. `<start-point>` defaults to `HEAD`; `A...B` means their merge base. Starting from a remote-tracking branch sets it as upstream (see `--track`).
- Rename / copy: `git branch (-m | -M | -c | -C) [<old-branch>] <new-branch>`. Config and reflog go along. `<old-branch>` defaults to current.
- Delete: `git branch (-d | -D) [-r] <branch-name>...`. Also deletes its reflog. `-r` deletes remote-tracking branches (only useful if gone from the remote; see `git remote prune`).
- Upstream: `git branch (-u <upstream> | --unset-upstream) [<branch-name>]`.
- Description: `git branch --edit-description [<branch-name>]`.
- Clean up (2.56+): `git branch [--dry-run] (--delete-merged <pattern>)... [<branch-pattern>...]`.

## Flags

| Flag | What it does |
|------|--------------|
| `-d`, `--[no-]delete` | Delete a branch. Must be merged into its upstream (or `HEAD` if none). |
| `-D` | Delete even if unmerged (`--delete --force`). |
| `--[no-]create-reflog` | Keep a reflog, so `<branch>@{yesterday}` works. Usually already on (`core.logAllRefUpdates`). `--no-` only cancels an earlier `--create-reflog`. |
| `-f`, `--[no-]force` | Create: reset an existing branch to `<start-point>` (not if checked out in any worktree). Delete: even if unmerged or broken. Move/copy: overwrite the target. |
| `-m`, `--[no-]move` | Rename. |
| `-M` | Rename, overwriting the target (`--move --force`). |
| `-c`, `--[no-]copy` | Copy. |
| `-C` | Copy, overwriting the target (`--copy --force`). |
| `--[no-]color[=<when>]` | Color branches: `always` (default), `never`, `auto`. |
| `-i`, `--[no-]ignore-case` | Case-insensitive sort and filter. |
| `--[no-]omit-empty` | No newline for empty `--format` output. |
| `--[no-]column[=<style>]` | List in columns (not with `-v`). Bare = `always`. Style: see `column.branch`. |
| `--[no-]sort=<key>` | Sort by field (keys as in [`for-each-ref`](for-each-ref.md)); `-` prefix descends. Repeatable; last is primary. Default `branch.sort` or ref name (detached `HEAD`, local, remote). |
| `-r`, `--remotes` | List (or with `-d` delete) remote-tracking branches. |
| `-a`, `--all` | List local and remote-tracking branches. |
| `-l`, `--[no-]list` | List branches, optionally matching `<pattern>`s. |
| `--[no-]show-current` | Print current branch name. Nothing if detached. |
| `-v`, `-vv`, `--[no-]verbose` | Show commit ID, title, ahead/behind. Twice: also upstream name and worktree path. |
| `-q`, `--[no-]quiet` | No non-error messages on create/delete. |
| `--[no-]abbrev[=<n>]` | `-v` IDs at least `<n>` digits (default 7, `core.abbrev`). `--no-abbrev`: full IDs. |
| `-t`, `--[no-]track[=(direct\|inherit)]` | On create, set upstream. `direct` (default): `<start-point>` itself. `inherit`: copy `<start-point>`'s upstream. Unset: `branch.autoSetupMerge` decides (default: only from remote-tracking branches). `--no-track`: never. |
| `--[no-]recurse-submodules` | Experimental: also create in submodules (needs `submodule.propagateBranches`). Create only. |
| `--set-upstream` | Removed. Use `--track` or `--set-upstream-to`. |
| `-u`, `--[no-]set-upstream-to=<upstream>` | Set upstream of `<branch-name>` (default: current). |
| `--[no-]unset-upstream` | Remove upstream of `<branch-name>` (default: current). |
| `--[no-]edit-description` | Edit branch description (used by `format-patch`, `request-pull`, `merge`). |
| `--contains [<commit>]` | Only branches containing `<commit>` (default `HEAD`). |
| `--no-contains [<commit>]` | Only branches not containing `<commit>`. |
| `--merged [<commit>]` | Only branches merged into `<commit>` (default `HEAD`). |
| `--no-merged [<commit>]` | Only branches not merged into `<commit>`. |
| `--forked <branch>` | (2.56+) Only branches whose upstream matches a ref (`origin/main`), remote (`origin` = `origin/HEAD`'s target), or glob (`'origin/*'`). Repeatable. |
| `--[no-]points-at=<object>` | Only branches pointing at `<object>`. |
| `--[no-]format=<format>` | Output template with `%(fieldname)`, as in [`for-each-ref`](for-each-ref.md). |
| `--delete-merged <pattern>` | (2.56+) Delete local branches whose upstream matches `<pattern>` (ref, remote, or glob) and already contains their work. Repeatable; `<branch-pattern>`s limit candidates. Skips branches unmerged, checked out anywhere, with deleted upstream, that would update upstream on push, that are a kept branch's upstream, or with `branch.<name>.deleteMerged=false`. |
| `--dry-run` | (2.56+) With `--delete-merged`: only print what would be deleted. |
