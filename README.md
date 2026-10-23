# gitwhy

**Do anything with git without looking it up. See what will happen first. Undo it after.**

For interns, new grads, and anyone who still searches "how to undo last commit".
No AI, no account, no network: one binary that runs the real `git`.

![gitwhy demo: undo last commit with a preview, then undo the undo](docs/demo.gif)

## What it does

- **Say what you want.** `gitwhy move my last commit to a new branch` finds the guided recipe and asks only what it needs (which commit? which branch?).
- **See it before it happens.** gitwhy predicts the result with git's own dry runs: commits that leave a branch, files that change, and **conflicts before they happen** (for merge, rebase, cherry-pick, revert).
- **Learn as you go.** Every command it runs is shown and broken down part by part.
- **Undo anything it ran.** Before running, gitwhy saves branches, staged and unstaged changes, untracked files, stashes, and merge/rebase state. `gitwhy undo` puts it all back, even after `reset --hard` or `clean -fdx`.
- **Recover from plain git too.** `gitwhy recover` reads git's history of your branch ("2h ago: pulled 14 commits from origin/master") and takes you back to before any of it. It also finds lost commits and staged files git still has.
- **Every git command and flag.** `gitwhy build` walks you through any command with every flag explained in a few words, from [the docs in this repo](docs/README.md) (181 commands, ~3,500 flags, git 2.56).

## The home screen

Run `gitwhy` in a repo. A banner shows an operation in progress ("MERGE IN PROGRESS"), the status line shows
how far ahead or behind the remote you are, and **Next:** suggests what to do for the current state
(resolve conflicts, commit what's staged, push). Three tabs: Changes (with a live diff), Commits, Branches.
Keys act on the selected item and open the same guided recipe as the CLI, with the answers already filled
in: `r` reword a commit, `d` drop it, `f` add changes to it, `s` split it, `m` move it to a new branch.
Staging with space is instant, since it's always reversible. `?` shows the keys for the current tab.
`enter` on a changed file shows its hunks: `space` stages one, `v` picks single lines.
Commits tab draws the branch graph; `a` shows all branches, `p` copies a commit here.

## Commands

| | |
|---|---|
| `gitwhy` | Full-screen home: changes, commits, branches, and the next step for your repo's state |
| `gitwhy menu` | Simple menu |
| `gitwhy <words>` | Search, e.g. `gitwhy squash commits` |
| `gitwhy do [recipe]` | Run a guided recipe by name (no name: list them) |
| `gitwhy run <git command>` | Preview, explain, run with undo: `gitwhy run git rebase main` |
| `gitwhy explain <git command>` | Explain each part without running it |
| `gitwhy build [command]` | Build any git command from its flags |
| `gitwhy undo [--list]` | Undo the last gitwhy action (or pick one) |
| `gitwhy recover` | What happened recently, and go back |
| `gitwhy resolve` / `continue` / `abort` | Fix conflicts, then continue or cancel a merge/rebase |

`-q` / `GITWHY_EXPLAIN=0` hides the breakdown. `ACCESSIBLE=1` uses plain-text prompts (screen readers, scripts).

## Guided recipes

Undo last commit (keep staged / unstaged / delete) · undo a pushed commit safely · change a commit message (any commit) · remove a commit · reorder, squash, fixup, rename or drop recent commits · squash last commits · split a commit in two · add changes to an older commit · move commits to a new or existing branch · copy a commit from another branch · make a branch the same as master · make a branch exactly match the remote · create, switch, rename, delete branches (local and remote) · merge · update my branch from master (rebase or merge) · resolve conflicts, continue, abort · push (normal / safe force / force) · pull · stash and restore · discard, stage, unstage files · delete untracked files · fix a commit's author · tags: create, delete, push · remotes: add, change URL, set upstream · stop tracking a file but keep it · rename a file · get an old version of a file back · Git LFS: track, migrate, pull, list · remove a file from all history · worktrees · sparse checkout on/off · submodules · find a commit by message or code · find which commit broke something (bisect) · history, file history, compare branches, blame, see my changes.

`gitwhy do` lists them all. Each is a small YAML file in [`recipes/`](recipes); add your own.

## Install

Needs git 2.38+. Linux, WSL and Windows (Git Bash or PowerShell, CI-tested) work; macOS builds but is untested.
On Windows, with [Scoop](https://scoop.sh):

```powershell
scoop bucket add gitwhy https://github.com/dpologdvinity/homebrew-gitwhy
scoop install gitwhy
```

Or download the zip from the [releases](https://github.com/dpologdvinity/gitwhy/releases) page.

```sh
curl -fsSL https://raw.githubusercontent.com/dpologdvinity/gitwhy/main/install.sh | sh
```

**Debian, Ubuntu** (apt):

```sh
sudo curl -fsSL https://dpologdvinity.github.io/gitwhy/gitwhy.asc -o /etc/apt/keyrings/gitwhy.asc
echo "deb [signed-by=/etc/apt/keyrings/gitwhy.asc] https://dpologdvinity.github.io/gitwhy/apt ./" | sudo tee /etc/apt/sources.list.d/gitwhy.list
sudo apt update && sudo apt install gitwhy
```

**Fedora, RHEL** (dnf):

```sh
sudo curl -fsSL https://dpologdvinity.github.io/gitwhy/gitwhy.repo -o /etc/yum.repos.d/gitwhy.repo
sudo dnf install gitwhy
```

Each release also has `.deb`, `.rpm`, `.apk` (Alpine) and `.pkg.tar.zst` (Arch) files you can install directly.

Or with Homebrew: `brew install dpologdvinity/gitwhy/gitwhy`. Or with npm or bun (Node 18+):
`npm i -g @kaitlynbassford/gitwhy` / `bun add -g @kaitlynbassford/gitwhy`, or `npx @kaitlynbassford/gitwhy` to try it (the command is still `gitwhy`). The binary is downloaded from the
release and checked against its checksum, at install or on first run.

That downloads the latest [release](https://github.com/dpologdvinity/gitwhy/releases), checks its
checksum, and puts `gitwhy` in `~/.local/bin`. Or build from source (Go 1.24+):

```sh
git clone https://github.com/dpologdvinity/gitwhy && cd gitwhy
go build -o ~/.local/bin/gitwhy ./cmd/gitwhy
```

(`go install ...@latest` doesn't work: `go.mod` replaces a clipboard library with a copy that
doesn't slow startup on WSL, and `go install` refuses modules with `replace` directives.)

## How undo works

Snapshots live in `.git/gitwhy/` (last 20). The working-tree copy is a normal git commit kept alive by
`refs/gitwhy/snapshots/<id>`, so it costs little space. Ignored files are not saved, except files a
command is about to delete (gitwhy asks git's dry run which ones). Anything already sent to a remote
(`push`) can't be undone locally.

`git push --mirror` would also push `refs/gitwhy/*`; gitwhy warns when it runs one. After removing a
leaked secret from history, the old commits stay in your local undo points until they're pruned
(delete `.git/gitwhy` and `git update-ref -d` the `refs/gitwhy/*` refs to drop them sooner).

Undo leaves alone any branch that changed after the action (for example a commit you made later in
plain git, or work in another worktree), and tells you which.

## Undo for plain git (optional)

```sh
gitwhy hook install
```

This wraps `git` in your shell so `gitwhy undo` works for commands you type yourself. It snapshots
before risky commands only (like `reset --hard` or `clean -fd`), costs a few ms on other commands, and works
in bash and zsh. Set `GITWHY_HOOK_OFF=1` to pause it; `gitwhy hook uninstall` removes it.

## Development

```sh
go test ./...                    # unit + end-to-end tests (real git, temp repos)
python3 scripts/check_coverage.py   # every flag in your git's -h/man pages is documented
python3 scripts/build_index.py      # rebuild docs/README.md
```

Docs are original plain-English descriptions (git's own docs are GPL-2.0; nothing is copied).
Format: [docs/STYLE.md](docs/STYLE.md).

## License

MIT. `third_party/clipboard` is BSD-3-Clause (github.com/atotto/clipboard, changed to detect tools lazily).
