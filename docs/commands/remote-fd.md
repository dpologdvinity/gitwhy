# git remote-fd

Internal: handles `fd::<infd>[,<outfd>][/<anything>]` URLs, talking to a git server over already-open file descriptors. For programs that set up the connection themselves.

- One number: a two-way socket.
- Two numbers: input pipe, output pipe.
- `/<anything>`: ignored label.

**Usage:** `git fetch fd::17 main`, `git push fd::7,8 main`

## Flags

None.

## Environment variables

| Variable | What it does |
|----------|--------------|
| `GIT_TRANSLOOP_DEBUG` | Print debug info for each read and write. |
