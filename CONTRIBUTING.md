# Contributing

Thanks for helping. This page covers how to build, test and change gitwhy. For how the program
is put together, read [ARCHITECTURE.md](ARCHITECTURE.md) first.

## Set up

You need:

- **git 2.40 or newer** (previews use `git merge-tree --write-tree --merge-base`).
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

Run these before you open a pull request. CI runs the same ones, plus a few more (see below).

```sh
gofmt -l $(git ls-files '*.go' | grep -v '^third_party/')   # must print nothing
go vet ./...
go test ./...                       # unit and end-to-end tests; real git, temporary repositories (about 90 s)
golangci-lint run ./...             # must report "0 issues."
node --test npm/lib.test.js         # the npm package
```

CI (`.github/workflows/ci.yml`) is the authority. On top of the commands above it runs `govulncheck`,
`actionlint`, `shellcheck`, checks that `go.mod`/`go.sum` are tidy and that `THIRD_PARTY_LICENSES` is up
to date (`scripts/third-party-licenses.sh` regenerates it), builds every release artifact with goreleaser
in snapshot mode, and builds the Nix flake. All of it must be green before a merge.

### Coverage

The Linux test job measures statement coverage and fails below the floor set by `COVERAGE_FLOOR` in
`ci.yml` (45% today; the real number was 46.6% when it was set). Only code run in-process by `go test`
counts: the end-to-end tests run the built binary, so `cmd/gitwhy` and `internal/ui` show 0%. When you
add tests that raise the total, raise the floor in the same change so it never drops back. Check it
locally with:

```sh
go test -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -1
```

### Pinned actions and tools

Every third-party GitHub Action is pinned to a full commit SHA with its version in a comment, and every
tool the workflows install has an exact version. Do not use a tag like `@v4`. Dependabot opens a weekly
grouped pull request that moves both. To pin a new action by hand, resolve the tag to a commit with
`gh api repos/<owner>/<repo>/git/ref/tags/<tag>` (for an annotated tag, follow the `tag` object to its
commit). If `go.mod` or `go.sum` change, rebuild the Nix flake and update `vendorHash` in `flake.nix`:
set it to `pkgs.lib.fakeHash`, build once, and copy the hash from the error.

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

The release workflow runs the CI gates first and then builds and publishes everything; see
[Build, release and packaging](ARCHITECTURE.md#build-release-and-packaging).

Repository secrets the release uses (Settings -> Secrets and variables -> Actions):

| Secret | Used for | Required |
|---|---|---|
| `HOMEBREW_TAP_TOKEN` | pushes the Homebrew cask and Scoop manifest to `dpologdvinity/homebrew-gitwhy` (needs Contents read and write on that repository) | yes |
| `GPG_PRIVATE_KEY` | signs the apt and dnf repositories | yes |
| `GPG_PASSPHRASE` | unlocks that key | only if the key has a passphrase |
| `NPM_TOKEN` | publishes `@kaitlynbassford/gitwhy` | yes |
| `WINGET_TOKEN` | opens the winget-pkgs pull request | no (the step is skipped without it) |

Before tagging, run the `check-secrets` workflow from the Actions tab: it reports which secrets are set,
and checks that the tap token can write to the tap and that the signing key can sign. It never prints a
secret. The `verify-install`
workflow then installs the new release every supported way. If a job fails, fix it with a pull
request and tag a new patch version; do not re-use a published tag, because npm and package
repositories do not allow it.

## Reporting a security problem

Do not open a public issue. Use GitHub's private vulnerability reporting on the repository's
Security tab.
