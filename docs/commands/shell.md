# git shell

Restricted login shell for SSH accounts that only push and fetch. Set it (`chsh -s $(command -v git-shell) <user>`); the account can then run only the commands below plus scripts in `~/git-shell-commands/`.

**Usage:** `git shell -c "<command> <argument>"`, or `git shell` with no arguments (interactive; needs `~/git-shell-commands/`)

## Flags

| Flag | What it does |
|------|--------------|
| `-c <command>` | Run one allowed command (how SSH passes it). |

## Allowed commands

| Command | What it does |
|---------|--------------|
| `git receive-pack <argument>` | Serve `git push`. |
| `git upload-pack <argument>` | Serve `git fetch`/`git clone`. |
| `git upload-archive <argument>` | Serve `git archive --remote`. |
| `cvs server` | Act as a CVS server (see `git cvsserver`). |
| `<command> <arguments>` | Run `~/git-shell-commands/<command>`, if present. |

## Interactive use

Needs `~/git-shell-commands/`. Runs its `help` (if any), then a `git> ` prompt for commands in that folder, or `exit`.

If `~/git-shell-commands/no-interactive-login` exists, it runs instead and the session ends. To block logins with a message (exit 128):

```sh
mkdir $HOME/git-shell-commands
cat >$HOME/git-shell-commands/no-interactive-login <<\EOS
#!/bin/sh
printf '%s\n' "Hi $USER! You're authenticated, but there's no shell access."
exit 128
EOS
chmod +x $HOME/git-shell-commands/no-interactive-login
```
