# git fast-import

Build history from a text stream on stdin: commits, files, tags, branches. Writes packs directly and updates refs at the end. Back end for importers from other VCSs and for `git fast-export | <edit> | git fast-import`. Works on empty or existing repos.

Bad input exits non-zero and writes a crash report (`fast_import_crash_<pid>`) in `.git`. `SIGUSR1` triggers a `checkpoint`.

**Usage:** `<frontend> | git fast-import [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `--force` | Update branches even if commits would be lost. |
| `--quiet` | Hide end stats. Stream `progress` lines still show. |
| `--stats` | Print object, pack, and memory stats at the end (default). |
| `--allow-unsafe-features` | Allow stream `feature` marks commands, which can touch files outside the repo. Trusted streams only. Auto-on for remote helpers. |
| `--signed-commits=<mode>` | Commit signatures: `verbatim` (default, keep), `warn-verbatim`, `abort`, `strip`, `warn-strip`, `strip-if-invalid` (drop bad ones, warn), `sign-if-invalid[=<keyid>]` (re-sign bad ones), `abort-if-invalid`. |
| `--signed-tags=<mode>` | Same choices for tags. Default `verbatim`. |
| `--cat-blob-fd=<fd>` | Write `get-mark`, `cat-blob`, `ls` answers to this fd instead of stdout. |
| `--date-format=<fmt>` | Date format: `raw` (default, `<unix-seconds> <+hhmm>`), `raw-permissive` (no checks), `rfc2822`, `now` (literal `now`). |
| `--done` | Fail unless the stream ends with `done`. Catches a crashed frontend. |
| `--export-marks=<file>` | Write marks (`:<num> <id>`) at each checkpoint and the end. Can be the import file. |
| `--import-marks=<file>` | Load marks first (file must exist). Repeatable; later wins. |
| `--import-marks-if-exists=<file>` | Like `--import-marks`; skip if missing. |
| `--[no-]relative-marks` | Later marks paths are relative to `.git/info/fast-import`. |
| `--rewrite-submodules-from=<name>:<file>` | Old submodule IDs from this `git fast-export` marks file. Pair with `--rewrite-submodules-to` of the same `<name>`. For SHA-1/SHA-256 conversion. Repeatable. |
| `--rewrite-submodules-to=<name>:<file>` | Matching marks file from the submodule's `git fast-import`. |
| `--active-branches=<n>` | Branches held in memory. Default 5. |
| `--big-file-threshold=<n>` | No delta compression above `<n>` bytes. Default `512m`. |
| `--depth=<n>` | Max delta chain depth. Default 50. |
| `--export-pack-edges=<file>` | After each pack, append its name and last commit per branch. Helps imports over 4 GiB. |
| `--max-pack-size=<n>` | Max pack size. Default unlimited (4 GiB cap still applies). |

## Input stream (summary)

Text lines; exact spacing matters. `#` lines are comments (except inside `data`). Full spec: [git-fast-import docs](https://git-scm.com/docs/git-fast-import#_input_format).

| Command | What it does |
|---------|--------------|
| `commit <ref>` | Commit on `<ref>` (e.g. `refs/heads/main`). Then optional `mark :<n>`, `original-oid <id>`, `author ...`; required `committer <name> <<email>> <when>`; optional `gpgsig`, `encoding <enc>`; message `data`; `from <commit>` (parent); `merge <commit>` (more parents); file changes. |
| File changes (in `commit`) | `M <mode> <dataref> <path>` or `M <mode> inline <path>` + `data`: add/change. Modes: `100644` file, `100755` executable, `120000` symlink, `160000` submodule, `040000` folder. `D <path>`: delete. `C <src> <dst>`: copy. `R <src> <dst>`: rename. `deleteall`: empty tree. `N <dataref> <commit>` or `N inline <commit>` + `data`: set a note. |
| `tag <name>` | Annotated tag `refs/tags/<name>`: `from <commit>`, `tagger ...`, `data`. Lightweight: use `reset`. |
| `reset <ref>` | Create or move a branch/lightweight tag to `from <commit>`. |
| `blob` | Store file contents (`data`), usually with a `mark`. |
| `data <count>` | Exactly `<count>` raw bytes follow. Or `data <<<delim>`: until a `<delim>` line. |
| `mark :<n>` | Label the object as `:<n>` (from 1) for later reference. |
| `alias` | `mark :<n>` + `to <commit>`: point a mark at an existing object. |
| `checkpoint` | Finish the pack and write refs, tags, marks now. Slow. |
| `progress <text>` | Print `progress <text>` to stdout. |
| `get-mark :<n>` | Print a mark's object ID. |
| `cat-blob <dataref>` | Print file contents (`git cat-file --batch` format). |
| `ls [<dataref>] <path>` | Print a path's entry (`git ls-tree` format) or `missing <path>`. |
| `feature <name>[=<arg>]` | Require a feature or abort: `date-format`, `export-marks`, `import-marks`, `import-marks-if-exists`, `relative-marks`, `no-relative-marks`, `force`, `get-mark`, `cat-blob`, `ls`, `notes`, `done`. |
| `option <flag>` | Set a flag (no `--`) from the stream. Must come first. Not: `date-format`, `import-marks`, `export-marks`, `cat-blob-fd`, `force`. Command line wins. |
| `done` | End of stream. Required with `--done`. |

`<commit>`/`<dataref>`: branch, mark `:<n>`, full or short ID, or revision expression. `from refs/heads/x^0` continues an existing branch. 40 zeros in `from` deletes the branch.

Reading answers while writing: use a named pipe and drain output before writing more, or it deadlocks.
