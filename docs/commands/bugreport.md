# git bugreport

Create a bug report file with blanks to fill in plus auto-collected info (`git version --build-options`, `uname`, compiler, enabled hooks, `$SHELL`). Send it to the git mailing list. May fail if git setup is broken (e.g. bad config).

**Usage:** `git bugreport [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `-o <path>`, `--[no-]output-directory <path>` | Write into `<path>`. Default: current folder. |
| `-s <format>`, `--[no-]suffix <format>` | File name suffix as a `strftime` time format (e.g. `%Y-%m-%d`). `--no-suffix`: plain `git-bugreport`. |
| `--[no-]diagnose[=<mode>]` | Also write a diagnostics zip. `<mode>`: `stats` (default) or `all`; see [`git diagnose`](diagnose.md). |
