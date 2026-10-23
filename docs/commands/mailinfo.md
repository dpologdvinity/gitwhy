# git mailinfo

Low-level: split one email from stdin into commit message `<msg>` and patch `<patch>`; print author name, email, subject. Used by [`git am`](am.md).

**Usage:** `git mailinfo [<flags>] <msg> <patch> < <email>`

## Flags

| Flag | What it does |
|------|--------------|
| `<msg>` | Commit message output (title from Subject). |
| `<patch>` | Patch output. |
| `-k` | Keep Subject as is. Default strips whitespace, leading `Re:`/`re:`/`:`, and `[...]` tags like `[PATCH]`. Pairs with `git format-patch -k`. |
| `-b` | Without `-k`: strip only `[...]` tags containing `PATCH`. |
| `-u` | Convert message, name, email to `i18n.commitEncoding` (default UTF-8). Default. Patch never converted. |
| `--encoding=<encoding>` | Like `-u`, to `<encoding>`. |
| `-n` | No charset conversion. |
| `-m`, `--[no-]message-id` | Append the `Message-ID` header to the message. |
| `--[no-]scissors` | Drop body above a scissors line (`-- >8 --`). Default from `mailinfo.scissors`. |
| `--quoted-cr=<action>` | If base64/quoted-printable body decodes to CRLF: `nowarn`, `warn` (default, or `mailinfo.quotedCR`), `strip` (to LF). |
