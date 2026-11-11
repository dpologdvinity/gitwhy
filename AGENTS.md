# AGENTS.md

Instructions for coding agents (Claude Code, Codex and others) working in this repository.
`CLAUDE.md` imports this file, so this is the single source of truth.

## What this is

gitwhy is a Go CLI and TUI that helps people do things in git without reading git's docs. It previews
what a command will do, explains it, runs the real `git`, and can undo it. Read
[ARCHITECTURE.md](ARCHITECTURE.md) before changing code and [CONTRIBUTING.md](CONTRIBUTING.md) for the
full workflow.

## Setup

- Go version from `go.mod`, git 2.40 or newer, optional Node 18+ for the npm package tests.
- `go build -o gitwhy ./cmd/gitwhy`
- `lefthook install` once after cloning. The pre-commit hooks run gitleaks, gofmt, golangci-lint and
  typos on staged files.

## Checks (run all before you call a change done)

```sh
gofmt -l $(git ls-files '*.go' | grep -v '^third_party/')   # must print nothing
go vet ./...
go test ./...                    # real git in temporary repositories
golangci-lint run ./...          # must report "0 issues."
node --test npm/lib.test.js      # npm package
```

While working, run only the package you're changing (`go test ./internal/hunk/`). CI
(`.github/workflows/ci.yml`) is the authority. It also enforces a coverage floor (`COVERAGE_FLOOR`),
tidy modules, `THIRD_PARTY_LICENSES` and a goreleaser snapshot build.

## Rules that are easy to break

- Every repo change gitwhy makes goes through a snapshot (`snapshot.Take` + `snapshot.Finish`) so it can
  be undone. The one exception is staging and unstaging in the TUI (`quick` in `internal/tui/tui.go`).
- Tests use real git through `internal/testrepo`. Do not mock git.
- Keep the `replace github.com/atotto/clipboard => ./third_party/clipboard` line in `go.mod`. Do not
  reformat `third_party/`.
- Recipes in `recipes/*.yaml` pass user answers as separate arguments, never glued into one string.
- Write the docs in `docs/commands/` yourself. Never copy git's GPL-2.0 man pages. Check them with
  `python3 scripts/check_coverage.py <command>` and rebuild the index with `python3 scripts/build_index.py`.
- Keep user-facing text plain and short ([docs/STYLE.md](docs/STYLE.md)).
- Write the test first. For a bug, start with a failing test. Fix causes and never skip tests or
  silence linters.
- Pin GitHub Actions to full commit SHAs with the version in a comment.

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org) (`fix(tui): ...`) with one logical change
per commit. Don't add AI attribution trailers. Never commit secrets or `.env` files.

## Agent workflow

- **Claude Code:** `.claude/settings.json` enables the shared plugins: `gopls-lsp` (Go LSP),
  `commit-commands`, and [`ponytail`](https://github.com/DietrichGebert/ponytail) (pushes for the
  smallest complete change). Put personal overrides in `.claude/settings.local.json` (gitignored).
- **Codex:** `.codex/config.toml` enables the same `ponytail` plugin.
- **Specs and plans:** larger work starts with a design spec and a task-by-task implementation plan,
  written with the [superpowers](https://github.com/obra/superpowers) brainstorming and writing-plans
  skills and stored in `docs/superpowers/specs/` and `docs/superpowers/plans/`. The plans assume
  subagent-driven execution, with narrow tests per task and the full suite at the end of each phase.
- No MCP servers are required. The project needs no API keys or environment variables at build or
  run time. Release secrets live only in GitHub Actions (see CONTRIBUTING.md, "Releasing").
