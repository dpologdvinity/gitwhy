# git filter-repo

Rewrite all history fast: remove files or secrets, move into a subfolder, fix authors, rename tags, and more. Recommended over [`git filter-branch`](filter-branch.md). Separate install (`git-filter-repo`; Python 3, git 2.36+).

**Destroys old history:** changed commits get new IDs, so others must re-clone. Refuses to run outside a fresh clone unless `--force`. Afterward by default: turns `refs/remotes/origin/*` into local branches, removes `origin` (prevents mixed pushes), deletes unexported refs, expires reflogs, runs `git gc`. `--partial` skips all that.

Renames aren't followed: give both paths (`--path old/ --path new/`). Run `--analyze` first.

**Usage:** `git filter-repo --analyze` or `git filter-repo [<flags>]`

Examples:

- Keep only two folders: `git filter-repo --path guides/ --path tools/releases`
- Delete a file and a folder everywhere: `git filter-repo --path foo.zip --path bar/zips/ --invert-paths`
- Replace a password in all files: `git filter-repo --replace-text <(echo 'hunter2==>REDACTED')`
- Make `src/` the new root: `git filter-repo --subdirectory-filter src`
- Move all into `my-module/` and prefix tags: `git filter-repo --to-subdirectory-filter my-module --tag-rename '':'my-module-'`

## Flags

### Analysis

| Flag | What it does |
|------|--------------|
| `--analyze` | Change nothing; write reports on renames and sizes by path, folder, extension, blob. |
| `--report-dir <dir>` | `--analyze` output folder. Default `.git/filter-repo/analysis`. Must not exist, unless `--force` (deletes it). |

### Choosing paths

Multiple path flags mean "any of these".

| Flag | What it does |
|------|--------------|
| `--path <dir-or-file>`, `--path-match <dir-or-file>` | Keep only this file or folder (folders end in `/`). Repeatable. |
| `--path-glob <glob>` | Keep paths matching this glob, e.g. `'*.txt'`. Repeatable. |
| `--path-regex <regex>` | Keep paths matching this regex. Repeatable. |
| `--invert-paths` | Remove matching paths instead; keep the rest. |
| `--use-base-name` | Match file names only. Can't match folders; not with `--path-rename`. |

### Renaming paths

| Flag | What it does |
|------|--------------|
| `--path-rename <old>:<new>`, `--path-rename-match <old>:<new>` | Rename a path prefix throughout history. Repeatable. Doesn't select; with path filters, select it too. |

### Path shortcuts

| Flag | What it does |
|------|--------------|
| `--paths-from-file <file>` | Path filters and renames from a file, one per line: `literal:` (default), `glob:`, `regex:`; `<old>==><new>` renames. `#` and blank lines ignored. |
| `--subdirectory-filter <dir>` | Keep only `<dir>` as the new root. Same as `--path <dir>/ --path-rename <dir>/:`. |
| `--to-subdirectory-filter <dir>` | Move everything into `<dir>/`. Same as `--path-rename :<dir>/`. |

### Changing file contents

| Flag | What it does |
|------|--------------|
| `--replace-text <file>` | Replace text in all file versions. One per line: literal, `regex:...`, or `glob:...`, then `==><replacement>` (default `***REMOVED***`). |
| `--strip-blobs-bigger-than <size>` | Remove file versions over `<size>` (e.g. `5M`, `2G`). |
| `--strip-blobs-with-ids <file>` | Remove blobs whose IDs are listed in `<file>`. |

### Renaming refs

| Flag | What it does |
|------|--------------|
| `--tag-rename <old>:<new>` | Change tag prefix `<old>` to `<new>` (`foo:bar` makes `foo-1.2` `bar-1.2`). Either may be empty. |

### Commit messages

| Flag | What it does |
|------|--------------|
| `--replace-message <file>` | Replace text in commit and tag messages. Format as `--replace-text`. |
| `--preserve-commit-hashes` | Don't rewrite old commit IDs mentioned in messages. |
| `--preserve-commit-encoding` | Don't re-encode messages to UTF-8. |

### Names and emails

| Flag | What it does |
|------|--------------|
| `--mailmap <file>` | Rewrite names/emails with this mailmap (format: `git shortlog`). |
| `--use-mailmap` | Same as `--mailmap .mailmap`. |

### Parents and empty commits

| Flag | What it does |
|------|--------------|
| `--replace-refs <mode>` | Replace refs (see [`git replace`](replace.md)) mapping old IDs to new: `delete-no-add`, `delete-and-add`, `update-no-add` (default), `update-or-add` (add where missing), `update-and-add`, `old-default` (pre-2.45 behavior). |
| `--prune-empty <when>` | Drop commits with no changes: `auto` (default, only newly empty), `always`, `never`. Children reattach to the nearest kept ancestor. |
| `--prune-degenerate <when>` | Drop merges that became pointless and change no files: `auto` (default), `always`, `never`. |
| `--no-ff` | With `--prune-degenerate`: keep merges whose first parent is an ancestor of another. For `merge --no-ff` projects. |

### Python callbacks

Each takes a Python function body (or file). Wrapped as `def <name>_callback(<name>): BODY`, e.g. `--name-callback 'return name.replace(b"Jon", b"John")'`. Values are bytes. Fields: [callbacks docs](https://htmlpreview.github.io/?https://github.com/newren/git-filter-repo/blob/docs/html/git-filter-repo.html#CALLBACKS).

| Flag | What it does |
|------|--------------|
| `--filename-callback <code>` | Change each path (`filename`); `None` drops it. |
| `--file-info-callback <code>` | Gets `filename, mode, blob_id, value`; returns `(filename, mode, blob_id)`. `value` has `get_contents_by_identifier`, `get_size_by_identifier`, `insert_file_with_contents`, `is_binary`, `apply_replace_text`, `data`. Not called for deletions. |
| `--message-callback <code>` | Change messages (`message`). |
| `--name-callback <code>` | Change person names (`name`). |
| `--email-callback <code>` | Change emails (`email`). |
| `--refname-callback <code>` | Change ref names (`refname`). |
| `--blob-callback <code>` | Change file contents (`blob.data`). |
| `--commit-callback <code>` | Change commits (`commit`). |
| `--tag-callback <code>` | Change annotated tags (`tag`). Lightweight tags aren't seen. |
| `--reset-callback <code>` | Change stream `reset` items (`reset`). |

### Removing sensitive data

| Flag | What it does |
|------|--------------|
| `--sensitive-data-removal`, `--sdr` | Secret removal mode: first mirror-fetches all `origin` refs (may discard local-only changes), reports changed commits and orphaned LFS objects, prints server cleanup steps. |
| `--no-fetch` | With `--sdr`: skip the fetch. Other refs may still hold the secret. Implied by `--partial`. |

### Source and target

These imply `--partial` (keeping the normal `--replace-refs` default).

| Flag | What it does |
|------|--------------|
| `--source <repo>` | Read history from this repo. |
| `--target <repo>` | Write filtered history here (overwrites). |

### Commit order

| Flag | What it does |
|------|--------------|
| `--date-order` | Process commits by timestamp. |

### Other

| Flag | What it does |
|------|--------------|
| `-h`, `--help` | Help. |
| `--version` | Print version. |
| `--proceed` | Run with no filter flags. |
| `-f`, `--force` | Run outside a fresh clone. Irreversible: prunes reflogs and old objects immediately. |
| `--partial` | Leave old and new history mixed: keep `origin`, its refs, unexported refs, reflogs; no `gc`. `--tag-rename`/`--refname-callback` then add refs, keeping old ones. |
| `--no-gc` | Skip `git gc`. |
| `--refs <ref>...` | Rewrite only these refs. Implies `--partial`. |
| `--dry-run` | Change nothing; save original and filtered streams in `.git/filter-repo/` to compare. |
| `--debug` | Print extra detail. |
| `--stdin` | Filter a `git fast-export` stream from stdin (needs `original-oid` lines). |
| `--quiet` | Pass `--quiet` to git commands it runs. |
