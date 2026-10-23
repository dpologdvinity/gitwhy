# git add

Stage files' current content in the index for the next commit. Edit again and you must re-add. Ignored files skipped unless `-f`.

**Usage:** `git add [<flags>] [--] [<pathspec>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<pathspec>...` | Files to stage. Quote globs (`'*.c'`) so git matches subfolders too. A directory stages all inside, deletions included. |
| `-n`, `--[no-]dry-run` | Show what would be staged or ignored. |
| `-v`, `--[no-]verbose` | Print each added file. |
| `-f`, `--[no-]force` | Stage ignored files. Also needed for a submodule with `submodule.<name>.ignore=all` (name its path). |
| `--[no-]sparse` | Allow files outside the sparse-checkout area. |
| `-i`, `--[no-]interactive` | Menu to pick what to stage. See [Interactive mode](#interactive-mode). |
| `-p`, `--[no-]patch` | Pick changes hunk by hunk. |
| `--[no-]auto-advance` | With `-p`: go to the next file after its last hunk (default). `--no-auto-advance`: stay; `<`/`>` switch files. |
| `-U<n>`, `--unified=<n>` | With `-p`: `<n>` context lines (default `diff.context` or 3). |
| `--inter-hunk-context=<n>` | With `-p`: merge hunks up to `<n>` lines apart (default `diff.interHunkContext` or 0). |
| `-e`, `--[no-]edit` | Edit the diff in your editor; stage the saved patch (empty = nothing). See [Editing patches](#editing-patches). |
| `-u`, `--[no-]update` | Stage changes and deletions of tracked files only; no new files. No pathspec = whole repo. |
| `-A`, `--[no-]all`, `--no-ignore-removal` | Stage new, changed, and deleted files. No pathspec = whole repo. |
| `--no-all`, `--ignore-removal` | Stage new and changed files, not deletions. Only matters with a pathspec. |
| `-N`, `--[no-]intent-to-add` | Record a new file with no content yet, so `git diff` shows it and `git commit -a` includes it. |
| `--[no-]refresh` | Don't stage; only refresh cached stat info. |
| `--[no-]ignore-errors` | Skip files that fail; still exit non-zero. Default from `add.ignoreErrors`. |
| `--[no-]ignore-missing` | With `--dry-run`: check if nonexistent files would be ignored. |
| `--no-warn-embedded-repo` | No warning when adding a nested repo without `git submodule add`. |
| `--[no-]renormalize` | Re-stage all tracked files through line-ending/filter rules (after changing `core.autocrlf` or `text`). Implies `-u`. |
| `--resolved` | (2.56+) Stage conflicted files with no markers left (incl. deletions, binaries); leave other changes. Stages nothing if any match still has markers. Not with `-u` or `-A`. |
| `--[no-]chmod=(+\|-)x` | Set/clear the executable bit in the index only. |
| `--[no-]pathspec-from-file=<file>` | Read pathspecs from `<file>` (`-` = stdin), one per line. |
| `--[no-]pathspec-file-nul` | With `--pathspec-from-file`: NUL-separated, literal. |
| `--` | Everything after is a file name. |

## Interactive mode

`git add -i` shows status and a `What now>` menu. Type a number or unique prefix. At `>>` prompts pick several: `2-5 7,9`, `7-` (to end), `*` (all), `-2` (unpick). Empty line finishes.

| Command | What it does |
|---------|--------------|
| `status` | Staged and unstaged line counts per file. |
| `update` | Stage picked files. |
| `revert` | Unstage picked files (new files become untracked). |
| `add untracked` | Stage picked untracked files. |
| `patch` | Stage a file hunk by hunk (like `-p`). |
| `diff` | Show staged vs `HEAD`. |
| `quit`, `help` | Exit; show help. |

Hunk keys (`interactive.singleKey` skips Enter):

| Key | What it does |
|-----|--------------|
| `y` / `n` | Stage / skip hunk. |
| `q` | Quit; skip remaining hunks. |
| `a` / `d` | Stage / skip this and later hunks in the file. |
| `g` | Go to hunk by number. |
| `/` | Go to hunk matching regex. |
| `j` / `J` | Next undecided / next hunk (wraps). |
| `k` / `K` | Previous undecided / previous hunk (wraps). |
| `<` / `>` | Previous / next file (with `--no-auto-advance`). |
| `s` | Split hunk. |
| `e` | Edit hunk by hand. |
| `p` / `P` | Print hunk / in pager. |
| `?` | Help. |

## Editing patches

With `-e` or `e`, you edit a diff applied to the index only.

- Skip an added (`+`) line: delete it.
- Skip a removal (`-`): change `-` to a space.
- Skip a change: turn its `-` lines into spaces, delete its `+` lines.
- Don't alter context (space) or `-` lines' text, or the patch won't apply.

Staged content not in your working file will look "undone" in the working tree.
