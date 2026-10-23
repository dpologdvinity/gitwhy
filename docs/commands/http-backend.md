# git http-backend

Server side: CGI program a web server runs to serve repos over HTTP(S): fetch/clone (smart and dumb), push (smart), protocol v2. Not run by hand.

Serves only repos with a `git-daemon-export-ok` file unless `GIT_HTTP_EXPORT_ALL`. Fetch on by default; push only for authenticated users.

**Usage:** `git http-backend` (as a CGI script)

## Flags

None. Configured by environment and per-repo config.

## Environment

| Variable | What it does |
|----------|--------------|
| `GIT_PROJECT_ROOT` | Repos folder, joined with `PATH_INFO`. Unset: uses `PATH_TRANSLATED`. |
| `GIT_HTTP_EXPORT_ALL` | Serve every repo, skipping the export-ok check. |
| `GIT_HTTP_MAX_REQUEST_BUFFER` | Max fetch negotiation request (default `10M`). Same as `http.maxRequestBuffer`. |
| `GIT_PROTOCOL` | Client's `Git-Protocol` header (needed for v2). Auto-copied from `HTTP_GIT_PROTOCOL`. |
| `GIT_NAMESPACE` | Serve one namespace of a repo (see `gitnamespaces`). |
| `REMOTE_USER`, `REMOTE_ADDR` | From the server. Push reflog identity: `$REMOTE_USER@http.$REMOTE_ADDR`. |
| `PATH_INFO`, `PATH_TRANSLATED`, `QUERY_STRING`, `REQUEST_METHOD`, `CONTENT_TYPE` | Standard CGI variables. |

## Per-repo config

| Setting | What it does |
|---------|--------------|
| `http.uploadpack` | Serve fetch, clone, `ls-remote`. Default on. |
| `http.receivepack` | Accept push. Default: authenticated only. `true` = all, `false` = none. |
| `http.getanyfile` | Let pre-1.6.6 clients read any repo file, including unreachable objects. Default on. |
| `http.uploadarchive` | Serve `git archive --remote` (v2 only). Default off. |

Anonymous read, authenticated push: require auth on URLs ending `/git-receive-pack` and queries with `service=git-receive-pack`.
