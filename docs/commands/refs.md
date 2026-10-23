# git refs

Low-level ref tools: list, check, pack, convert storage format, and (2.56+) create, update, delete, rename.

**Usage:** `git refs <subcommand> [<flags>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `migrate` | Convert ref storage to `files` or `reftable`. |
| `verify` | Check the ref database. |
| `list` | List refs. Same as [`git for-each-ref`](for-each-ref.md). |
| `exists` | Exit 0 if the ref exists, 2 if missing, 1 on error. Doesn't check its object. |
| `optimize` | Pack refs. Same as [`git pack-refs`](pack-refs.md). |
| `create` | (2.56+) Create a ref; fails if it exists. |
| `update` | (2.56+) Set a ref, optionally only if it has an expected old value. |
| `delete` | (2.56+) Delete a ref, optionally only if it has an expected value. Like `git update-ref -d`. |
| `rename` | (2.56+) Rename a ref. Old must exist, new must not. |

### git refs migrate

`git refs migrate --ref-format=<format> [<flags>]`. Not for repos with extra worktrees. Doesn't lock out other writers; stop them (and unregister `git maintenance`) first.

| Flag | What it does |
|------|--------------|
| `--ref-format=<format>` | Target: `files` or `reftable`. Required. |
| `--dry-run` | Write results to a separate printed directory; repo unchanged. |
| `--[no-]reflog` | Migrate reflogs (default). `--no-reflog` discards them. |

### git refs verify

| Flag | What it does |
|------|--------------|
| `--strict` | Warnings are errors. |
| `--verbose` | More detail. |

### git refs list

`git refs list [<flags>] [--stdin | <pattern>...]`. Patterns: glob, full name, or prefix ending at `/`.

| Flag | What it does |
|------|--------------|
| `--stdin` | Patterns from stdin. |
| `--count=<count>` | Stop after `<count>` refs. |
| `--sort=<key>` | Sort key; `-` prefix descends. Repeatable; last is primary. Default `refname`. |
| `--format[=<format>]` | `%(fieldname)` template; `%%` = `%`, `%xx` = hex char. Default `%(objectname) %(objecttype)<TAB>%(refname)`. Fields: see [`for-each-ref`](for-each-ref.md). |
| `--color[=<when>]` | Format colors: `always` (default), `never`, `auto`. |
| `--shell`, `--perl`, `--python`, `--tcl` | Quote values for `eval` in that language. |
| `--points-at=<object>` | Only refs pointing at `<object>`. |
| `--merged[=<object>]` | Only refs reachable from `<object>` (default `HEAD`). |
| `--no-merged[=<object>]` | Only refs not reachable from `<object>` (default `HEAD`). |
| `--contains[=<object>]` | Only refs containing `<object>` (default `HEAD`). |
| `--no-contains[=<object>]` | Only refs not containing `<object>` (default `HEAD`). |
| `--ignore-case` | Case-insensitive sort and filter. |
| `--omit-empty` | No newline for empty output. |
| `--exclude=<pattern>` | Hide matching refs. Repeatable. |
| `--include-root-refs` | Also `HEAD` and pseudorefs (`MERGE_HEAD`, ...). |
| `--start-after=<marker>` | Start after this ref, for paging. Not with `--sort`, `--stdin`, or patterns. |

### git refs exists

`git refs exists <ref>`. No flags.

### git refs optimize

| Flag | What it does |
|------|--------------|
| `--all` | Pack all refs (except hidden, broken, symbolic). Default: tags and already-packed refs. |
| `--no-prune` | Keep loose ref files. |
| `--auto` | Pack only if needed. |
| `--include <pattern>` | Pack matching refs. Repeatable. Drops the default tags. No effect with `--all`. `--no-include` clears. |
| `--exclude <pattern>` | Don't pack matching refs (beats `--include`). Repeatable. `--no-exclude` clears. |

### git refs create

(2.56+) `git refs create [<flags>] <ref> <new-value>`

| Flag | What it does |
|------|--------------|
| `--message=<reason>` | Reflog message, non-empty. |
| `--no-deref` | Act on a symbolic ref itself, not its target. |
| `--create-reflog` | Create a reflog even where not normally kept. |

### git refs update

(2.56+) `git refs update [<flags>] <ref> <new-value> [<old-value>]`. `<old-value>`: only update if it matches. All-zero `<new-value>` deletes; all-zero `<old-value>` means must not exist.

| Flag | What it does |
|------|--------------|
| `--message=<reason>` | Reflog message, non-empty. |
| `--no-deref` | Act on a symbolic ref itself, not its target. |
| `--create-reflog` | Create a reflog even where not normally kept. |

### git refs delete

(2.56+) `git refs delete [<flags>] <ref> [<old-value>]`. `<old-value>`: only delete if it matches.

| Flag | What it does |
|------|--------------|
| `--message=<reason>` | Reflog message, non-empty. |
| `--no-deref` | Delete a symbolic ref itself, not its target. |

### git refs rename

(2.56+) `git refs rename [<flags>] <old-ref> <new-ref>`. Both names must be valid.

| Flag | What it does |
|------|--------------|
| `--message=<reason>` | Reflog message, non-empty. |
