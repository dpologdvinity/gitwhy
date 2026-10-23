# Revision options

Flags that pick which commits to list, in what order, and how each prints; shared by `log`, `rev-list`, `shortlog`, `show`, `format-patch` and friends.

Not every command takes every section. `rev-list` takes all of them. `log` takes all but [Object traversal](#object-traversal). `shortlog` takes [Commit limiting](#commit-limiting) and [History simplification](#history-simplification). `show` takes [Commit formatting](#commit-formatting). How each commit's changes print (`--stat`, `-p`, ...) is in [diff options](diff-options.md).

## Picking commits

Name commits on the command line; git walks back through their parents.

| You type | Means |
|----------|-------|
| `foo bar` | Commits reachable from `foo` or `bar`. |
| `^baz` | Leave out commits reachable from `baz`. |
| `A..B` | Reachable from `B` but not `A` (same as `^A B`). |
| `A...B` | Reachable from `A` or `B`, not both. |
| `[--] <path>...` | Only commits that changed these paths (see [History simplification](#history-simplification)). Use `--` if a path could be mistaken for a branch. |

## Commit limiting

Limits stack. They apply before ordering (`-n 3 --reverse` reverses the 3 newest).

| Flag | What it does |
|------|--------------|
| `-<n>`, `-n <n>`, `--max-count=<n>` | Show at most the first `<n>` commits. |
| `--max-count-oldest=<n>` | Show only the last `<n>` commits. |
| `--skip=<n>` | Skip the first `<n>` commits. |
| `--since=<date>`, `--after=<date>` | Only commits newer than `<date>` (`2.weeks.ago`, `2024-01-01`; `today` = last midnight). Stops at the first older commit. |
| `--since-as-filter=<date>` | Like `--since`, but checks every commit instead of stopping. |
| `--until=<date>`, `--before=<date>` | Only commits older than `<date>`. |
| `--max-age=<timestamp>` | `--since` with a Unix timestamp. |
| `--min-age=<timestamp>` | `--until` with a Unix timestamp. |
| `--author=<pattern>` | Author name/email matches regex. Repeat to match any. |
| `--committer=<pattern>` | Same, for committer. |
| `--grep=<pattern>` | Message matches regex. Repeat to match any (see `--all-match`). With `--notes`, notes count as message. |
| `--grep-reflog=<pattern>` | Reflog entry message matches. Needs `-g`. |
| `--all-match` | Must match every `--grep`. |
| `--invert-grep` | Message does not match `--grep`. |
| `-i`, `--regexp-ignore-case` | Ignore case in the patterns above. |
| `--basic-regexp` | Basic regexes (default). |
| `-E`, `--extended-regexp` | Extended regexes. |
| `-F`, `--fixed-strings` | Plain text, not regexes. |
| `-P`, `--perl-regexp` | Perl regexes (if git built with support). |
| `--remove-empty` | Stop when the given path stops existing. |
| `--merges` | Only merges (`--min-parents=2`). |
| `--no-merges` | Skip merges (`--max-parents=1`). |
| `--min-parents=<n>` | At least `<n>` parents. `3` = octopus merges. |
| `--max-parents=<n>` | At most `<n>` parents. `0` = root commits; `-1` = no limit. |
| `--no-min-parents`, `--no-max-parents` | Remove that parent-count limit. |
| `--first-parent` | Follow only first parents at merges: a branch's own history. |
| `--exclude-first-parent-only` | Follow only first parents when finding excluded (`^`) commits. |
| `--maximal-only` | Only tips: commits no other listed commit reaches. |
| `--not` | Flip `^` for every following revision until the next `--not`. |
| `--all` | As if every ref and `HEAD` were listed. |
| `--branches[=<pattern>]` | As if every local branch (matching glob) were listed. |
| `--tags[=<pattern>]` | Same, for tags. |
| `--remotes[=<pattern>]` | Same, for remote-tracking branches. |
| `--glob=<pattern>` | As if every ref matching glob were listed (`refs/` added if missing). |
| `--exclude=<pattern>` | Skip matching refs in the next `--all`, `--branches`, `--tags`, `--remotes` or `--glob`. Repeatable. |
| `--exclude-hidden=(fetch\|receive\|uploadpack)` | Skip refs that service hides (`*.hideRefs`) in the next `--all` or `--glob`. |
| `--reflog` | As if every commit in every reflog were listed. |
| `--alternate-refs` | As if alternate repositories' ref tips were listed. |
| `--single-worktree` | `--all`, `--reflog`, `--indexed-objects` use only the current worktree. |
| `--ignore-missing` | Silently skip nonexistent revision names. |
| `--bisect` | `log`: as if bisect "bad" and `--not` "good" refs were listed. `rev-list`: print only the commit halfway between. |
| `--stdin` | Also read revisions (and `--all`-style flags; paths after `--`) from stdin. |
| `--cherry-mark` | Like `--cherry-pick`, but keep duplicates, marked `=` (others `+`). |
| `--cherry-pick` | With `A...B`, drop commits whose change already exists on the other side. |
| `--left-only` | With `A...B`, only left-side (`A`) commits. |
| `--right-only` | With `A...B`, only right-side (`B`) commits. |
| `--cherry` | `--right-only --cherry-mark --no-merges`: your commits, marking ones already upstream (`git log --cherry upstream...mine`). |
| `-g`, `--walk-reflogs` | Walk reflog entries (newest first) instead of parents. No `^`, `..`, `...` or `--reverse`. Shows reflog name (`ref@{2}`). |
| `--merge` | During a conflicted merge, cherry-pick, revert or rebase: commits on either side touching conflicted files. |
| `--boundary` | Also show excluded commits at the range edge, marked `-`. |

## History simplification

With paths, git hides commits that didn't change them. These flags control how.
"TREESAME" = commit left the given paths identical to its parent.

| Flag | What it does |
|------|--------------|
| `--simplify-by-decoration` | Only commits a branch or tag points to (plus those changing given paths). Shows overall shape. |
| (default) | Simplest history explaining the final files. At a merge, follows only a TREESAME parent, so side branches can vanish. |
| `--show-pulls` | Default, plus merges that brought the change in from a side branch. |
| `--full-history` | Follow every parent of every merge. Finds hidden commits; adds noisy merges. |
| `--dense` | Only commits that changed the paths (plus connectors). Default. |
| `--sparse` | Every commit walked, even unchanged ones. |
| `--simplify-merges` | With `--full-history`, drop merges that added nothing. Walks all history before printing. |
| `--ancestry-path[=<commit>]` | In `A..B`, only ancestors or descendants of `<commit>` (default `A`), or it. Repeatable. |

## Commit ordering

Default: newest first by commit date.

| Flag | What it does |
|------|--------------|
| `--date-order` | No parent before all its children; otherwise by commit date. |
| `--author-date-order` | Same, by author date. |
| `--topo-order` | No parent before its children; don't interleave parallel branches. |
| `--reverse` | Opposite order (after limiting). Not with `-g`. |
| `--no-walk[=(sorted\|unsorted)]` | Only the named commits, not ancestors. `unsorted` keeps command-line order; `sorted` (default) by date. Ignored for ranges. Not with `--graph`. |
| `--do-walk` | Undo `--no-walk`. |

## Object traversal

Mostly for packing; used by `rev-list`.

| Flag | What it does |
|------|--------------|
| `--objects` | Also list every tree and blob the commits use, with paths. |
| `--in-commit-order` | Print trees and blobs right after the commit that first uses them. |
| `--objects-edge` | `--objects`, plus excluded commits marked `-` (for thin packs). |
| `--objects-edge-aggressive` | `--objects-edge`, trying harder (for shallow repos). Slower. |
| `--indexed-objects` | As if every tree and blob in the index were listed. Pair with `--objects`. |
| `--unpacked` | With `--objects`: only unpacked objects. |
| `--[no-]object-names` | With `--objects`: print path/name next to each ID (default on). `--no-object-names` suits `git cat-file`. |
| `--filter=<filter>` | With `--objects`: omit objects. `blob:none` (all files), `blob:limit=<n>[kmg]` (files that big or bigger), `object:type=<type>` (all but that type), `tree:<depth>` (that deep or deeper), `sparse:oid=<blob>` (outside that sparse spec), `combine:<f1>+<f2>` (all). Repeat to combine. |
| `--no-filter` | Cancel earlier `--filter`. |
| `--filter-provided-objects` | Filter directly named objects too (normally always printed). |
| `--filter-print-omitted` | Also print filtered-out objects, prefixed `~`. |
| `--missing=<action>` | Missing objects: `error` (default), `allow-any` (skip), `allow-promisor` (skip only ones a partial-clone remote should have), `print` (list with `?`), `print-info` (plus path and type). |
| `--missing-only` | With `--missing=print` or `print-info`: print only missing objects, no `?`. Not with `--count` or `--disk-usage`. (2.56+) |
| `--exclude-promisor-objects` | Internal (partial clone): don't walk past promisor-remote objects. |

## Commit formatting

| Flag | What it does |
|------|--------------|
| `--pretty[=<format>]`, `--format=<format>` | Print each commit in a [format](#formats) (default `medium`). Anything with `%` counts as `tformat:`. |
| `--abbrev-commit` | Short commit IDs. Length via `--abbrev=<n>`. |
| `--no-abbrev-commit` | Full commit IDs, even with `--oneline` or `log.abbrevCommit`. |
| `--oneline` | `--pretty=oneline --abbrev-commit`. |
| `--encoding=<encoding>` | Convert messages to this encoding (default UTF-8). |
| `--expand-tabs[=<n>]`, `--no-expand-tabs` | Expand tabs in messages to every `<n>`th column (default 8; `--no-expand-tabs` = 0). On by default for `medium`, `full`, `fuller`. |
| `--notes[=<ref>]` | Show notes (`git notes`). Default without `--pretty`/`--format`/`--oneline`. `<ref>` picks notes ref (`foo` = `refs/notes/foo`). Repeatable. |
| `--no-notes` | Hide notes. Clears earlier `--notes` refs; later `--notes=<ref>` still count. |
| `--show-notes-by-default` | Show default notes unless specific notes refs are asked for. |
| `--show-notes[=<ref>]`, `--[no-]standard-notes` | Deprecated. Use `--notes` / `--no-notes`. |
| `--show-signature` | Verify signed commits with gpg; show result. |
| `--relative-date` | `--date=relative`. |
| `--date=<format>` | Date style. See [date formats](#date-formats). |
| `--parents` | Print parent IDs after each commit. Turns on parent rewriting. |
| `--children` | Print child IDs after each commit. Turns on parent rewriting. |
| `--left-right` | With `A...B`, mark commits `<` (`A`) or `>` (`B`); boundary `-`. |
| `--graph` | ASCII branch/merge graph on the left. Implies `--topo-order` (`--date-order` allowed). Not with `--no-walk`. |
| `--show-linear-break[=<barrier>]` | Without `--graph`, print a separator (or `<barrier>`) between adjacent non-parent/child commits. |
| `--graph-lane-limit=<n>` | With `--graph`, at most `<n>` lanes; extras become `~`. `0` = no limit. |
| `--[no-]graph-indent` | With `--graph`, indent commits with no shown parents so they don't look connected. On by default; set by `log.graphIndent`. (2.56+) |

### Formats

Merge commits also get a `Merge: <parents>` line, except in `oneline`, `email`, `raw`.

| Format | Output |
|--------|--------|
| `oneline` | `<hash> <subject>` |
| `short` | `commit`, `Author:`, subject. |
| `medium` | `commit`, `Author:`, `Date:`, full message. Default. |
| `full` | `commit`, `Author:`, `Commit:`, full message. |
| `fuller` | `commit`, `Author:`, `AuthorDate:`, `Commit:`, `CommitDate:`, full message. |
| `reference` | `<short hash> (<subject>, <short date>)`, for citing commits. |
| `email` | Email headers (`From`, `Date`, `Subject: [PATCH]`) then message. |
| `mboxrd` | `email`, but message lines starting `From ` get `>` so mail tools don't split. |
| `raw` | Commit object as stored (full hashes, real parents). |
| `format:<string>` | Custom layout with [placeholders](#placeholders). Newline between commits. |
| `tformat:<string>` | `format:`, but newline after every commit, including the last. |
| `<name>` | Format defined in `pretty.<name>` config. |

### Placeholders

| Placeholder | Expands to |
|-------------|------------|
| `%n` | Newline. |
| `%%` | Literal `%`. |
| `%x<hh>` | Byte with hex value `<hh>` (`%x00` = NUL). |
| `%Cred`, `%Cgreen`, `%Cblue`, `%Creset` | Switch / reset color. |
| `%C(<spec>)` | Color from config-style spec (`%C(bold yellow)`). Only when color is on; `%C(always,<spec>)` forces; `%C(auto)` auto-colors following placeholders. |
| `%m` | `<`, `>` or `-` side mark (see `--left-right`). |
| `%w([<w>[,<i1>[,<i2>]]])` | Wrap at `<w>` columns, indent first line `<i1>`, rest `<i2>`. |
| `%<(<n>[,trunc\|ltrunc\|mtrunc])` | Right-pad next placeholder to `<n>` columns; optionally cut with `..` at end, start or middle. |
| `%<\|(<m>)` | Right-pad next placeholder to column `<m>` (negative = from right edge). |
| `%>(<n>)`, `%>\|(<m>)` | Same, left-pad. |
| `%>>(<n>)`, `%>>\|(<m>)` | Left-pad, borrowing space on the left if too long. |
| `%><(<n>)`, `%><\|(<m>)` | Center. |
| `%H`, `%h` | Commit hash, full / short. |
| `%T`, `%t` | Tree hash, full / short. |
| `%P`, `%p` | Parent hashes, full / short. |
| `%an`, `%aN` | Author name (`N` = after `.mailmap`). |
| `%ae`, `%aE` | Author email (`E` = after `.mailmap`). |
| `%al`, `%aL` | Author email before `@` (`L` = after `.mailmap`). |
| `%ad` | Author date, `--date` format. |
| `%aD`, `%ar`, `%at`, `%ai`, `%aI`, `%as`, `%ah` | Author date: RFC 2822, relative, Unix timestamp, ISO-like, strict ISO, `YYYY-MM-DD`, human. |
| `%cn`, `%cN`, `%ce`, `%cE`, `%cl`, `%cL` | Committer name / email / email local part (capital = after `.mailmap`). |
| `%cd`, `%cD`, `%cr`, `%ct`, `%ci`, `%cI`, `%cs`, `%ch` | Committer date, same variants as author. |
| `%d` | Ref names like `--decorate`: ` (HEAD -> main, tag: v1)`. |
| `%D` | Ref names without ` (` `)`. |
| `%(decorate[:<opts>])` | Ref names with custom `prefix=`, `suffix=`, `separator=`, `pointer=`, `tag=`. Write commas as `%x2C`, `)` as `%x29`. |
| `%(describe[:<opts>])` | `git describe` name; empty if none. Options `tags`, `abbrev=<n>`, `match=<glob>`, `exclude=<glob>`. |
| `%(count)`, `%(total)` | Patch number / count in a series. Only in `format-patch --commit-list-format`. |
| `%S` | Command-line ref that reached this commit (`git log` only). |
| `%e` | Message encoding. |
| `%s` | Subject (first line). |
| `%f` | Subject as a filename. |
| `%b` | Body. |
| `%B` | Raw message (subject and body, unwrapped). |
| `%GG` | Raw gpg output for a signed commit. |
| `%G?` | Signature status: `G` good, `B` bad, `U` good but unknown trust, `X` expired signature, `Y` expired key, `R` revoked key, `E` can't check, `N` none. |
| `%GS`, `%GK`, `%GF`, `%GP`, `%GT` | Signer name, key, key fingerprint, primary key fingerprint, trust level. |
| `%gD`, `%gd` | Reflog selector (`refs/stash@{1}`), full / short. Only with `-g`. |
| `%gn`, `%gN`, `%ge`, `%gE` | Reflog identity name / email (capital = after `.mailmap`). |
| `%gs` | Reflog entry message. |
| `%(trailers[:<opts>])` | Message trailers (`Signed-off-by:` ...). Options: `key=<key>`, `only`, `separator=<sep>`, `unfold`, `keyonly`, `valueonly`, `key_value_separator=<sep>`. |

Modifiers after `%`: `%+x` adds a newline before it if non-empty; `%-x` removes newlines before it if empty; `% x` adds a space before it if non-empty.

### Date formats

Dates use the commit's time zone. Add `-local` (`iso-local`) for yours. Default from `log.date` config.

| `--date=` | Output |
|-----------|--------|
| `default` | `Thu Jan 1 00:00:00 1970 +0000` |
| `local` | `default-local`. |
| `relative` | `2 hours ago`. |
| `iso`, `iso8601` | `1970-01-01 00:00:00 +0000` |
| `iso-strict`, `iso8601-strict` | `1970-01-01T00:00:00Z` (`Z` = UTC; else `+01:00` style). |
| `rfc`, `rfc2822` | `Thu, 1 Jan 1970 00:00:00 +0000` (email style). |
| `short` | `1970-01-01` |
| `raw` | `0 +0000` (Unix seconds and zone). |
| `unix` | `0` (Unix seconds). |
| `human` | Drops parts matching today (year, or whole date for recent days). |
| `format:<strftime>` | Custom `strftime` layout (`format:%Y-%m`). `format-local:` for your zone. |
