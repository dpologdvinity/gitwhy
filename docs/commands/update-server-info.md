# git update-server-info

Write `info/refs` and `objects/info/packs` so "dumb" HTTP/FTP servers (plain file serving) can be cloned from. Run after each push (the sample `post-update` hook does).

**Usage:** `git update-server-info [-f]`

## Flags

| Flag | What it does |
|------|--------------|
| `-f`, `--[no-]force` | Rebuild from scratch. |
