# git remote-http

Internal: handles `http://` remote URLs for `git fetch`, `git clone`, `git push`, `git ls-remote`. Same program as [`git remote-https`](remote-https.md), [`git remote-ftp`](remote-ftp.md), [`git remote-ftps`](remote-ftps.md) (`git-remote-curl`). Supports the "smart" protocol and the old "dumb" one (plain file downloads; see [`git http-fetch`](http-fetch.md)).

Speaks the [remote-helper protocol](https://git-scm.com/docs/gitremote-helpers) on stdin. Configure with `http.*` and env vars (`GIT_SSL_NO_VERIFY`, `GIT_CURL_VERBOSE`, `GIT_TRACE_CURL`, ...); see [`git config`](config.md).

**Usage:** `git remote-http <remote> [<url>]`

## Flags

None. Arguments:

| Argument | What it is |
|----------|------------|
| `<remote>` | Remote name (e.g. `origin`), or the URL if unnamed. |
| `<url>` | URL to use. Default: `remote.<name>.url`. |
