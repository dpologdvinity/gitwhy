# git tag

Create, list, delete, or verify tags (fixed names for commits, usually releases). Lightweight tags are just names. Annotated tags (`-a`, `-s`, `-u`, or a message flag) also store tagger, date, message, and optional signature; use them for releases. Tags aren't pushed unless you push them (`git push origin <tag>` or `--tags`).

Don't move a tag others fetched; their copy won't update. Make a new one.

**Usage:** `git tag [-a | -s | -u <key-id>] [-f] [-m <msg> | -F <file>] [-e] [--trailer <token>[(=|:)<value>]]... <tagname> [<commit> | <object>]`
or `git tag -d <tagname>...`
or `git tag [-n[<num>]] -l [<listing flags>] [<pattern>...]`
or `git tag -v [--format=<format>] <tagname>...`

No arguments: list all tags.

## Flags

| Flag | What it does |
|------|--------------|
| `<tagname>` | Tag name. Must be a valid ref name (see `git check-ref-format`). |
| `<commit>`, `<object>` | What the tag points to. Default `HEAD`. |
| `-l`, `--list` | List tags, optionally matching wildcard `<pattern>`s (e.g. `'v1.*'`). |
| `-d`, `--delete` | Delete tags locally. |
| `-v`, `--verify` | Verify tag signatures. Takes `--format`. |
| `-n[<num>]` | List with `<num>` message lines (default 1). Implies `-l`. |

### Creating

| Flag | What it does |
|------|--------------|
| `-a`, `--[no-]annotate` | Annotated, unsigned tag. Opens the editor unless `-m`/`-F`. |
| `-s`, `--[no-]sign` | Signed tag with your default key (`gpg.format`: OpenPGP, SSH, X.509). `--no-` overrides `tag.gpgSign`. |
| `-u`, `--[no-]local-user=<key-id>` | Signed tag with this key. |
| `-m`, `--message=<msg>` | Message. Repeat for paragraphs. Implies `-a`. |
| `-F`, `--[no-]file=<file>` | Message from a file (`-` = stdin). Implies `-a`. |
| `--[no-]trailer <token>[(=\|:)<value>]` | Add a trailer like `Reviewed-by: X`. Repeatable. Implies `-a`. |
| `-e`, `--[no-]edit` | Edit a `-m`/`-F` message. |
| `--[no-]cleanup=<mode>` | `strip` (default: extra blanks and `#` comments), `whitespace`, or `verbatim`. |
| `-f`, `--[no-]force` | Replace an existing tag. |
| `--[no-]create-reflog` | Keep a reflog for this tag. `--no-` doesn't override `core.logAllRefUpdates`. |

### Listing

| Flag | What it does |
|------|--------------|
| `--[no-]sort=<key>` | Sort by a [for-each-ref field](for-each-ref.md#format-fields); `-` for descending; repeatable (last is primary). `version:refname` sorts as versions. Default: `tag.sort`, else alphabetical. |
| `--[no-]format=<format>` | Output template ([fields](for-each-ref.md#format-fields)). Default `%(refname:strip=2)`. |
| `--[no-]color[=<when>]` | Colors from `--format`: `always` (bare), `never`, `auto`. |
| `-i`, `--[no-]ignore-case` | Case-insensitive sort and match. |
| `--[no-]omit-empty` | No newline for empty formatted lines. |
| `--[no-]column[=<options>]` | Show in columns (see `column.tag`). Not with `-n`. |
| `--contains [<commit>]` | Tags containing `<commit>` (default `HEAD`). |
| `--no-contains [<commit>]` | Tags not containing `<commit>` (default `HEAD`). |
| `--merged [<commit>]` | Tags reachable from `<commit>` (default `HEAD`). |
| `--no-merged [<commit>]` | Tags not reachable from `<commit>` (default `HEAD`). |
| `--[no-]points-at [<object>]` | Tags pointing at `<object>` (default `HEAD`). |
