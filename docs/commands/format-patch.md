# git format-patch

Turn commits into email-ready patch files, one per commit (`0001-fix-typo.patch`, ...). Send with `git send-email`, apply with `git am`. Skips merges.

**Usage:** `git format-patch [<flags>] [<since> | <revision-range>]`

`git format-patch origin/main` = your commits not on `origin/main`. `-3` = last 3. `-1 <commit>` = just that one. `--root <commit>` = all history up to it. Each file: headers, message, `---`, diffstat, patch.

## Flags

Also takes all [diff options](../shared/diff-options.md) (`-M`, `-U<n>`, `--stat`, ...) and the commit limiting [revision options](../shared/revision-options.md#commit-limiting).

### Choosing commits and files

| Flag | What it does |
|------|--------------|
| `-<n>` | Newest `<n>` commits. |
| `--root` | Single commit means "up to here", not "since here". |
| `--[no-]ignore-if-in-upstream` | Skip commits already upstream. |
| `--always` | Include empty commits. |
| `-o <dir>`, `--output-directory=<dir>` | Write into `<dir>` (created if needed). |
| `--[no-]stdout` | Print all as one mailbox to stdout. |
| `--[no-]numbered-files` | Name files `1`, `2`, ... only. |
| `--[no-]start-number=<n>` | Start numbering at `<n>`. |
| `--[no-]suffix=.<sfx>` | Extension instead of `.patch` (empty = none). |
| `--[no-]filename-max-length=<n>` | Truncate names near `<n>` bytes (default 64). |
| `-q`, `--[no-]quiet` | Don't print file names. |
| `--[no-]progress` | Progress on stderr. |

### Subject line

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]numbered` | `[PATCH n/m]` even for one patch. |
| `-N`, `--no-numbered` | Plain `[PATCH]` even for many. |
| `-k`, `--keep-subject` | Don't strip or add `[PATCH]`. |
| `--subject-prefix=<prefix>` | `[<prefix>]` instead of `[PATCH]`. |
| `--[no-]rfc[=<rfc>]` | Put `<rfc>` (default `RFC`) before `PATCH`. Leading `-` puts it after: `--rfc=-(WIP)` gives `PATCH (WIP)`. |
| `-v <n>`, `--[no-]reroll-count=<n>` | Version `<n>` of the series: `[PATCH v<n>]`, `v<n>-` file prefix. |

### Message content

| Flag | What it does |
|------|--------------|
| `-p`, `--no-stat` | No diffstat. |
| `-s`, `--[no-]signoff` | Add `Signed-off-by:`. |
| `--notes[=<ref>]`, `--no-notes` | Add notes after `---`. Default off unless `format.notes`. |
| `--[no-]signature=<signature>` | Signature after `-- ` (default: git version). |
| `--[no-]signature-file=<file>` | Signature from a file. |
| `--binary` | Full binary diffs (default). |
| `--no-binary` | Only note binary changes. Unappliable, fine for review. |
| `--[no-]zero-commit` | All-zero hash in `From` lines. |
| `--[no-]base[=<commit>]` | Record the series' base commit (`auto` = guess). `--no-base` overrides `format.useAutoBase`. |
| `--[no-]interdiff=<previous>` | Add a diff against the previous series tip to the cover letter (or lone patch). |
| `--[no-]range-diff=<previous>` | Add a `git range-diff` against the previous version (tip or range) to the cover letter (or lone patch). |
| `--[no-]creation-factor=<percent>` | Tune `--range-diff` commit matching (default 999). |

### Cover letter

| Flag | What it does |
|------|--------------|
| `--[no-]cover-letter` | Also write `0000-cover-letter.patch` with commit list and diffstat, to fill in. |
| `--[no-]cover-from-description=<mode>` | Fill from branch description: `message`/`default` (body), `subject` (first paragraph = subject), `auto` (`subject` if first paragraph ≤ 100 bytes), `none`. |
| `--[no-]description-file=<file>` | Use this file instead of the branch description. |
| `--[no-]commit-list-format=<format>` | Commit list: `shortlog` (default), `modern` (`[n/m] subject`), or `log:<format>` with [placeholders](../shared/revision-options.md#placeholders). Implies `--cover-letter`. |

### Email headers

| Flag | What it does |
|------|--------------|
| `--[no-]to=<email>` | Add `To:`. Repeatable; `--no-to` clears. |
| `--[no-]cc=<email>` | Add `Cc:`. Repeatable; `--no-cc` clears. |
| `--[no-]add-header=<header>` | Add any header (`"Organization: foo"`). Repeatable; `--no-add-header` clears To, Cc, custom. |
| `--[no-]from[=<ident>]` | Send as `<ident>` (default: you); real author goes in a body `From:` line. |
| `--[no-]force-in-body-from` | With `--from`: body `From:` even if sender is author. |
| `--[no-]in-reply-to=<message-id>` | First mail (all with `--no-thread`) replies to this message. |
| `--[no-]thread[=<style>]` | Thread headers: `shallow` (all reply to first, default) or `deep` (each to previous). Off by default. Don't combine with `git send-email`, which threads itself. |
| `--attach[=<boundary>]` | Patch as MIME attachment. |
| `--no-attach` | Don't attach (overrides config). |
| `--inline[=<boundary>]` | Patch as inline MIME part. |
| `--[no-]encode-email-headers` | RFC 2047-encode non-ASCII headers. |
