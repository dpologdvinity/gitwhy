# git replay

Experimental: copy commits onto a new base (like `rebase`/`cherry-pick`) without touching working files or index. Works in bare repos. Updates refs atomically. Exit 0 success, 1 conflicts (not reported), other on error. Can't replay merges (2.56+: unless `--linearize`).

Examples:
- `git replay --onto=target origin/main..mybranch`: rebase `mybranch` onto `target`.
- `git replay --advance=target origin/main..mybranch`: cherry-pick those commits onto `target`, moving `target`.
- `git replay --contained --onto=origin/main origin/main..tipbranch`: rebase a stack of branches together.
- `git replay --revert=main topic~2..topic`: revert the last two commits of `topic` on `main`.

**Usage:** `git replay ([--contained] --onto=<newbase> | --advance=<branch> | --revert=<branch>) [--ref=<ref>] [--ref-action=<mode>] [--linearize] <revision-range>`

## Flags

Also takes [commit selection options](../shared/revision-options.md) to narrow `<revision-range>`.

| Flag | What it does |
|------|--------------|
| `<revision-range>` | Commits to replay, e.g. `A..B`, `^base b1 b2`. `--advance`/`--revert` need one tip. Changes already in the target are dropped. |
| `--onto=<newbase>` | Build on this commit; move branches in the range, like `git rebase --update-refs`. |
| `--advance=<branch>` | Build on this branch and move it. Like cherry-pick. |
| `--revert=<branch>` | Add commits undoing the range onto this branch and move it. Messages start `Revert`/`Reapply`; you're author. Not with `--onto`, `--advance`, `--contained`. |
| `--[no-]contained` | With `--onto`: also move every branch pointing into the range. |
| `--ref=<ref>` | Update this full ref (e.g. `refs/heads/x`) instead of the inferred one. With `--onto`, range needs one tip. Not with `--contained`. |
| `--ref-action=<mode>` | `update` (default): update refs. `print`: print `update <ref> <new> <old>` for `git update-ref --stdin`. Default from `replay.refAction`. |
| `--linearize` | (2.56+) Flatten: stack commits in a line, drop merges, like `git rebase --no-rebase-merges`. Ref on a merge moves to its last replayed ancestor. Not with `--contained` or multiple branches. |
