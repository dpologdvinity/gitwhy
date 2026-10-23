# git credential-cache

Credential helper that keeps passwords in memory for a while. A daemon holds them, reachable only by you via a Unix socket; lost on timeout or daemon exit (e.g. reboot). Poor fit for long-lived tokens.

**Usage:** `git config credential.helper 'cache [<flags>]'`
or `git credential-cache [<flags>] <action>`

Actions: `get`, `store`, `erase` (sent by git), and `exit` (stop the daemon, forget everything).

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]timeout <seconds>` | How long to keep credentials. Default 900 (15 min). |
| `--[no-]socket <path>` | Absolute daemon socket path; starts a daemon if none. Default `~/.git-credential-cache/socket` if that folder exists, else `$XDG_CACHE_HOME/git/credential/socket`. Avoid network-mounted homes. |
