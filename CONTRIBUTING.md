# Contributing

Thanks for helping. This page covers how to build, test and change gitwhy. For how the program
is put together, read [ARCHITECTURE.md](ARCHITECTURE.md) first.

## Set up

You need:

- **git 2.38 or newer** (previews use `git merge-tree --write-tree`).
- **Go 1.26 or newer** (the version in `go.mod`).
- Optional: Node 18+ (the npm package tests), Docker (the package and repo tests), and
  [golangci-lint](https://golangci-lint.run) v2 (the same linter CI runs).

```sh
git clone https://github.com/dpologdvinity/gitwhy && cd gitwhy
go build -o gitwhy ./cmd/gitwhy
./gitwhy version
```

`go install ...@latest` does not work for this module; see "Vendored clipboard library" in
[ARCHITECTURE.md](ARCHITECTURE.md#decisions-and-trade-offs).

## Run the checks

Run these before you open a pull request. CI runs the same ones on Linux and Windows.

```sh
go vet ./...
go test ./...                       # unit and end-to-end tests; real git, temporary repositories (about 90 s)
golangci-lint run ./...             # must report "0 issues."
node --test npm/lib.test.js         # the npm package
```

Run just one package or test while you work:

```sh
go test ./internal/hunk/
go test ./cmd/gitwhy/ -run TestUndoLastCommit
```

If you change packaging (`.goreleaser.yaml`, `scripts/build-repos.sh`, the workflows), also run the
Docker-based install tests:

```sh
goreleaser release --snapshot --clean --skip=publish   # builds dist/
scripts/test-repos.sh dist     # signed apt and dnf repos: build, install, reject tampering
scripts/test-pkgs.sh dist      # .apk (Alpine) and Arch packages
```

## How we work

- **Write the test first.** For a bug, add a test that fails, then fix it. For new logic, add a
  test that would fail if the logic broke. Tests use real git in temporary repositories
  (`internal/testrepo`); do not mock git.
- **Fix the cause, not the symptom.** Do not skip a test or silence a linter to get green.
- **Comment the why.** Every package has a package comment, and every function, type and constant
  has a doc comment that starts with its name and says what it does. Comment non-obvious steps,
  not obvious ones.
- **Keep text plain and short.** Messages shown to users and the docs use everyday words. See
  [docs/STYLE.md](docs/STYLE.md).
- **Use [Conventional Commits](https://www.conventionalcommits.org):** `type(scope): description`,
  for example `fix(tui): keep the cursor on the same file after a reload`. Keep each commit to one
  logical change.
- **Never commit secrets.** No keys, tokens or `.env` files.

## Common changes

### Add or change a recipe

A recipe is one YAML file in `recipes/`. Copy a similar one (`recipes/copy-commit.yaml` is short)
and follow the model in `internal/recipe/recipe.go`. Rules:

- Pass user answers as **separate arguments** (`git: [cherry-pick, "{{.commit}}"]`), never glued
  into a string, and put `--` before file or branch names that could start with a dash.
- Write a `predict` line that says what really happens. If a step cannot be undone, say so.
- `go test ./internal/recipe/` loads every recipe, and the end-to-end tests in
  `cmd/gitwhy/e2e_test.go` run many of them against a real repository. Add an end-to-end test for
  a recipe that changes history.

### Add a danger warning

Edit the `rules` table in `internal/danger/danger.go`: the command, a condition on the parsed
flags, and a plain-English message. Add a case to `TestWarnings`. The shell hook picks up a new
command automatically (`danger.Commands()`).

### Document a git command or flag

Edit `docs/commands/<command>.md` and follow [docs/STYLE.md](docs/STYLE.md). Write the text
yourself; do not copy git's man pages (they are GPL-2.0). Then check your work:

```sh
python3 scripts/check_coverage.py <command>   # flags in your git that the docs miss
python3 scripts/build_index.py                # rebuild docs/README.md
```

### Change the shell hook

The shell function lives in `internal/hook/hook.go`. It must always run the real git and keep its
exit code, even when gitwhy is missing. `internal/hook/hook_test.go` and the end-to-end test
`TestHookUndoesPlainGit` check this; keep them passing.

## Pull requests

- Branch from `main` and keep the pull request small and focused.
- Describe what changed and why, and say how you checked it. Say plainly what you could **not**
  check (for example, anything that only runs on a real release).
- CI must be green on Linux and Windows. Windows has no local test machine, so CI is the test.
- Merge with a merge commit.

## Releasing

Maintainers only. Tag the commit on `main`:

```sh
git tag vX.Y.Z && git push origin vX.Y.Z
```

The release workflow builds and publishes everything; see
[Build, release and packaging](ARCHITECTURE.md#build-release-and-packaging). The `verify-install`
workflow then installs the new release every supported way. If a job fails, fix it with a pull
request and tag a new patch version; do not re-use a published tag, because npm and package
repositories do not allow it.

## Reporting a security problem

Do not open a public issue. Use GitHub's private vulnerability reporting on the repository's
Security tab.
