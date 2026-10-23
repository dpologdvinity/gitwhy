# git daemon

Tiny `git://` server (TCP 9418). No auth, no encryption: best for public read-only repos. Serves only repos with a `git-daemon-export-ok` file (or all with `--export-all`), inside the listed folders.

Serves only fetches by default. See [Services](#services) for `git archive --remote` or pushes.

**Usage:** `git daemon [<flags>] [<directory>...]`

Example: `git daemon --reuseaddr --base-path=/srv/git /srv/git` serves `/srv/git/hello.git` as `git://<host>/hello.git`.

## Flags

| Flag | What it does |
|------|--------------|
| `<directory>...` | Only serve repos inside these. With `--strict-paths`, exact match. |
| `--strict-paths` | Exact paths: no `.git` or `/.git` guessing, no `~user`. Requires `<directory>` args. |
| `--base-path=<path>` | Resolve requested paths under `<path>`. |
| `--base-path-relaxed` | With `--base-path`: on failure, also try the path without the base. |
| `--interpolated-path=<pathtemplate>` | Build repo path from a template (virtual hosting): `%H` requested host (lowercased), `%CH` canonical host, `%IP` server IP, `%P` port, `%D` path. E.g. `/pub/%H%D`. Must stay inside a `<directory>`. |
| `--export-all` | Serve any repo-like folder, even without `git-daemon-export-ok`. |
| `--inetd` | Run under inetd (one connection on stdin/stdout). Implies `--syslog`. Not with `--detach`, `--port`, `--listen`, `--user`, `--group`. |
| `--listen=<host-or-ipaddr>` | Listen only on this address/host. Repeatable. Host names need IPv6 support. Not with `--inetd`. |
| `--port=<n>` | Port instead of 9418. Not with `--inetd`. |
| `--init-timeout=<n>` | Seconds to wait for the first request. Keep low. |
| `--timeout=<n>` | Seconds per sub-request, including waiting. |
| `--max-connections=<n>` | Max clients (default 32). `0`: unlimited. |
| `--syslog` | Same as `--log-destination=syslog`. |
| `--log-destination=(stderr\|syslog\|none)` | Log target. Default `syslog` with `--inetd`/`--detach`, else `stderr`. Errors only unless `--verbose`. |
| `--user-path`, `--user-path=<path>` | Allow `~user` paths: `git://host/~alice/foo` means `foo` (or `<path>/foo`) in alice's home. |
| `--verbose` | Log each connection and repo. |
| `--reuseaddr` | Rebind the port immediately on restart. |
| `--detach` | Run in background. Implies `--syslog`. |
| `--pid-file=<file>` | Write PID to `<file>`. Ignored with `--inetd`. |
| `--user=<user>`, `--group=<group>` | Drop to this user (and group; default user's main group). Names only. Not with `--inetd`. `HOME` isn't changed: export it first and make its git config readable. |
| `--enable=<service>`, `--disable=<service>` | Turn a service on/off site-wide. Repos may override if allowed. |
| `--allow-override=<service>`, `--forbid-override=<service>` | Allow/forbid per-repo override of a service. Default: allowed. |
| `--[no-]informative-errors` | Send real errors ("no such repository" vs "not exported"); reveals unexported repos. Default off: always "access denied". |
| `--access-hook=<path>` | Run `<path> <service> <repo-path> <host> <canonical-host> <ip> <port>` before each request. Non-zero exit refuses; its stdout line goes to the client. Sets `$REMOTE_ADDR`, `$REMOTE_PORT`. |

## Services

Toggle with `--enable`/`--disable`, or per repo in its `config` file.

| Service | What it serves | Default | Per-repo config |
|---------|----------------|---------|-----------------|
| `upload-pack` | `git fetch`, `git clone`, `git ls-remote`. | On | `daemon.uploadpack` |
| `upload-archive` | `git archive --remote`. | Off | `daemon.uploadarch` |
| `receive-pack` | `git push`. Unauthenticated: anyone can push or delete anything. Trusted closed networks only. | Off | `daemon.receivepack` |

## inetd setup

Add to `/etc/inetd.conf` (one line), with `git 9418/tcp` in `/etc/services`:

```
git stream tcp nowait nobody /usr/bin/git git daemon --inetd --verbose --export-all /pub/foo /pub/bar
```

## Environment

Sets `REMOTE_ADDR` to the client IP for hooks.
