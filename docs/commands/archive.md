# git archive

Write one commit's or tree's tracked files as a tar or zip (stdout by default). From a commit or tag: file times are the commit time and the commit ID is embedded (read with `git get-tar-commit-id`). From a tree: times are now.

`export-ignore` files are left out; `export-subst` files get placeholders expanded (see `gitattributes`).

**Usage:** `git archive [<flags>] <tree-ish> [<path>...]`
or `git archive --remote=<repo> [--exec=<cmd>] [<flags>] <tree-ish> [<path>...]`
or `git archive [--remote=<repo>] --list`

`<path>...`: only these paths. Default: everything under the current folder.

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]format=<fmt>` | `tar`, `zip`, `tar.gz`, `tgz`, or a `tar.<format>.command` format. Default: from `-o`'s extension, else `tar`. |
| `-l`, `--[no-]list` | List formats. |
| `-v`, `--[no-]verbose` | Print each file to stderr. |
| `--[no-]prefix=<prefix>/` | Prepend `<prefix>/` to paths. Repeatable: last applies to tracked files; each `--add-file` uses the one before it. Not sanitized (`../`, leading `/` kept). |
| `-o`, `--[no-]output=<file>` | Write to `<file>`. |
| `--[no-]add-file=<file>` | Add an untracked file at the last `--prefix` plus its base name. Repeatable. |
| `--[no-]add-virtual-file=<path>:<content>` | Add a file with this content. Ignores `--prefix`. C-quote `<path>` if it has `:` or starts/ends with `"`. Repeatable. |
| `--[no-]worktree-attributes` | Also read working-tree `.gitattributes`. |
| `--mtime=<time>` | File time to record. |
| `-<N>` | Compression level. zip: `-0` (store) to `-9`, default `-6`. tar: passed to the compressor. |
| `--[no-]remote=<repo>` | Archive from a remote repo. It may only allow ref names (see [`upload-archive`](upload-archive.md)). |
| `--[no-]exec=<cmd>` | With `--remote`: path to remote `git-upload-archive`. |
