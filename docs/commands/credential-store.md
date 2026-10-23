# git credential-store

Credential helper that saves logins to a plain-text file. Git runs it; enable with `git config credential.helper store`. **Warning:** passwords are unencrypted, protected only by file permissions. Prefer `git credential-cache` or an OS keychain.

**Usage:** `git config credential.helper 'store [<flags>]'` (git calls `git credential-store [<flags>] <get|store|erase>`)

Files read in order: `~/.git-credentials`, then `$XDG_CONFIG_HOME/git/credentials` (or `~/.config/git/credentials`). Writes go to the first that exists, else `~/.git-credentials`. Erase hits both.

Format: one URL per line, `https://user:pass@example.com`. No blanks or comments. Don't hand-edit.

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]file=<path>` | Use this file instead of the defaults. |
