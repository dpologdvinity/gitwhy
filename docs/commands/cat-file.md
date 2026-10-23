# git cat-file

Plumbing: print an object's contents, type, or size. One object as an argument, or many on stdin with a `--batch` flag.

**Usage:** `git cat-file <type> <object>`
or `git cat-file (-e | -p | -t | -s) <object>`
or `git cat-file (--textconv | --filters) [<rev>:<path|tree-ish> | --path=<path|tree-ish> <rev>]`
or `git cat-file (--batch | --batch-check | --batch-command) [--batch-all-objects] [--buffer] [--follow-symlinks] [--unordered] [--textconv | --filters] [-Z]`

## Flags

| Flag | What it does |
|------|--------------|
| `<object>` | ID, branch, `HEAD:path`, etc. |
| `<type>` | Print raw contents as this type (`blob`, `tree`, `commit`, `tag`). Can peel, e.g. `tree` of a commit. |
| `-p` | Pretty-print contents. |
| `-t` | Print type. |
| `-s` | Print size in bytes (after mailmap with `--use-mailmap`). |
| `-e` | Exit 0 if the object exists and is valid; print nothing. |
| `--[no-]use-mailmap`, `--[no-]mailmap` | Rewrite identities via `.mailmap`. |
| `--textconv` | Content after its `textconv` filter. Object must be `<tree-ish>:<path>` or `:<path>`. |
| `--filters` | Content as checked out (smudge, line endings). Object must be `<tree-ish>:<path>` or `:<path>`. |
| `--[no-]path=<path>` | Path for `--textconv`/`--filters`, given apart from the object. Not with batch modes. |
| `--batch[=<format>]` | For each stdin object: info line, then contents. Combines only with `--textconv`, `--filters`, `--use-mailmap` and the batch helpers below. |
| `--batch-check[=<format>]` | Like `--batch`, info line only. |
| `--batch-command[=<format>]` | Read stdin commands: `contents <object>`, `info <object>`, `remote-object-info <remote> <object>...` (2.56+: info from a remote without downloading; needs server `object-info`; only `%(objectname)`, `%(objectsize)`, `%(objecttype)` filled), `flush` (with `--buffer`), `mailmap <bool>`. |
| `--batch-all-objects` | With `--batch`/`--batch-check`: every object in repo and alternates, not stdin. Sorted by ID; ignores `git replace`. |
| `--[no-]buffer` | Buffer batch output. Much faster. |
| `--[no-]unordered` | With `--batch-all-objects`: fastest order. Each still once. |
| `--[no-]follow-symlinks` | Batch with `<tree-ish>:<path>`: report symlink targets. Out-of-tree links print `symlink <size>` and the path; also `dangling`, `loop`, `notdir`. Not for index paths. |
| `--[no-]filter=<filter-spec>` | Batch only: omit objects. `blob:none`, `blob:limit=<n>[kmg]` (blobs that size or bigger), `object:type=(tag\|commit\|tree\|blob)` (all other types). Omitted ones print `<object> excluded`. |
| `-Z` | Batch: NUL-separated input and output. |
| `-z` | Batch: NUL-separated input only. Deprecated; use `-Z`. |

Batch `<format>` atoms: `objectname`, `objecttype`, `objectmode`, `objectsize`, `objectsize:disk`, `deltabase` (delta base, else zeros), `rest` (input after first space). Default `%(objectname) %(objecttype) %(objectsize)`. Unknown: `<object> missing`; ambiguous: `<object> ambiguous`; submodule commits: `<oid> submodule`.
