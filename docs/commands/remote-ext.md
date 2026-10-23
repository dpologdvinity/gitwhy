# git remote-ext

Internal: handles `ext::` URLs by running your command and talking to the git server over its stdin/stdout. Reach a server through any program (custom SSH flags, socat, tunnels) without editing config.

Blocked by default (`fatal: transport 'ext' not allowed`): the URL runs any command, so allow it only for URLs you trust, e.g. `git -c protocol.ext.allow=always clone ext::...`.

The far end must be `git-upload-pack`, `git-receive-pack`, `git-upload-archive`, or a `git://` daemon.

**Usage:** `git remote add <name> "ext::<command>[ <arguments>...]"` (or any `git fetch`/`clone`/`push` with an `ext::` URL)

## Flags

None. Arguments split on unescaped spaces. Special codes:

| Code | Meaning |
|------|---------|
| `% ` | Literal space. |
| `%%` | Literal `%`. |
| `%s` | Short service name: `upload-pack`, `receive-pack`, or `upload-archive`. |
| `%S` | Long service name: `git-upload-pack`, etc. |
| `%G<repo>` | Must start an argument; not passed on. First send a `git://` request for `<repo>`. For `git://` servers behind a tunnel. |
| `%V<host>` | Must start an argument; not passed on. Virtual host for that `git://` request. |

Examples:

- `ext::ssh -i ~/.ssh/key user@host.example %S foo/repo`: like `host.example:foo/repo` with that key and user.
- `ext::socat -t3600 - ABSTRACT-CONNECT:/git-server %G/somerepo`: `git://` repo `/somerepo` at an abstract socket.

## Environment variables

| Variable | What it does |
|----------|--------------|
| `GIT_TRANSLOOP_DEBUG` | Print debug info for each read and write. |
| `GIT_EXT_SERVICE` | Set for the command: long service name. |
| `GIT_EXT_SERVICE_NOPREFIX` | Set for the command: short service name. |
