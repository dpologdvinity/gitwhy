# Writing command docs

Every git command gets one file: `docs/commands/<command>.md` (e.g. `cherry-pick.md`).
Global `git` options (`-C`, `-c`, `--no-pager`, ...) live in `docs/commands/git.md`.

## Rules

- Fewest words that fully say what it does. Fragments are fine. No filler ("This option allows you to...").
- Plain words. Explain jargon on first use only if the entry is unclear without it.
- Write it yourself. Never copy or closely paraphrase git's man pages.
- Every flag the command accepts gets a row. Include short and long forms in one row: `` `-a`, `--all` ``.
- Negatable flags take one row: `` `--[no-]verify` ``. Describe the positive form; add the negative only if it is not just "don't".
- Show values in angle brackets: `` `--author=<who>` ``.
- Escape `|` inside table cells as `\|`.
- Deprecated or alias commands: one line saying what to use instead. Still list their flags.
- Internal commands (git runs them for you): say so in the summary.

## Template

```markdown
# git <command>

<One sentence. A few more only if the command is complicated.>

**Usage:** `git <command> [<flags>] <args>`

## Flags

| Flag | What it does |
|------|--------------|
| `-a`, `--all` | Stage all changed tracked files first. |
| `--[no-]verify` | Run pre-commit and commit-msg hooks. |
```

Commands with subcommands (`stash`, `remote`, `worktree`, `lfs`, ...) add a subcommand table,
then one `### git <command> <subcommand>` section per subcommand with its own flag table:

```markdown
## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `push` | Save changes to the stash and clean the working tree. |

### git stash push

| Flag | What it does |
|------|--------------|
| `-u`, `--include-untracked` | Also stash untracked files. |
```

Commands that accept the shared option sets link to them instead of repeating them:

- Diff output options: [`shared/diff-options.md`](shared/diff-options.md)
- Commit selection and history options: [`shared/revision-options.md`](shared/revision-options.md)

Write the link as a line under `## Flags`: `Also takes all [diff options](../shared/diff-options.md).`
