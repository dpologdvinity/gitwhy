# git describe

Name a commit by its nearest earlier tag: `v1.0.4-14-g2414721` = 14 commits after `v1.0.4`, at commit `2414721`. A tagged commit shows just the tag. Annotated tags only unless `--tags`/`--all`. Given a blob ID, prints `<commit>:<path>` where it first appears.

**Usage:** `git describe [--all] [--tags] [--contains] [--abbrev=<n>] [<commit-ish>...]`
or `git describe [--all] [--tags] [--contains] [--abbrev=<n>] --dirty[=<mark>]`
or `git describe <blob>`

## Flags

| Flag | What it does |
|------|--------------|
| `<commit-ish>...` | Commits to describe. Default `HEAD`. |
| `--[no-]dirty[=<mark>]` | Describe the working tree; append `<mark>` (default `-dirty`) if changed. |
| `--[no-]broken[=<mark>]` | Like `--dirty`; on a corrupt repo append `<mark>` (default `-broken`) instead of failing. |
| `--[no-]all` | Use any ref, e.g. `heads/main-7-g975b`. |
| `--[no-]tags` | Include lightweight tags. |
| `--[no-]contains` | Use the first tag *after* (containing) the commit. Implies `--tags`. |
| `--[no-]abbrev[=<n>]` | `<n>` ID digits (more if needed). `0`: tag name only. |
| `--[no-]candidates=<n>` | Consider the `<n>` newest tags (default 10). `0`: exact match only. |
| `--[no-]exact-match` | Only a tag right on the commit, else fail. Same as `--candidates=0`. |
| `--[no-]debug` | Print search details to stderr. |
| `--[no-]long` | Always long form, even on a tag (`v1.2-0-gdeadbee`). |
| `--[no-]match <pattern>` | Only tags matching this glob (no `refs/tags/`). Repeatable; `--no-match` clears. |
| `--[no-]exclude <pattern>` | Skip tags matching this glob. Repeatable; `--no-exclude` clears. |
| `--[no-]always` | Fall back to the short commit ID if no tag. |
| `--[no-]first-parent` | Follow only first parents, ignoring tags on merged branches. |
