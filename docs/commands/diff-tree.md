# git diff-tree

Plumbing: compare two trees or commits; one commit is compared with its parents. Default output is `--raw`, top level only unless `-r`. For humans: `git diff` or `git show`.

**Usage:** `git diff-tree [--stdin] [-m] [-s] [-v] [--no-commit-id] [--pretty] [-t] [-r] [-c | --cc] [--combined-all-paths] [--root] [--merge-base] [<flags>] <tree-ish> [<tree-ish>] [<path>...]`

## Flags

Also takes all [diff options](../shared/diff-options.md).

| Flag | What it does |
|------|--------------|
| `-r` | Recurse into subfolders. |
| `-t` | Also list folder entries. Implies `-r`. |
| `--root` | Show a root commit as adding every file. |
| `--merge-base` | Use the two commits' merge base as the old side. Needs exactly two. |
| `--stdin` | Read argument sets from stdin, one per line: two trees, a commit, or a commit plus its parents to use. Prints each ID before its diff. |
| `-m` | Include merges, diffed against each parent. |
| `-s` | With `--stdin`: hide diffs. |
| `-v` | With `--stdin`: show commit messages. |
| `--no-commit-id` | Omit the commit ID line. |
| `--always` | Show the commit even if its diff is empty. |
| `-c` | Merges: combined diff of all parents, only files changed from every parent. Needs one commit or `--stdin`. |
| `--cc` | `-c` plus `-p`, hiding hunks taken unchanged from one parent. |
| `--combined-all-paths` | With `-c`/`--cc`: show each parent's file name (for renames). |
| `--pretty[=<format>]`, `--format=<format>` | Message format: `oneline`, `short`, `medium` (default), `full`, `fuller`, `reference`, `email`, `raw`, `format:<string>`, `tformat:<string>`. A bare `%` string acts as `tformat:`. |
| `--[no-]abbrev-commit` | Short commit IDs. |
| `--oneline` | `--pretty=oneline --abbrev-commit`. |
| `--encoding=<encoding>` | Re-encode messages. |
| `--[no-]expand-tabs[=<n>]` | Tabs to spaces every `<n>` columns (default 8). |
| `--notes[=<ref>]` | Show notes from `<ref>` or the defaults. Repeatable. |
| `--no-notes` | No notes; clears earlier `--notes`. |
| `--show-notes-by-default` | Show default notes unless specific ones were asked for. |
| `--show-notes[=<ref>]`, `--[no-]standard-notes` | Old spellings of `--notes`/`--no-notes`. |
| `--show-signature` | Verify and show commit signatures. |
| `<tree-ish>` | Tree or commit to compare. |
| `<path>...` | Only these paths. |
