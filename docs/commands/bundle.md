# git bundle

Pack commits and refs into one file for offline transfer (USB, email). Others can `git clone`, `fetch`, `pull`, or `ls-remote` from it like a remote; can't push into it.

Holds only refs and their commits: no index, working files, stash, config, or hooks. A range bundle (`old..new`) needs the receiver to have `old`.

**Usage:** `git bundle <subcommand> [<flags>] <file> ...` (`<file>` can be `-` for stdin/stdout)

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `create` | Write a bundle. |
| `verify` | Check validity and that this repo has the needed commits. |
| `list-heads` | List the bundle's refs. |
| `unbundle` | Low-level: store objects in this repo, print refs. Used by `git fetch`. |

### git bundle create

**Usage:** `git bundle create [<flags>] <file> <rev-list-args>`

`<rev-list-args>` selects content like `git log` args; must name a ref. E.g. `main`, `--all`, `--branches --tags` (what a clone gets), `v1.0..main`, `--since=10.days main`, `-10 main`, `--stdin`. A bare commit ID or range ending in a non-ref (`main~10..main~5`) fails: "Refusing to create empty bundle."

| Flag | What it does |
|------|--------------|
| `-q`, `--[no-]quiet` | No progress meter. |
| `--[no-]progress` | Progress even if stderr isn't a terminal. |
| `--[no-]version=<n>` | Format `2` (SHA-1 only) or `3` (extensions). Default: oldest that fits the repo's hash. |

### git bundle verify

**Usage:** `git bundle verify [-q] <file>`

Lists missing prerequisites and capabilities. Exits non-zero if invalid or unusable.

| Flag | What it does |
|------|--------------|
| `-q`, `--[no-]quiet` | Don't print details. |

### git bundle list-heads

**Usage:** `git bundle list-heads <file> [<refname>...]`

| Flag | What it does |
|------|--------------|
| `<refname>...` | Only these refs. |

### git bundle unbundle

**Usage:** `git bundle unbundle [--progress] <file> [<refname>...]`

| Flag | What it does |
|------|--------------|
| `--[no-]progress` | Progress meter. |
| `<refname>...` | Print only these refs. |
