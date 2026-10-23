# git upload-archive

Internal: server side of `git archive --remote`. Builds and sends the archive.

Clients may only request a ref (`v1.0`) or a path in one (`v1.0:docs`); `main^` or raw IDs are refused so unreachable objects stay private. `uploadArchive.allowUnreachable=true` allows any object.

**Usage:** `git upload-archive <repository>`

## Flags

None. `<repository>` is the repo to archive from.
