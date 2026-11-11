# gitwhy v1 design

## Goal

A CLI that lets people who don't know every git command (interns, devs who look things up) do anything
with git without reading git's docs, and recover when something goes wrong.

Success: a user can state a goal ("undo last commit, keep changes staged"), see exactly what will happen,
run it, understand the command that ran, and undo it.

## Constraints

- No AI, LLM, network service or API key at runtime.
- Runs the real `git` CLI (behaves exactly like git). Requires git >= 2.44 (`git replay`).
- Platforms: WSL/Linux at launch. Git Bash next. PowerShell last. macOS untested.
- Command name: `gitwhy`. Go, single static binary.
- `docs/commands/*.md` and `docs/shared/*.md` are the single source of truth for command and flag text.

## Scope

v1:
- **Do something** (search, category menu, wizard questions, generic builder).
- **What just happened / undo** (gitwhy snapshots; reflog recovery for plain-git commands).
- Preview before every run; danger warnings; confirm.
- Teaching breakdown of every command run (on by default, `GITWHY_EXPLAIN=0` or `--quiet` to hide).
- Conflict resolution wizard.

Later (not v1): explain a pasted command; shell hook that snapshots plain-git commands; TUI with
lazygit parity (see end); graph views; Git Bash and PowerShell support.

## Architecture

```
cmd/gitwhy/      main: args -> ui
embed.go         package gitwhy: go:embed docs/commands docs/shared recipes
internal/git     the only package that runs git; Runner{Dir}; Run/Output/OK
internal/kb      parse docs markdown -> Command{Name, Summary, Usage, Flags, Subcommands}; load recipes
internal/recipe  Recipe type, template rendering, checks
internal/search  fuzzy search over recipes, commands, flags
internal/snapshot save/list/restore repo state
internal/preview  exact (plumbing) or predicted preview of a command list
internal/recover  reflog -> events -> undo targets; lost staged files
internal/explain  per-token breakdown of a git command (flags from kb, revision glossary)
internal/ui       huh/lipgloss prompts; wires everything
```

Dependencies: `charm.land/huh` (or `github.com/charmbracelet/huh`), `lipgloss`, `gopkg.in/yaml.v3`,
`github.com/sahilm/fuzzy`. Standard library for args.

## Entry points

- `gitwhy` - home menu: Do something / What just happened? / Undo last gitwhy action / Any git command.
- `gitwhy <words...>` - search, pick result.
- `gitwhy undo [--list]` - restore a gitwhy snapshot.
- `gitwhy recover` - list recent events from reflogs, undo one.
- `gitwhy run <git args...>` - preview + snapshot + run + explain an exact command (used by the builder).
- `gitwhy help`, `gitwhy version`.

## Knowledge base parsing

Each command file: `# git <name>`, summary paragraph, optional `**Usage:**`, `## Flags` table,
optional `## Subcommands` table and `### git <name> <sub>` sections with their own flag tables, links to
shared option pages. Flag cell: one or more backticked spellings separated by `, `; `--[no-]x` means
both `--x` and `--no-x`; `=<value>` / `[=<value>]` marks a required/optional value. A flag is
dangerous if its description contains "**lost**", "deleted", "discard", "overwrite", "rewrites
history" or similar marker words (heuristic; recipes declare danger explicitly).

## Recipes

YAML in `recipes/`, embedded. Fields: `id, title, aliases, category, checks, questions, steps, preview,
explain` (optional extra text). Question types: `choice`, `confirm`, `text`, `branch`, `commit`,
`files`, `reorder`. Steps are git arg lists rendered with Go `text/template` over answers; a step may
have `when:` to run conditionally. Checks are Go functions keyed by name: `dirty-tree`, `head-is-pushed`,
`in-progress-op`, `detached-head`, `has-conflicts`, `no-upstream`; each can `warn`, `block`, or
`suggest` another recipe.

Categories: undo, rewrite history, branches, sync with remote, merge & conflicts, stash, files, LFS,
monorepo, inspect history, recover. v1 ships the user's 16 situations first, then grows to ~60.

## Snapshot (`.git/gitwhy/`)

Taken before every gitwhy-run command. Contents: refs + HEAD (`for-each-ref`, `symbolic-ref`), byte copy
of `.git/index`, working-tree commit (temp index seeded from the real index, `add -A`, `write-tree`,
`commit-tree`) pinned by `refs/gitwhy/snapshots/<id>`, in-progress op state files (MERGE_HEAD,
rebase-merge/, rebase-apply/, CHERRY_PICK_HEAD, REVERT_HEAD, BISECT_*), stash reflog entries, and the
command that ran. Ignored files are not saved, except paths a command's own dry-run says it will delete.
Keep last 20.

Restore: snapshot current state first; reset refs and HEAD (delete refs created since); materialize the
snapshot tree into the working tree via a temp index (`read-tree` + `checkout-index -a -f`) and delete
files added since (diff of the two snapshot trees, added only); copy the index back; restore op state
files and stash entries.

Cannot undo remote effects; push undo is offered as a separate force-push recipe with a warning.
Known risk: `git push --mirror` pushes `refs/gitwhy/*`; gitwhy warns when it runs one.

## Preview

Per step: `exact` when gitwhy can simulate with plumbing, else `predict` text from the recipe or the
flag descriptions. Always lists danger warnings.

- `reset`: ref move, commits leaving the branch, files changed per mode (`--soft/--mixed/--hard/--keep/--merge`).
- `branch -f`, `update-ref`, `switch -C`: ref move + commits gained/lost.
- `rebase`/`cherry-pick`/`revert`: `git replay` / `merge-tree --write-tree` on throwaway refs; report
  resulting commits and predicted conflicts.
- `merge`/`pull` (local part): `merge-tree --write-tree`; fast-forward detection.
- `clean`, `rm`, `add`, `mv`: their `-n`/`--dry-run`.
- network commands: predict only.

## Recover

Parse `git reflog` for HEAD and each branch; group entries into events (commit, amend, reset, merge,
pull, rebase start..finish, checkout, cherry-pick, stash). Each event: time, plain-English text, undo =
move branch (and HEAD) to the pre-event hash, via a gitwhy-run `reset --keep` or `update-ref` so it is
itself snapshotted. Plus: list dangling blobs from `fsck --lost-found` as "lost staged files".

## Teaching

After the preview and after running: the command, then one line per token. Flags: kb description.
Revisions: glossary (`HEAD`, `HEAD~n`, `HEAD^`, `@{u}`, `A..B`, `A...B`, `origin/x`, `<sha>`, `-`).
Paths after `--`: "files to act on".

## Error handling

- Not in a repo: only search/builder/explain work; say so for run/undo.
- git missing or < 2.44: clear message with the version found.
- A step fails: stop, show git's stderr, offer `gitwhy undo` (snapshot exists).
- Snapshot fails: do not run the command.

## Testing

- Unit: kb parser against real docs (every command file parses; flags count > 0 where tables exist).
- Integration: temp repos created in tests (`t.TempDir()`), real git. Snapshot round-trip
  (staged, unstaged, untracked, deleted, conflict state, branch moves). Each recipe renders and its
  preview runs. Recover parses reflogs from scripted histories.
- `scripts/check_coverage.py` keeps docs complete.

## Build order

1. Module, git runner, kb parser.
2. Snapshot + undo.
3. Recipe engine + checks + first 16 recipes.
4. Preview.
5. Explain.
6. UI: search, menu, wizard, builder, run.
7. Recover.
8. Conflict wizard; more recipes; README.

## TUI phase (later): lazygit parity checklist

Status/files/branches/commits/stash panels; line and hunk staging; keyboard interactive rebase
(move, squash, fixup, drop, reword); copy/paste cherry-pick; conflict view with per-hunk
ours/theirs/both; diff viewer; filter commits by path/author; branch compare; custom commands and
keybindings; worktrees, submodules, bisect, remotes panels.
