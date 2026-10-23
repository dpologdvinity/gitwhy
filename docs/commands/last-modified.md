# git last-modified

Experimental: show the commit that last changed each file and folder (like GitHub's file listing). Renames and mode changes count.

Output: `<commit-id> TAB <path>` per line. Unusual paths C-quoted unless `-z`.

**Usage:** `git last-modified [-r] [-t] [--max-depth=<depth>] [-z] [<revision-range>] [[--] <pathspec>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<revision-range>` | Only search these commits (e.g. `v1.0..main`). Default: history up to `HEAD`. |
| `[--] <pathspec>...` | Paths to report. Default: top level. |
| `-r`, `--[no-]recursive` | List every file in all subfolders. Same as `--max-depth=-1`. |
| `-t`, `--[no-]show-trees` | When recursing, also list folders. |
| `--max-depth=<depth>` | Max folder levels below each pathspec. Negative: unlimited. Default `0`. |
| `-z` | NUL-terminate lines; no quoting. |
