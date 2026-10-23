# git show-branch

Show branches side by side as a text grid of which commits each has, back to where they meet. Max 26 branches.

**Usage:** `git show-branch [<flags>] [<rev> | <glob>...]` or `git show-branch (-g | --reflog)[=<n>[,<base>]] [--list] [<ref>]`

`<rev>`: branch, tag or commit. `<glob>`: ref pattern (`topic/*`). Default: `showBranch.default`, else all branches.

Output marks: `*` current branch, `!` other branch, `+` commit on this branch, `-` merge.

## Flags

| Flag | What it does |
|------|--------------|
| `-r`, `--[no-]remotes` | Remote-tracking branches. |
| `-a`, `--[no-]all` | Remote-tracking and local branches. |
| `--[no-]current` | Include the current branch even if not named. |
| `--topo-order` | Children before parents, each branch's commits together. |
| `--date-order` | Children before parents, else by date. |
| `--[no-]sparse` | Also show merges reachable from only one branch. |
| `--[no-]more=<n>` | Go `<n>` commits past the common ancestor. Negative: list refs only. |
| `--[no-]list` | Same as `--more=-1`. |
| `--[no-]merge-base` | Print possible merge bases instead. |
| `--[no-]independent` | Print only given refs no other given ref can reach. |
| `--no-name` | Hide names next to commits. |
| `--name` | Undo `--no-name`. |
| `--[no-]sha1-name` | Name commits by short ID, not `main~2`. |
| `--[no-]topics` | Only commits not on the first branch. |
| `-g`, `--reflog[=<n>[,<base>]] [<ref>]` | Show `<n>` newest reflog entries of `<ref>` (default current branch) from `<base>` (count or date). |
| `--color[=<when>]`, `--no-color` | Color the marks: `always` (default), `never`, `auto`. |

`--more`, `--list`, `--independent` and `--merge-base` can't be combined.
