# git apply

Apply a patch to files, optionally the index. Doesn't commit (use `git am` for emailed patches). Works outside repos, like `patch`. All-or-nothing by default. From a subfolder, skips paths outside it.

**Usage:** `git apply [<flags>] [<patch>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<patch>...` | Patch files. `-` or none: stdin. |
| `--[no-]stat` | Show diffstat; don't apply. |
| `--[no-]numstat` | Added/deleted counts and full path, machine-readable (binary: `-` `-`). Don't apply. |
| `--[no-]summary` | Summarize created, deleted, renamed, mode-changed files; don't apply. |
| `--[no-]check` | Test if it applies cleanly; don't apply. |
| `--[no-]apply` | Apply anyway with `--stat`, `--numstat`, `--summary`, `--check`. Put it after them. |
| `--[no-]index` | Apply to files and index; they must match for touched paths. |
| `--[no-]cached` | Apply to index only. |
| `-N`, `--[no-]intent-to-add` | Files-only mode: mark new files intent-to-add (`git add -N`). Ignored with `--index`/`--cached`. |
| `-3`, `--[no-]3way` | 3-way merge using the patch's original versions (if present; else apply normally); may leave conflict markers. Implies `--index` unless `--cached`. Not with `--reject`. |
| `--ours`, `--theirs`, `--union` | With `--3way`: resolve conflicts with ours, theirs, or both. |
| `--[no-]build-fake-ancestor=<file>` | Write a temp index of the patch's original versions to `<file>`. Used by `git am`. |
| `-R`, `--[no-]reverse` | Apply in reverse (undo). |
| `--[no-]reject` | Apply what fits; save failed hunks as `*.rej`. |
| `-z` | With `--numstat`: NUL-terminated, unquoted. |
| `-p<n>` | Strip `<n>` leading path parts. Default 1 (`a/`, `b/`). |
| `-C<n>` | Require only `<n>` context lines to match. Default: all. |
| `--[no-]unidiff-zero` | Accept zero-context patches (`--unified=0`). Less safe. |
| `--no-add` | Apply only deletions. |
| `--add` | Apply additions too. Cancels `--no-add`. |
| `--allow-binary-replacement`, `--binary` | No-op; binary patches always allowed. |
| `--exclude=<path-pattern>` | Skip matching files. Repeatable. |
| `--include=<path-pattern>` | Apply only to matching files. Repeatable. With both, first matching pattern wins; unmatched skipped if any `--include`. |
| `--[no-]ignore-space-change`, `--[no-]ignore-whitespace` | Ignore context whitespace when matching. Default from `apply.ignoreWhitespace`. |
| `--[no-]whitespace=<action>` | Whitespace errors (per `core.whitespace`) in added lines: `nowarn`, `warn` (default), `fix`/`strip`, `error`, `error-all` (list all). Default from `apply.whitespace`. |
| `--[no-]inaccurate-eof` | Work around tools mis-recording missing final newlines. |
| `--[no-]recount` | Recount hunk header line counts. For hand-edited patches. |
| `--[no-]directory=<root>` | Prefix paths with `<root>/` (after `-p`). |
| `--[no-]unsafe-paths` | Allow paths outside the working tree. No effect with `--index`/`--cached`. |
| `--[no-]allow-empty` | Don't fail on empty patches. |
| `--[no-]allow-overlap` | Allow overlapping hunks. |
| `-v`, `--[no-]verbose` | More stderr progress. |
| `-q`, `--[no-]quiet` | No stderr messages. |
