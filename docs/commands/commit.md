# git commit

Record the index as a new commit (child of `HEAD`) with a message and move the current branch to it. Detached `HEAD`: only `HEAD` moves.

Choose content with `git add`/`git rm` first (changed files too), by listing files (their current content, ignoring other staged changes), `-a` (all tracked changes), or `-p`/`--interactive`. `--dry-run` previews. Undo a fresh mistake with `--amend` or `git reset`.

If commit fails after you wrote a message, it's kept in `.git/COMMIT_EDITMSG` until the next commit.

**Usage:** `git commit [<flags>] [--] [<pathspec>...]`

## Flags

### What goes in

| Flag | What it does |
|------|--------------|
| `<pathspec>...` | Commit only these tracked files' current content (also staging them); other staged changes wait. Not while finishing a merge (except `-i`). |
| `-a`, `--[no-]all` | Stage all tracked changes and deletions first. Not new files. |
| `-p`, `--[no-]patch` | Pick hunks interactively (as `git add -p`, see [Interactive mode](add.md#interactive-mode)). |
| `--[no-]interactive` | Pick with the `git add -i` menu first. |
| `-U<n>`, `--unified=<n>` | With `-p`: context lines (default `diff.context` or 3). Bare `-U` acts like `-p`. |
| `--inter-hunk-context=<n>` | With `-p`: merge hunks up to `<n>` lines apart (default `diff.interHunkContext` or 0). |
| `-i`, `--[no-]include` | Stage listed files on top of what's staged, commit all. Mainly for conflicted merges. |
| `-o`, `--[no-]only` | Commit only listed files (default when listing). With `--amend` and no files: reword only. With `--allow-empty` and no files: empty commit. |
| `--[no-]pathspec-from-file=<file>` | Read pathspecs from `<file>` (`-` = stdin), one per line. Quoting per `core.quotePath`. |
| `--[no-]pathspec-file-nul` | With `--pathspec-from-file`: NUL-separated, literal. |
| `--[no-]amend` | Replace the last commit (same parents and author, its message as starting text) with what's staged now. Works on merges. Rewrites history: don't amend shared commits. ≈ `git reset --soft HEAD^` + `git commit -c ORIG_HEAD`. |
| `--allow-empty` | Allow a commit with no changes. |
| `--allow-empty-message` | Allow an empty message. |
| `--` | Everything after is a file name. |

### Message

| Flag | What it does |
|------|--------------|
| `-m <msg>`, `--[no-]message=<msg>` | Message. Multiple `-m` = paragraphs. Not with `-c`, `-C`, `-F`. |
| `-F <file>`, `--[no-]file=<file>` | Message from `<file>` (`-` = stdin). |
| `-C <commit>`, `--[no-]reuse-message=<commit>` | Reuse `<commit>`'s message and author (with date). |
| `-c <commit>`, `--[no-]reedit-message=<commit>` | Like `-C`, but edit it. |
| `--[no-]fixup=[(amend\|reword):]<commit>` | Commit for `git rebase --autosquash` to fold into `<commit>`. Plain: `fixup! <title>`; changes merge in, message (and `-m`) dropped. `amend:`: `amend! <title>`; its edited message also replaces `<commit>`'s (non-empty unless `--allow-empty-message`). `reword:`: `amend:` with `--only`, message only. Author unchanged. |
| `--[no-]squash=<commit>` | `squash! <title>` commit for `git rebase --autosquash`; messages combined when squashed. Works with `-m`/`-c`/`-C`/`-F`. |
| `--[no-]reset-author` | With `-C`, `-c`, `--amend`, or after a conflicted cherry-pick: you become author, new date. |
| `--[no-]author=<author>` | `"Name <email>"`, or a pattern copying the latest matching commit's author (`git rev-list --all -i --author=<author>`). |
| `--[no-]date=<date>` | Author date. Formats below, plus `yesterday` etc. |
| `-t <file>`, `--[no-]template=<file>` | Editor starts with `<file>` (default `commit.template`). Unedited = abort. Ignored with `-m`/`-F`. |
| `-e`, `--[no-]edit` | Edit even with `-m`, `-F`, `-C`. `--no-edit`: use as is (`git commit --amend --no-edit`). |
| `--[no-]cleanup=<mode>` | `strip`: drop edge blank lines, trailing spaces, `#` comments; squeeze blanks. `whitespace`: `strip` but keep comments. `verbatim`: none. `scissors`: `whitespace`, plus cut from `# ---- >8 ----` when editing. `default`: `strip` if edited, else `whitespace`. Default `commit.cleanup`; comment char `core.commentChar`. |
| `-s`, `--[no-]signoff` | Add your `Signed-off-by:` (meaning is per project, often the DCO). No config to always add it. |
| `--[no-]trailer <token>[(=\|:)<value>]` | Add a trailer, e.g. `--trailer "Helped-by: Name <email>"`. Repeatable. Follows `trailer.*` config (see `git interpret-trailers`). |
| `--[no-]status` | Include `git status` in the editor template (default; `commit.status`). |
| `-v`, `--[no-]verbose` | Show the diff in the editor (not saved). Twice: also unstaged changes. Default `commit.verbose`. |
| `-S[<key-id>]`, `--[no-]gpg-sign[=<key-id>]` | GPG-sign. Key defaults to committer; no space (`-SABC`). `--no-gpg-sign` overrides `commit.gpgSign`. |

### Hooks, output, dry run

| Flag | What it does |
|------|--------------|
| `-n`, `--no-verify` | Skip `pre-commit` and `commit-msg` hooks. |
| `--verify` | Run them (default). |
| `--no-post-rewrite` | Skip `post-rewrite` hook (after `--amend`). |
| `--post-rewrite` | Run it (default). |
| `-q`, `--[no-]quiet` | No summary. |
| `--[no-]dry-run` | Commit nothing; list what would be committed, left out, and untracked. |
| `--[no-]short` | Dry run in `git status --short` format. |
| `--[no-]branch` | Branch and tracking info in short format. |
| `--[no-]porcelain` | Dry run in `git status --porcelain` format. |
| `--[no-]long` | Dry run in long `git status` format. |
| `-z`, `--[no-]null` | Short/porcelain: raw NUL-separated names. Alone: implies `--porcelain`. |
| `--[no-]ahead-behind` | Count exact ahead/behind vs upstream. `--no-ahead-behind` is faster on big repos. |
| `-u[<mode>]`, `--[no-]untracked-files[=<mode>]` | Untracked files in status: `no`, `normal` (files and folders), `all` (bare `-u`; files inside untracked folders). Default `normal` or `status.showUntrackedFiles`. |

## Author and committer

Name and email: `GIT_AUTHOR_NAME`, `GIT_AUTHOR_EMAIL`, `GIT_COMMITTER_NAME`, `GIT_COMMITTER_EMAIL`; else `author.*`/`committer.*`; else `user.name`/`user.email`; else `$EMAIL`; else system user and host. Usually set `user.name` and `user.email`. Unrelated to login (see `credential.username`).

Dates: `GIT_AUTHOR_DATE`, `GIT_COMMITTER_DATE`. Formats: `@<unix-seconds> <+hhmm>` (`@` needed under 100,000,000), RFC 2822 (`Thu, 07 Apr 2005 22:13:13 +0200`), ISO 8601 (`2005-04-07T22:13:13`, space ok for `T`; date also `YYYY.MM.DD`, `MM/DD/YYYY`, `DD.MM.YYYY`).

## Notes

- Message style: title of 50 characters or fewer, blank line, details. Tools use the title (e.g. `git format-patch` subject).
- Editor: `GIT_EDITOR`, `core.editor`, `VISUAL`, then `EDITOR`.
- Messages should be UTF-8; git warns otherwise. Legacy encoding: `i18n.commitEncoding` (recorded); `i18n.logOutputEncoding` for `git log` display.
- Hooks: `pre-commit`, `prepare-commit-msg`, `commit-msg`, `post-commit`, `post-rewrite`.
- After a conflicted merge, fix and `git add` each file, then `git commit` with no file names.
