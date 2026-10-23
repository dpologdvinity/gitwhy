# git url-parse

Split git URLs (including scp-style `user@host:path`) into parts, or check they parse. Accepts any well-formed URL, even unfetchable schemes.

**Usage:** `git url-parse [-c <component>] [--] <url>...`

Without `-c`: no output; exit 0 only if every URL parses.

## Flags

| Flag | What it does |
|------|--------------|
| `-c`, `--[no-]component=<component>` | Print this part of each URL, one per line: `scheme`, `user`, `password`, `host`, `port`, or `path`. Missing part prints an empty line. |
