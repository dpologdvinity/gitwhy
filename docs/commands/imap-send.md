# git imap-send

Upload patches from an mbox (e.g. `git format-patch` output) into an IMAP folder, usually Drafts, to send from any mail client. Messages need `From`, `Date`, `Subject` in that order.

Example: `git format-patch --signoff --stdout --attach origin | git imap-send`

**Usage:** `git imap-send [<flags>] < <mbox>`, `git imap-send --list`

## Flags

| Flag | What it does |
|------|--------------|
| `-v`, `--[no-]verbose` | More output. |
| `-q`, `--[no-]quiet` | Less output. |
| `-f <folder>`, `--[no-]folder=<folder>` | Target folder, e.g. `[Gmail]/Drafts`. Overrides `imap.folder`. |
| `--curl` | Use libcurl (unless `imap.tunnel`). Ignored if built without `USE_CURL_FOR_IMAP_SEND`. |
| `--no-curl` | Use git's own IMAP code. Ignored if built with `NO_OPENSSL`. |
| `--[no-]draft` | (2.56+) Set the IMAP `\Draft` flag. Off by default. Via libcurl needs 8.13.0+; older uploads without it. |
| `--[no-]list` | List server folders and exit. |

## Setup

Set `imap.folder`, plus `imap.host` (with `imap.user`, `imap.pass`, `imap.port`) or `imap.tunnel` (a command, e.g. `ssh -q -C user@example.com /usr/bin/imapd ./Maildir`). Also: `imap.sslVerify`, `imap.authMethod`, `imap.preformattedHTML`. Use `imaps://`; `imap://` sends your password in plain text.

```ini
[imap]
    folder = "[Gmail]/Drafts"
    host = imaps://imap.gmail.com
    user = you@gmail.com
    port = 993
```

- Gmail: use an app password, or OAuth2 with `authMethod = OAUTHBEARER`. Folder may be `[Google Mail]/Drafts` or translated.
- Outlook: `host = imaps://outlook.office365.com`, `folder = Drafts`, `authMethod = XOAUTH2` (OAuth2 only).
- Mail clients may mangle patches (line wraps, whitespace). Check before sending.
