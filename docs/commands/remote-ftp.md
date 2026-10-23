# git remote-ftp

Internal helper git runs for `ftp://` remotes (fetch, clone, `ls-remote` only). Same program as [`git remote-http`](remote-http.md), [`git remote-https`](remote-https.md), [`git remote-ftps`](remote-ftps.md). Slow and deprecated; use `https://` or `ssh://`.

Speaks the [remote-helper protocol](https://git-scm.com/docs/gitremote-helpers) on stdin. Configure with `http.*` and env vars like `GIT_SSL_NO_VERIFY`, `GIT_CURL_VERBOSE`, `GIT_TRACE_CURL` (see [`git config`](config.md)).

**Usage:** `git remote-ftp <remote> [<url>]`

## Flags

None. Arguments:

| Argument | What it is |
|----------|------------|
| `<remote>` | Remote name, or the URL if unnamed. |
| `<url>` | URL. Default: `remote.<name>.url`. |
