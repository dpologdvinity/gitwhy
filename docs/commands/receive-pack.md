# git receive-pack

Internal: server side of a push. Receives objects, updates refs, runs server hooks. Started by `git push` (via `git send-pack`).

Honors `receive.*` config, e.g. `receive.denyNonFastForwards`.

**Usage:** `git receive-pack [<flags>] <git-dir>`

## Flags

| Flag | What it does |
|------|--------------|
| `<git-dir>` | Repo to push into. |
| `-q`, `--[no-]quiet` | Less output. |
| `--http-backend-info-refs` | For `git http-backend`'s `info/refs` reply. Old name: `--advertise-refs`. |
| `--stateless-rpc` | Internal: one round trip, for smart HTTP. |
| `--skip-connectivity-check` | Skip checking pushed history is complete. Can corrupt the repo unless checked elsewhere. |
| `--reject-thin-pack-for-testing` | Internal: git's tests. |

## Hooks

In `$GIT_DIR/hooks/`, executable. Stdin ref lines: `<old-id> <new-id> <refname>`. All-zero old ID = new ref; all-zero new ID = deleted.

| Hook | When | Input | Effect of failing |
|------|------|-------|-------------------|
| `pre-receive` | Once, before any ref changes. | One line per ref on stdin. | Non-zero rejects the whole push. |
| `update` | Per ref, before it changes. | Arguments: `<refname> <old-id> <new-id>`. | Non-zero rejects that ref. Passing doesn't guarantee the update. |
| `post-receive` | Once, after any successful update. | One line per updated ref on stdin. | Ignored. Good for notifications. |
| `post-update` | Last, if any ref changed. | Updated ref names as arguments. | Ignored. E.g. `git update-server-info`. |

By `post-receive`, the ref may have moved again; trust `<new-id>`.

## Signed pushes

For `git push --signed`, `pre-receive` and `post-receive` get:

| Variable | Meaning |
|----------|---------|
| `GIT_PUSH_CERT` | Blob ID of the push certificate. |
| `GIT_PUSH_CERT_SIGNER` | Key owner's name and email. |
| `GIT_PUSH_CERT_KEY` | GPG key ID. |
| `GIT_PUSH_CERT_STATUS` | GPG result, as `%G?` in `git log` (`G` = good). |
| `GIT_PUSH_CERT_NONCE` | Nonce the server requested. Mismatch may mean replay. |
| `GIT_PUSH_CERT_NONCE_STATUS` | `UNSOLICITED`, `MISSING`, `BAD`, `OK`, or `SLOP` (earlier session). |
| `GIT_PUSH_CERT_NONCE_SLOP` | With `SLOP`: seconds apart. See `receive.certNonceSlop`. |

## Quarantine

Pushed objects wait in a temp folder in `objects/` until `pre-receive` passes. So:

1. Failed pushes leave nothing on disk.
2. Objects `pre-receive` creates also wait.
3. `pre-receive` can't update refs.
