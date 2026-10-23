# git remote-ftps

Internal: handles `ftps://` remote URLs for `git fetch`, `git clone`, `git ls-remote` (fetch only). Same program as [`git remote-http`](remote-http.md), [`git remote-https`](remote-https.md), [`git remote-ftp`](remote-ftp.md) (`git-remote-curl`). Slow and deprecated; use `https://` or `ssh://`.

Speaks the [remote-helper protocol](https://git-scm.com/docs/gitremote-helpers) on stdin. Configure with `http.*` and env vars (`GIT_SSL_NO_VERIFY`, `GIT_CURL_VERBOSE`, `GIT_TRACE_CURL`, ...); see [`git config`](config.md).

**Usage:** `git remote-ftps <remote> [<url>]`

## Flags

None. Arguments:

| Argument | What it is |
|----------|------------|
| `<remote>` | Remote name (e.g. `origin`), or the URL if unnamed. |
| `<url>` | URL to use. Default: `remote.<name>.url`. |
