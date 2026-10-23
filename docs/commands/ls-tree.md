# git ls-tree

List a tree's entries (mode, type, ID, name), like `ls` for a past version. `git ls-tree -r HEAD` lists every file in the last commit.

Paths are patterns relative to the current folder. Don't mix a subfolder tree with a relative path (`git ls-tree HEAD:sub dir` looks for `sub/sub/dir`); use `--full-tree`.

Default line: `<mode> <type> <id> TAB <path>` (valid `git update-index --index-info` input).

**Usage:** `git ls-tree [<flags>] <tree-ish> [<path>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<tree-ish>` | Commit, tag, or tree, e.g. `HEAD`, `main:src`. |
| `[<path>...]` | Only matching entries. Default: current folder's entries (top level at repo root). |
| `-d` | Only folder entries, not their contents. Implies `-t`. |
| `-r` | Recurse into subfolders. |
| `-t` | With `-r`, also list folders. |
| `-l`, `--long` | Show file size in bytes (`-` for folders, submodules). |
| `-z` | NUL-terminated, unquoted names. |
| `--name-only`, `--name-status` | Paths only (both identical). |
| `--object-only` | Object IDs only; fast `--format='%(objectname)'`. |
| `--[no-]abbrev[=<n>]` | Shorten IDs to at least `<n>` unique hex digits. |
| `--[no-]full-name` | Paths from repo root. |
| `--[no-]full-tree` | Whole tree, ignoring current folder. Implies `--full-name`. |
| `--format=<format>` | Custom line: `%(objectmode)`, `%(objecttype)`, `%(objectname)`, `%(objectsize)` or `%(objectsize:padded)`, `%(path)`, `%%`, `%xNN` (hex byte, `%x09` = tab). Not with `-l`, `--name-only`, `--object-only`. |
