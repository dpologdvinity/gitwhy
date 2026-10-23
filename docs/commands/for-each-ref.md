# git for-each-ref

List refs with custom format, sorting, and filters. Scripting tool behind `git branch` and `git tag` listings.

**Usage:** `git for-each-ref [<flags>] [--stdin | <pattern>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<pattern>...` | Only refs matching a glob (`refs/heads/feat*`) or slash-ended prefix (`refs/tags`). |
| `--[no-]stdin` | Read patterns from stdin. |
| `--[no-]count=<n>` | Stop after `<n>` refs. |
| `--[no-]sort=<key>` | Sort by field (see [Format fields](#format-fields)); `-` prefix descends. Repeatable; last is primary. Default `refname`. `version:refname` (`v:refname`) sorts versions. |
| `--[no-]format=<format>` | Template with `%(field)`. `%%` is `%`, `%xx` a hex byte (`%00` NUL, `%09` tab, `%0a` newline). Default `%(objectname) %(objecttype)<TAB>%(refname)`. |
| `--[no-]color[=<when>]` | Apply `%(color:...)`: `always` (bare default), `never`, `auto`. |
| `-s`, `--[no-]shell` | Quote fields for shell `eval`. |
| `-p`, `--[no-]perl` | Quote for Perl. |
| `--[no-]python` | Quote for Python. |
| `--[no-]tcl` | Quote for Tcl. |
| `--[no-]points-at=<object>` | Only refs pointing at this object. |
| `--merged[=<commit>]` | Only refs merged into `<commit>` (default `HEAD`). |
| `--no-merged[=<commit>]` | Only refs not merged into `<commit>` (default `HEAD`). |
| `--contains[=<commit>]` | Only refs containing `<commit>` (default `HEAD`). |
| `--no-contains[=<commit>]` | Only refs not containing `<commit>` (default `HEAD`). |
| `--[no-]ignore-case` | Case-insensitive sort and match. |
| `--[no-]omit-empty` | No newline for empty format output. |
| `--[no-]exclude=<pattern>` | Hide matching refs. Repeatable. |
| `--[no-]include-root-refs` | Also list `HEAD` and pseudorefs (`FETCH_HEAD`, `ORIG_HEAD`, ...). |
| `--[no-]start-after=<marker>` | Paging: start after `<marker>` alphabetically. Not with `--sort`, `--stdin`, or `<pattern>`. |

Multiple `--contains`/`--no-contains`: ref must contain at least one `--contains` commit and no `--no-contains` commit. Same for `--merged`/`--no-merged`.

## Format fields

For `%(field)` in `--format` and `--sort`. Inapplicable fields are empty. Also used by `git branch`, `git tag`, `git verify-tag`.

| Field | What it gives |
|-------|---------------|
| `refname` | Full ref name. `:short` shortest unambiguous. `:lstrip=<n>` / `:rstrip=<n>` drop `<n>` parts from front/back (negative: keep that many). `strip` = `lstrip`. |
| `objecttype` | `blob`, `tree`, `commit`, `tag`. |
| `objectsize` | Size in bytes. `:disk`: on-disk size. |
| `objectname` | Object ID. `:short`, `:short=<length>`. |
| `deltabase` | Delta base object ID, or zeros. |
| `upstream` | Upstream branch. `:short`, `:lstrip`, `:rstrip`; `:track` gives `[ahead N, behind M]` or `[gone]`; `:track,nobracket`; `:trackshort` gives `>`, `<`, `<>`, `=`; `:remotename`, `:remoteref`. |
| `push` | `@{push}` branch. Same modifiers as `upstream`. |
| `HEAD` | `*` if checked out, else a space. |
| `color:<color>` | Switch color, e.g. `%(color:bold red)`. |
| `align:<width>,<position>` ... `%(end)` | Pad text to `<width>`: `left` (default), `middle`, `right`. Also `width=`/`position=` form. |
| `if` ... `then` ... [`else` ...] `end` | Show `then` if `if` part isn't empty (ignoring spaces). `%(if:equals=<s>)`, `%(if:notequals=<s>)` compare instead. |
| `symref` | Symbolic ref target. `:short`, `:lstrip`, `:rstrip`. |
| `signature` | GPG signature. `:grade` (`G` good, `B` bad, `U` unknown validity, `X` expired, `Y` expired key, `R` revoked key, `E` can't check, `N` none), `:signer`, `:key`, `:fingerprint`, `:primarykeyfingerprint`, `:trustlevel`. |
| `worktreepath` | Worktree where the branch is checked out, else empty. |
| `ahead-behind:<commit-ish>` | Commits ahead and behind `<commit-ish>`. |
| `is-base:<commit-ish>` | `(<commit-ish>)` on the ref most likely to be `<commit-ish>`'s branch starting point. |
| `describe[:<options>]` | `git describe` name. Options: `tags=<bool>`, `abbrev=<n>`, `match=<pattern>`, `exclude=<pattern>`. |
| `tree`, `parent`, `object`, `type`, `tag` | Commit/tag header fields. `tree`/`parent` take `:short`. |
| `author`, `committer`, `tagger` | Name, email, date. Suffix `name`, `email`, `date` (e.g. `authorname`). Email takes `:trim` (no `<>`), `:localpart`, `:mailmap` (combinable). |
| `creator`, `creatordate` | Committer or tagger, whichever applies. |
| `<date-field>:<format>` | Date in a `--date` format, e.g. `%(committerdate:short)`. Sorts as text. |
| `raw`, `raw:size` | Raw object data, its size. Not with `--shell`/`--perl`/`--python`/`--tcl`. |
| `contents` | Whole message. `:size`, `:subject` (first paragraph; also `subject`, `subject:sanitize` for file-safe), `:body`, `:signature`, `:lines=<n>`. |
| `trailers[:<options>]` | Message trailers. Options as in `git log`'s `%(trailers)`. Old name `contents:trailers`. |
| `*<field>` | For tags: field of the tagged object, e.g. `%(*authordate)`. |

Numeric fields (`objectsize`, dates) sort numerically; others by bytes.
