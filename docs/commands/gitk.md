# gitk

GUI history browser: commit graph, commit details and changes, files at each commit. Run as `gitk`, not `git gitk`. Settings in `$XDG_CONFIG_HOME/git/gitk` (usually `~/.config/git/gitk`) or `~/.gitk`.

Often a separate package (`gitk` on Debian/Ubuntu). Needs Tcl/Tk.

**Usage:** `gitk [<options>] [<revision-range>] [--] [<path>...]`

## Flags

Also takes most [revision options](../shared/revision-options.md) and some [diff options](../shared/diff-options.md). Join values to flags (`--since=2.weeks`, not `--since 2.weeks`).

| Flag | What it does |
|------|--------------|
| `--all` | Show all refs. |
| `--branches[=<pattern>]` | Show local branches, optionally matching a glob. No `?`, `*`, `[` means `/*` is appended. |
| `--tags[=<pattern>]` | Same, for tags. |
| `--remotes[=<pattern>]` | Same, for remote-tracking branches. |
| `--since=<date>` | Only commits newer than this. |
| `--until=<date>` | Only commits older than this. |
| `--date-order` | Sort by date where possible. |
| `--merge` | During a stopped merge: commits on either side touching conflicted files. |
| `--left-right` | In `A...B`, mark commits `<` (A side) or `>` (B side). |
| `--full-history` | With paths: don't prune history. |
| `--simplify-merges` | With `--full-history`: drop merges that add nothing. |
| `--ancestry-path` | With `A..B`: only commits descended from A and ancestors of B. |
| `-L<start>,<end>:<file>`, `-L:<funcname>:<file>` | Follow history of a line range or function (regex) in one file. Ends: line number, `/regex/`, or (end only) `+N`/`-N`. Repeatable. At most one starting revision; no path limits. |
| `<revision-range>` | One commit (and its history) or `<from>..<to>`. |
| `<path>...` | Only commits touching these paths. Put `--` before them to avoid confusion with branches. |
| `--argscmd=<command>` | Run on every refresh; prints extra revisions, one per line. |
| `--select-commit=<ref>` | Select this commit after loading. Default `HEAD`. |
