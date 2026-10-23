# git help

Show help. No args: common commands. With a command or guide: open its manual page (`man` by default). With an alias: print its expansion. `git <command> --help` is the same.

**Usage:** `git help [-a] [--[no-]verbose] [--[no-]external-commands] [--[no-]aliases]`
or `git help [-i | -m | -w] [<command> | <guide>]`
or `git help (-g | -c | --user-interfaces | --developer-interfaces)`

## Flags

| Flag | What it does |
|------|--------------|
| `-a`, `--all` | List all commands. |
| `--[no-]external-commands` | With `--all`: include `git-*` programs on `$PATH`. On by default. |
| `--[no-]aliases` | With `--all`: include aliases. On by default. |
| `-v`, `--[no-]verbose` | With `--all`: one-line description per command. On by default. |
| `-g`, `--guides` | List concept guides (e.g. `tutorial`, `revisions`). |
| `-c`, `--config` | List all config variable names. |
| `--user-interfaces` | List docs on repo files and conventions (`.gitignore`, `.mailmap`, hooks, ...). |
| `--developer-interfaces` | List docs on file formats, protocols, internals. |
| `-m`, `--[no-]man` | Use `man` (or `man.viewer`). Overrides `help.format`. |
| `-i`, `--[no-]info` | Use `info`. |
| `-w`, `--[no-]web` | Open HTML page in a browser (`help.browser`, then `web.browser`, else guess). |
