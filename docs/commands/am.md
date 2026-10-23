# git am

Turn patch emails (mbox, Maildir, or stdin) into commits on the current branch, keeping each email's author, date, and message. Reverse of `git format-patch`.

Author comes from `From:`, date from `Date:`, title from `Subject:` (minus `[PATCH ...]`), message from the body. The patch starts at the first line that is `---` or begins with `diff -` or `Index: `, so indent any diff meant for the message.

On a failed patch, `am` stops: fix and `git add`, then `--continue`, `--skip`, or `--abort`. Sets `ORIG_HEAD` to the starting tip. Runs `applypatch-msg`, `pre-applypatch`, `post-applypatch` hooks.

**Usage:** `git am [<flags>] [(<mbox> | <Maildir>)...]`
or `git am (--continue | --skip | --abort | --quit | --retry | --show-current-patch[=(diff|raw)] | --allow-empty)`

## Flags

| Flag | What it does |
|------|--------------|
| `-s`, `--[no-]signoff` | Add your `Signed-off-by:` trailer. |
| `-k`, `--[no-]keep` | Keep the subject as-is (don't strip `[PATCH]` etc.). |
| `--[no-]keep-non-patch` | Strip only `[PATCH]`; keep other brackets like `[RFC]`. |
| `--[no-]keep-cr` | Keep `\r` line endings in mbox. Default: `am.keepcr`. |
| `-c`, `--[no-]scissors` | Drop body text above a scissors line (`-- >8 --`). Default: `mailinfo.scissors`. |
| `--quoted-cr=<action>` | `\r` in base64/quoted-printable lines: `nowarn`, `warn`, or `strip`. |
| `--empty=<stop\|drop\|keep>` | Email with no patch: `stop` (default), `drop`, or `keep` as empty commit. |
| `-m`, `--[no-]message-id` | Add the email's `Message-ID` to the message. Default: `am.messageid`. |
| `-q`, `--[no-]quiet` | Errors only. |
| `-u`, `--[no-]utf8` | Re-encode message to UTF-8 (or `i18n.commitEncoding`). Default on. |
| `-3`, `--[no-]3way` | If a patch fails, try a three-way merge (needs the original blobs). Default: `am.threeWay`. |
| `--[no-]rerere-autoupdate` | Stage conflict resolutions rerere reuses. `--no-`: leave unstaged to check. |
| `--[no-]ignore-space-change`, `--[no-]ignore-whitespace` | Ignore whitespace changes in context lines. Passed to [`git apply`](apply.md). |
| `--[no-]whitespace=<action>` | Whitespace errors: `nowarn`, `warn`, `fix`, `error`, `error-all`. Passed to `git apply`. |
| `-C <n>` | Need only `<n>` context lines to match. Passed to `git apply`. |
| `-p <n>` | Strip `<n>` leading folders from paths (default 1). Passed to `git apply`. |
| `--[no-]directory=<dir>` | Prefix every path with `<dir>/`. Passed to `git apply`. |
| `--[no-]exclude=<path>` | Skip paths matching this glob. Passed to `git apply`. |
| `--[no-]include=<path>` | Apply only paths matching this glob. Passed to `git apply`. |
| `--[no-]reject` | Apply what fits; write failed hunks to `.rej` files. Passed to `git apply`. |
| `--[no-]patch-format=<format>` | Force format: `mbox`, `mboxrd`, `stgit`, `stgit-series`, or `hg`. |
| `-i`, `--[no-]interactive` | Ask before each patch. |
| `--verify` | Run `pre-applypatch` and `applypatch-msg` hooks. Default. |
| `-n`, `--no-verify` | Skip those two hooks. `post-applypatch` still runs. |
| `--[no-]committer-date-is-author-date` | Committer date = email date. Can make dates go backwards. |
| `--[no-]ignore-date` | Author date = now, not the email's. |
| `-S`, `--[no-]gpg-sign[=<key-id>]` | GPG-sign commits. Key defaults to committer; attach it (`-SABC123`). `--no-` overrides `commit.gpgSign`. |
| `--[no-]resolvemsg=<msg>` | Internal (for `rebase`). Message shown when a patch fails. |

## Resume flags

For a stopped `am` session.

| Flag | What it does |
|------|--------------|
| `--continue`, `-r`, `--resolved` | Commit the fixed patch from the index with the email's author and message; go on. |
| `--skip` | Drop current patch; go on. |
| `--abort` | Stop; restore branch and files to before `am`. |
| `--quit` | Stop; leave `HEAD` and index as they are. |
| `--retry` | Retry the failed patch, e.g. with new flags like `--3way`. |
| `--show-current-patch[=<diff\|raw>]` | Show the stopped email: `raw` (default) or `diff` (patch only). |
| `--allow-empty` | Email had no patch: commit it empty with its message. |
