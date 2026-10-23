# git range-diff

Compare two versions of a commit series (e.g. a branch before and after rebase): pair old and new commits, show how each pair differs. Skips merges unless `--diff-merges`. Output is for humans; may change between versions.

Line format: `old-number: old-id <mark> new-number: new-id title`. Marks: `=` same, `!` changed (diff follows), `<` only old, `>` only new.

**Usage:**
`git range-diff [<flags>] <old-base>..<old-tip> <new-base>..<new-tip> [[--] <path>...]` two ranges (`<rev>^!` and `<rev>^-<n>` also work)
`git range-diff [<flags>] <old-tip>...<new-tip> [[--] <path>...]` same as `<new-tip>..<old-tip> <old-tip>..<new-tip>`
`git range-diff [<flags>] <base> <old-tip> <new-tip> [[--] <path>...]` same as `<base>..<old-tip> <base>..<new-tip>`

Example: after a rebase, `git range-diff @{u} @{1} @`. `<path>` limits both ranges.

## Flags

Also takes all [diff options](../shared/diff-options.md); they shape the diff *between* patches.

| Flag | What it does |
|------|--------------|
| `--[no-]dual-color` | Color inner diff lines like the original patches; dim old-only lines, bold new-only. On by default; `--no-dual-color` colors by outer markers only. |
| `--[no-]creation-factor=<percent>` | How readily commits pair vs. count as delete+add. Default 60. Raise if a changed commit shows as delete+add. |
| `--[no-]left-only` | Hide commits missing from the old range. |
| `--[no-]right-only` | Hide commits missing from the new range. |
| `--[no-]diff-merges=<format>` | Include merges, diffed like `git log --diff-merges=<format>`. `remerge` usually best. |
| `--[no-]remerge-diff` | Same as `--diff-merges=remerge`. |
| `--[no-]notes[=<ref>]` | Include notes in the patches. |
| `--[no-]max-memory=<size>` | Memory cap for commit pairing. Default 4G. |
