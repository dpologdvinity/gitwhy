# git status

Show staged changes, unstaged changes, and untracked files, plus the branch, ahead/behind counts, and any merge/rebase in progress. Paths are relative to the current directory (`status.relativePaths`).

**Usage:** `git status [<flags>] [--] [<pathspec>...]`

Rewrites the index to cache file info; in background scripts use `git --no-optional-locks status`. Slow in huge repos due to untracked scan; `-uno` is fastest.

## Flags

| Flag | What it does |
|------|--------------|
| `-s`, `--[no-]short` | One `XY <path>` line per file (see below). |
| `-b`, `--[no-]branch` | Show branch and upstream in short format (`## <branch>...<upstream> [ahead N, behind M]`). |
| `--[no-]show-stash` | Show stash entry count. |
| `--[no-]porcelain[=<version>]` | Stable script output, ignores config. `v1` (default) or `v2` (more detail). |
| `--[no-]long` | Human format (default). |
| `-v`, `--[no-]verbose` | Also diff staged changes. Twice: also unstaged. |
| `-u`, `--[no-]untracked-files[=<mode>]` | `no` (hide), `normal` (files and folders; default), `all` (every file in untracked folders; default for bare `-u`). Attach the mode: `-uno`. Config `status.showUntrackedFiles`. |
| `--[no-]ignored[=<mode>]` | Show ignored files. `traditional` (default): files and folders (files with `-uall`). `matching`: only paths matching a pattern themselves. `no`: none. |
| `--[no-]ignore-submodules[=<when>]` | Hide submodule changes. `all` (default): all. `dirty`: show only commit changes. `untracked`: ignore their untracked files. `none`: hide nothing, overriding config. |
| `-z`, `--[no-]null` | NUL-terminate entries. Implies `--porcelain=v1` if no format. |
| `--[no-]column[=<style>]` | Untracked files in columns. Bare = `always`. Style syntax: `column.status`. |
| `--[no-]ahead-behind` | Exact ahead/behind counts (default). `--no-...` is faster on big histories. |
| `--[no-]renames` | Rename detection on/off, ignoring config. |
| `-M`, `--find-renames[=<n>]` | Detect renames, optional similarity threshold `<n>`. |

## Short format codes

Line: `XY <path>`, or `XY <from> -> <to>` for renames. `X` = staged (index vs `HEAD`), `Y` = unstaged (worktree vs index). In a conflict, `X` and `Y` are the two sides.

| Code | Meaning |
|------|---------|
| ` ` | Unchanged. |
| `M` | Modified. |
| `T` | Type changed (file, symlink, submodule). |
| `A` | Added. |
| `D` | Deleted. |
| `R` | Renamed. |
| `C` | Copied (only if `status.renames=copies`). |
| `U` | Unmerged. |
| `??` | Untracked. |
| `!!` | Ignored (with `--ignored`). |

Examples: `M ` staged; ` M` unstaged; `MM` staged then changed again; `A ` new file staged.

Conflicts: `DD` both deleted, `AU` added by us, `UD` deleted by them, `UA` added by them, `DU` deleted by us, `AA` both added, `UU` both modified.

Submodules in `Y`: `M` new commit, `m` modified content, `?` untracked files.

## Porcelain formats

**v1:** short format with no color, repo-root paths, submodules only `M`. With `-z`: no quoting, NUL after each path, renames as `<to>` NUL `<from>`.

**v2:** headers start with `#` (ignore unknown ones):
- `# branch.oid <commit>` (or `(initial)`), `# branch.head <branch>` (or `(detached)`), `# branch.upstream <upstream>`, `# branch.ab +<ahead> -<behind>`: with `--branch`.
- `# stash <N>`: with `--show-stash`, if any.

Entry lines (any order), then untracked and ignored:

| Line | Meaning |
|------|---------|
| `1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>` | Changed file. |
| `2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>` | Renamed/copied. `<X><score>` like `R100`. `<sep>` is TAB (NUL with `-z`). |
| `u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>` | Unmerged, with stage 1/2/3 modes and IDs. |
| `? <path>` | Untracked. |
| `! <path>` | Ignored. |

Fields: `<XY>` as short format, `.` = unchanged. `<sub>`: `N...` if not a submodule, else `S<c><m><u>` (`C` commit changed, `M` tracked changes, `U` untracked; `.` if not). `<mH>`/`<mI>`/`<mW>`: octal mode in `HEAD`/index/worktree. `<hH>`/`<hI>`: object ID in `HEAD`/index. Without `-z`, odd paths are quoted (`core.quotePath`).
