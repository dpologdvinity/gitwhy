# git name-rev

Name a commit relative to a later branch or tag, e.g. `tags/v0.99~940` (940 commits before `v0.99`). Like `git describe`, but counts back from refs after the commit.

**Usage:** `git name-rev [<flags>] (--all | --annotate-stdin | <commit-ish>...)`

## Flags

| Flag | What it does |
|------|--------------|
| `<commit-ish>...` | Commits to name. |
| `--[no-]tags` | Only use tags. |
| `--[no-]refs=<pattern>` | Only use refs matching this shell pattern. Repeatable; `--no-refs` clears. |
| `--[no-]exclude=<pattern>` | Never use matching refs. Repeatable; `--no-exclude` clears. |
| `--[no-]all` | Name every commit reachable from any ref. |
| `--[no-]annotate-stdin` | Copy stdin to stdout, adding ` (<name>)` after each full-length commit ID. E.g. `git log \| git name-rev --annotate-stdin`. Old name: `--stdin`. |
| `--[no-]name-only` | Print only the name. With `--tags`, drop `tags/`. With `--annotate-stdin`, replace the ID. |
| `--[no-]undefined` | Print `undefined` for unnameable commits (default). `--no-undefined`: error instead. |
| `--[no-]always` | Print short ID for unnameable commits. |
