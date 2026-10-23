# git config

Read and change settings. Names are `<section>.<key>` (`user.name`) or `<section>.<subsection>.<key>` (`remote.origin.url`). `git help --config` lists all.

Reads system, global, local, and worktree files; last found wins. Writes local unless a location flag says otherwise.

Exit codes: 1 bad section/key (or `get` found nothing), 2 no section or name, 3 invalid config file, 4 can't write, 5 unsetting a missing setting, or set/unset hitting multiple lines, 6 bad regex.

**Usage:** `git config <subcommand> [<flags>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `list` | Show all settings. |
| `get <name>` | Print a value (last if multiple). Exit 1 if unset. |
| `set <name> <value>` | Set a value. Multi-valued settings need `--all`, `--value`, or `--append`. |
| `unset <name>` | Remove a setting. Multi-valued settings need `--all` or `--value`. |
| `rename-section <old-name> <new-name>` | Rename a section (`remote.old` to `remote.new`). |
| `remove-section <name>` | Delete a section. |
| `edit` | Open the file in your editor. |

## Flags for every subcommand

Choose the file.

| Flag | What it does |
|------|--------------|
| `--[no-]global` | `~/.gitconfig` (or `$XDG_CONFIG_HOME/git/config` if it exists and `~/.gitconfig` doesn't). Reads both. |
| `--[no-]system` | Machine-wide `$(prefix)/etc/gitconfig`. |
| `--[no-]local` | Repo's `.git/config`. Default for writes. |
| `--[no-]worktree` | `config.worktree` if `extensions.worktreeConfig`, else `--local`. |
| `-f`, `--[no-]file <config-file>` | This file. |
| `--[no-]blob <blob>` | A file in git, e.g. `main:.gitmodules`. Read-only. |

### git config list

| Flag | What it does |
|------|--------------|
| `--[no-]includes` | Follow `include.*` / `includeIf.*`. On by default, off with a location flag. |
| `-z`, `--[no-]null` | NUL after each entry, newline between name and value. Safe for multi-line values. |
| `--[no-]name-only` | Names only. |
| `--[no-]show-origin` | Show source: `file:<path>`, `standard input`, `blob:<id>`, `command line`. |
| `--[no-]show-scope` | Show scope: `worktree`, `local`, `global`, `system`, `command`. |
| `--[no-]show-names` | Show names with values. |
| `-t`, `--[no-]type <type>` | Show values in canonical form for a type; skip non-fitting. [Types](#types). |
| `--bool`, `--int`, `--bool-or-int`, `--bool-or-str`, `--path`, `--expiry-date` | Old `--type=<type>`. |

### git config get

| Flag | What it does |
|------|--------------|
| `--[no-]all` | Print every value. |
| `--[no-]regexp` | `<name>` is a regex; print all matches. Case-sensitive against lowercased section and key. |
| `--[no-]value=<pattern>` | Only values matching regex; `!<pattern>` inverts. `--no-value` clears. |
| `--[no-]fixed-value` | `--value` is exact text. |
| `--[no-]url=<URL>` | For `<section>.<key>`, return `<section>.<url>.<key>` for the best-matching URL, else `<section>.<key>`. With only `<section>`, list all its keys so. |
| `--[no-]default <value>` | Print `<value>` if not found. |
| `--[no-]includes` | Follow includes (see `list`). |
| `-z`, `--[no-]null` | NUL-terminated. |
| `--[no-]name-only` | Names only. |
| `--[no-]show-origin` | Show source. |
| `--[no-]show-scope` | Show scope. |
| `--[no-]show-names` | Show names. Off by default, except `--url` without subsection. |
| `-t`, `--[no-]type <type>` | Check and print as this type. |
| `--bool`, `--int`, `--bool-or-int`, `--bool-or-str`, `--path`, `--expiry-date` | Old `--type=<type>`. |

### git config set

| Flag | What it does |
|------|--------------|
| `--[no-]all` | Replace all lines of a multi-valued setting with one. |
| `--[no-]value=<pattern>` | Replace only lines with values matching regex (`!` inverts). |
| `--[no-]fixed-value` | `--value` is exact text. |
| `--[no-]append` | Add a line, keep existing. Same as `--value=^$`. |
| `--[no-]comment <message>` | Append `# <message>` to the written line. One line only. |
| `-t`, `--[no-]type <type>` | Validate and store in canonical form. |
| `--bool`, `--int`, `--bool-or-int`, `--bool-or-str`, `--path`, `--expiry-date` | Old `--type=<type>`. |

### git config unset

| Flag | What it does |
|------|--------------|
| `--[no-]all` | Remove all lines of a multi-valued setting. |
| `--[no-]value=<pattern>` | Remove only lines with values matching regex (`!` inverts). |
| `--[no-]fixed-value` | `--value` is exact text. |

### git config rename-section, remove-section, edit

Only the file location flags.

## Types

For `--type=<type>`. `--no-type` cancels.

| Type | Meaning |
|------|---------|
| `bool` | `true`/`yes`/`on`/nonzero number = `true`; `false`/`no`/`off`/`0` = `false`. |
| `int` | Decimal; `k`, `m`, `g` multiply by 1024, 1024², 1024³. |
| `bool-or-int` | Either. |
| `bool-or-str` | Boolean, else any text. |
| `path` | Expand leading `~` / `~user`. Reading only. |
| `expiry-date` | Date like `2.weeks.ago` to timestamp. Reading only. |
| `color` | Reading: name to terminal escape code. Writing: validate. |

## Other flags

| Flag | What it does |
|------|--------------|
| `--get-colorbool <name> [<stdout-is-tty>]` | Print `true`/`false`: use color for `<name>` (e.g. `color.diff`, else `color.ui`)? `<stdout-is-tty>` resolves `auto`; without it, checks own output and exits 0 for color, 1 for none. |
| `--replace-all` | Old mode: replace all lines matching name (and `<value-pattern>`): `git config --replace-all <name> <value> [<value-pattern>]`. Use `set --all`. |

## Old modes

Still work; prefer subcommands.

| Old form | Use instead |
|----------|-------------|
| `git config <name>` | `git config get <name>` |
| `git config <name> <value> [<value-pattern>]` | `git config set [--value=<pattern>] <name> <value>` |
| `-l`, `--list` | `git config list` |
| `--get <name> [<value-pattern>]` | `git config get [--value=<pattern>] <name>` |
| `--get-all <name> [<value-pattern>]` | `git config get --all [--value=<pattern>] <name>` |
| `--get-regexp <name-regexp> [<value-pattern>]` | `git config get --all --show-names --regexp <name-regexp>` |
| `--get-urlmatch <name> <URL>` | `git config get --url=<URL> <name>` |
| `--get-color <name> [<default>]` | `git config get --type=color [--default=<default>] <name>` |
| `--add <name> <value>` | `git config set --append <name> <value>` |
| `--unset <name> [<value-pattern>]` | `git config unset [--value=<pattern>] <name>` |
| `--unset-all <name> [<value-pattern>]` | `git config unset --all [--value=<pattern>] <name>` |
| `--rename-section <old-name> <new-name>` | `git config rename-section <old-name> <new-name>` |
| `--remove-section <name>` | `git config remove-section <name>` |
| `-e`, `--edit` | `git config edit` |

Old modes also take location, display (`-z`, `--name-only`, `--show-origin`, `--show-scope`, `--show-names`), type, `--default`, `--comment`, `--fixed-value`, and `--includes` flags.
