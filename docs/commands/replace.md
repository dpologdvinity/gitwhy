# git replace

Make git use one object in place of another, without rewriting history. Stored as ref `refs/replace/<old-id>`. Used to fix a bad commit or graft histories together while keeping commit IDs.

Most commands follow replacements; `prune`, pack push/fetch, and `fsck` don't. Shared only if you push/fetch `refs/replace/*`. See the original with `git --no-replace-objects ...` or `GIT_NO_REPLACE_OBJECTS`. To make it permanent, rewrite history with [`git filter-repo`](filter-repo.md).

Both objects must be the same type unless `-f` (merge and non-merge commits may swap). Replacements disable the commit-graph, which can slow things. Quirks: diffing an object against its replacement misbehaves; `git reset --hard <replaced>` lands on the replacement.

**Usage:** `git replace [-f] <object> <replacement>`
or `git replace [-f] --edit <object>`
or `git replace [-f] --graft <commit> [<parent>...]`
or `git replace [-f] --convert-graft-file`
or `git replace -d <object>...`
or `git replace [--format=<format>] [-l [<pattern>]]`

## Flags

| Flag | What it does |
|------|--------------|
| `<object> <replacement>` | Use `<replacement>` wherever `<object>` is asked for. |
| `-f`, `--[no-]force` | Overwrite an existing replacement. Also allows different types. |
| `-d`, `--delete` | Remove replacements for these objects. |
| `-e`, `--edit <object>` | Edit the object in your editor; replace it with the result. |
| `--[no-]raw` | With `--edit`: edit raw contents. Only affects trees; for trees too broken to pretty-print. |
| `-g`, `--graft <commit> [<parent>...]` | Replace the commit with a copy that has these parents (none = root). |
| `--convert-graft-file` | Turn `.git/info/grafts` entries into `--graft` replacements, then delete the file. |
| `-l`, `--list [<pattern>]` | List replaced objects, optionally matching `<pattern>`. Default action. |
| `--[no-]format=<format>` | List format: `short` (default, `<old>`), `medium` (`<old> -> <new>`), `long` (adds types). |
