# git upload-pack

Internal: server side of fetch and clone. `git fetch-pack` starts it on the remote (via SSH, `git daemon`, or `git http-backend`); it sends the objects the client lacks. Push uses `git receive-pack`.

Safe in untrusted repos: ignores dangerous config and hooks. For more safety, run as another user: `git clone --no-local --upload-pack='sudo -u nobody git-upload-pack' ...`.

**Usage:** `git upload-pack [--[no-]strict] [--timeout=<n>] [--stateless-rpc] [--http-backend-info-refs] <directory>`

## Flags

| Flag | What it does |
|------|--------------|
| `<directory>` | Repository to serve. |
| `--[no-]strict` | Don't also try `<directory>/.git`. |
| `--[no-]timeout=<n>` | Quit after `<n>` idle seconds. |
| `--[no-]stateless-rpc` | One request and response on stdin/stdout, then exit. For HTTP. |
| `--http-backend-info-refs`, `--advertise-refs` | List refs and exit. Used by `git http-backend` for `info/refs?service=git-upload-pack`. |

## Environment variables

| Variable | What it does |
|----------|--------------|
| `GIT_PROTOCOL` | Internal: wire-protocol version handshake. Some transports must pass it through. |
| `GIT_NO_LAZY_FETCH` | Default `1`: don't fetch missing objects from upstream when serving a partial clone (could run repo config commands). Set `0` only for trusted repos. |
