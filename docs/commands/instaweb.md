# git instaweb

Browse the current repo in a browser via a local gitweb server. Config: `instaweb.local`, `instaweb.httpd`, `instaweb.port`, `instaweb.browser`, `instaweb.modulePath`.

**Usage:** `git instaweb [--local] [--httpd=<httpd>] [--port=<port>] [--browser=<browser>] [--module-path=<path>]`
or `git instaweb (--start | --stop | --restart)`

## Flags

| Flag | What it does |
|------|--------------|
| `-l`, `--[no-]local` | Listen on 127.0.0.1 only. |
| `-d`, `--[no-]httpd=<httpd>` | Server command plus options; config file appended. `apache2`, `lighttpd` (default), `mongoose`, `plackup`, `python`, `webrick`. |
| `-m`, `--[no-]module-path=<path>` | Apache module folder. Default `/usr/lib/apache2/modules`. |
| `-p`, `--[no-]port=<port>` | Port. Default 1234. |
| `-b`, `--[no-]browser=<browser>` | Browser, via `git web--browse`. Default `instaweb.browser`, then `web.browser`. Prints URL on failure. |
| `--[no-]start`, `start` | Start server and exit. |
| `--[no-]stop`, `stop` | Stop server and exit. |
| `--[no-]restart`, `restart` | Restart server and exit. |
