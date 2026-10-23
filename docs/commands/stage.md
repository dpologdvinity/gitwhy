# git stage

Alias for `git add`; same flags. See [`git add`](add.md) for interactive mode and editing patches.

**Usage:** `git stage [<flags>] [--] [<pathspec>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<pathspec>...` | Files to stage. Quote globs (`'*.c'`) so git matches subfolders too. A directory stages all inside, deletions included. |
| `-n`, `--[no-]dry-run` | Show what would be staged or ignored. |
| `-v`, `--[no-]verbose` | Print each added file. |
| `-f`, `--[no-]force` | Stage ignored files. Also needed for a submodule with `submodule.<name>.ignore=all` (name its path). |
| `--[no-]sparse` | Allow files outside the sparse-checkout area. |
| `-i`, `--[no-]interactive` | Menu to pick what to stage. See [Interactive mode](add.md#interactive-mode). |
| `-p`, `--[no-]patch` | Pick changes hunk by hunk. |
| `--[no-]auto-advance` | With `-p`: go to the next file after its last hunk (default). `--no-auto-advance`: stay; `<`/`>` switch files. |
| `-U<n>`, `--unified=<n>` | With `-p`: `<n>` context lines (default `diff.context` or 3). |
| `--inter-hunk-context=<n>` | With `-p`: merge hunks up to `<n>` lines apart (default `diff.interHunkContext` or 0). |
| `-e`, `--[no-]edit` | Edit the diff in your editor; stage the saved patch (empty = nothing). See [Editing patches](add.md#editing-patches). |
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
