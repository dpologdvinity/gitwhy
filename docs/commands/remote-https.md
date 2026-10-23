# git remote-https

Internal helper git runs for `https://` remotes during fetch, clone, push, and ls-remote. Same program as [`git remote-http`](remote-http.md), [`git remote-ftp`](remote-ftp.md), [`git remote-ftps`](remote-ftps.md) (`git-remote-curl`). Speaks both "smart" and old "dumb" HTTP (see [`git http-fetch`](http-fetch.md)).

Not run by hand; speaks the [gitremote-helpers](https://git-scm.com/docs/gitremote-helpers) protocol on stdin. Configure via `http.*` config and env vars (`GIT_SSL_NO_VERIFY`, `GIT_CURL_VERBOSE`, `GIT_TRACE_CURL`, ...); see [`git config`](config.md).

**Usage:** `git remote-https <remote> [<url>]`

## Flags

None. Arguments:

| Argument | What it is |
|----------|------------|
| `<remote>` | Remote name (e.g. `origin`), or the URL if unnamed. |
| `<url>` | URL. Default: `remote.<name>.url`. |
