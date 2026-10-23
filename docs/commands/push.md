# git push

Upload commits to another repo and update its branches or tags. `git push origin main` sends local `main` to `main` on `origin`. Only missing objects are sent.

What gets pushed, first match wins: command-line refspecs or `--all`/`--mirror`/`--tags`; `remote.<name>.push`; `push.default` (default `simple`: current branch to its same-named upstream). Fails without an upstream; `git push -u origin <branch>` sets one.

Remote groups push to several remotes: `git config remotes.all-remotes "origin gitlab backup"`, then `git push all-remotes`. Each uses its own config.

Rejected as not a fast-forward? `git pull` (or `--rebase`), fix conflicts, push again. Force-push only to deliberately discard their commits.

**Usage:** `git push [<flags>] [<repository> [<refspec>...]]`

## Flags

| Flag | What it does |
|------|--------------|
| `<repository>` | Remote name, URL (see [URLs](fetch.md#urls)), or group. Default: the branch's upstream remote, else `origin`. |
| `<refspec>...` | `[+]<src>[:<dst>]`. See [Refspecs](#refspecs). |
| `--[no-]repo=<repository>` | Same as `<repository>`; the argument wins. |
| `--[no-]all`, `--[no-]branches` | Push all branches. No other refspecs allowed. |
| `--[no-]prune` | Delete remote branches with no local counterpart, per refspec: `git push --prune origin 'refs/heads/*:refs/tmp/*'` deletes `refs/tmp/foo` if `foo` is gone. |
| `--[no-]mirror` | Make the remote's `refs/` match yours: create, force-update, and delete. Default if `remote.<name>.mirror` is set. |
| `-n`, `--[no-]dry-run` | Do everything except send. |
| `--[no-]porcelain` | Tab-separated script output on stdout, full ref names. |
| `-d`, `--[no-]delete` | Delete the listed remote refs. `git push origin -d dev`. |
| `--[no-]tags` | Also push all tags. Not with `--all`, `--branches`, or `--mirror`. |
| `--[no-]follow-tags` | Also push missing annotated tags on pushed commits. Default `push.followTags`. |
| `--[no-]signed[=(true\|false\|if-asked)]` | GPG-sign the push for server hooks. `true`: fail if unsupported. `false`: don't sign. `if-asked`: sign if supported. Fails if `gpg` fails. |
| `--[no-]atomic` | All refs update or none. Fails if unsupported. |
| `-o <option>`, `--[no-]push-option=<option>` | Send a string to server `pre-receive`/`post-receive` hooks (no NUL or newline). Repeatable. Default `push.pushOption`. |
| `--[no-]receive-pack=<git-receive-pack>`, `--[no-]exec=<git-receive-pack>` | Remote `git-receive-pack` path if not on its `$PATH`. |
| `-f`, `--[no-]force` | Allow non-fast-forward updates, overriding [push rules](#push-rules) and `--force-with-lease`. Remote can lose commits. Applies to every pushed ref (may rewind others with `push.default=matching` or multiple `remote.<name>.push`); force one with `git push origin +main`. |
| `--[no-]force-with-lease[=<refname>[:<expect>]]` | Force only if the remote ref is where you expect. Alone: each ref must match its remote-tracking branch. `=<refname>`: just that ref. `=<refname>:<expect>`: must equal `<expect>` (empty: must not exist); the only stable form. Background fetches (IDEs, cron) defeat the plain forms; give `<expect>` or add `--force-if-includes`. `--no-force-with-lease` cancels. |
| `--[no-]force-if-includes` | With `--force-with-lease` (no `:<expect>`): also require the remote-tracking tip to be in your reflog, i.e. integrated. |
| `-u`, `--[no-]set-upstream` | Set each pushed (or up-to-date) branch's upstream. |
| `--[no-]thin` | Send a thin pack (default). |
| `-q`, `--[no-]quiet` | No output unless something fails. |
| `-v`, `--[no-]verbose` | More output. |
| `--[no-]progress` | Force progress off a terminal. |
| `--[no-]recurse-submodules=(check\|on-demand\|only\|no)` | Submodule commits being pushed: `check`: abort if any isn't on a submodule remote. `on-demand`: push them first (abort on failure). `only`: push submodules only. `no`: don't check (overrides `push.recurseSubmodules`). `on-demand`/`only` recurse per submodule config. |
| `--verify` | Run the `pre-push` hook (default). |
| `--no-verify` | Skip `pre-push`. |
| `-4`, `--ipv4` | IPv4 only. |
| `-6`, `--ipv6` | IPv6 only. |

## Refspecs

`[+]<src>[:<dst>]`. `<src>`: usually a branch, any commit expression (`HEAD^`); `<dst>`: a remote ref name. `+` forces that ref.

| Refspec | Means |
|---------|-------|
| `main` | `main:refs/heads/main` (unless `remote.<name>.push` maps it). |
| `main:other` | Local `main` to remote `other`. |
| `HEAD:v1.0` | Update existing remote ref `v1.0` (here a tag, so rejected without `+`). |
| `refs/heads/main:refs/heads/main` | Fully spelled out. |
| `:` (or `+:`) | Matching: every local branch that exists on the remote. |
| `refs/heads/*:refs/heads/*` | Glob: all branches (one `*` per side). |
| `^refs/heads/dev-*` | Exclude these. `git push origin 'refs/heads/*' '^refs/heads/dev-*'`. |
| `:dev` | Delete remote `dev`. |
| `tag v1.0` | `refs/tags/v1.0:refs/tags/v1.0`. |

Ambiguous refspecs fail with a hint (`advice.pushUnqualifiedRefname`). (2.56+) `git push origin/main` gets a likely-typo hint.

## Push rules

Creating and deleting refs is allowed (unless the server forbids it). Updates:

1. Branches (`refs/heads/*`): fast-forward only. Must be a commit.
2. Tags (`refs/tags/*`): always rejected.
3. Other refs: not trees or blobs. Commits and tags may fast-forward, swap a tag for another on the same commit, or swap tag and commit.

`--force` or `+` overrides these, except branches still need commits and server settings/hooks (`receive.denyNonFastForwards`, `receive.denyDeletes`, `pre-receive`, `update`) still apply.

## Output

Per ref: `<flag> <summary> <from> -> <to> (<reason>)`. `--porcelain`: `<flag>\t<from>:<to>\t<summary> (<reason>)`. Up-to-date refs only with `-v` or `--porcelain`.

| Flag | Meaning |
|------|---------|
| ` ` | Fast-forward. |
| `+` | Forced update. |
| `-` | Deleted. |
| `*` | New ref. |
| `!` | Rejected or failed. |
| `=` | Up to date. |

`<summary>`: `<old>..<new>` (`<old>...<new>` if forced), usable with `git log`. Failures: `rejected` (not sent; usually not a fast-forward), `remote rejected` (hook or `receive.deny*`, e.g. pushing to a checked-out branch), `remote failure` (no confirmation; network or server trouble).

## push.default

| Value | Pushes |
|-------|--------|
| `simple` (default) | Current branch to its upstream, only if names match. Non-upstream remote: like `current`. |
| `current` | Current branch to the same name. |
| `upstream` | Current branch to its upstream, any name. |
| `matching` | Every branch with a same-named remote branch. |
| `nothing` | Nothing without a refspec. |
