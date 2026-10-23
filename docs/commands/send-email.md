# git send-email

Email patches: patch files, folders of them, or a commit range (run through [`git format-patch`](format-patch.md), whose flags also work). Prompts for anything not given by flag or `sendemail.*` config.

Patch files: mbox (from `format-patch`) or old `send_lots_of_email.pl` format (line 1 Cc, line 2 Subject).

Often a separate package (`git-email` on Debian/Ubuntu). Needs Perl `Authen::SASL` and `Mail::Address` (plus `Net::SMTP` etc., usually bundled).

**Usage:** `git send-email [<options>] (<file>|<directory>)...`
or `git send-email [<options>] <format-patch-options>`
or `git send-email --dump-aliases`
or `git send-email --translate-aliases`

Typical: `git format-patch --cover-letter -M origin/master -o outgoing/`, edit `outgoing/0000-*`, then `git send-email outgoing/*`. Set `sendemail.smtpServer`, `smtpUser`, `smtpEncryption`, `smtpServerPort` once. Gmail needs an app password or OAuth (`smtpAuth = OAUTHBEARER`); Outlook needs OAuth (`XOAUTH2`).

## Flags

### Composing

| Flag | What it does |
|------|--------------|
| `--[no-]annotate` | Edit each patch before sending. Default `sendemail.annotate`. |
| `--bcc=<address>,...` | Bcc all. Repeatable. Default `sendemail.bcc`. |
| `--cc=<address>,...` | Cc all (plus automatic ones). Repeatable. Default `sendemail.cc`. |
| `--compose` | Write an intro email in your editor; its From, To, Cc, Bcc, Subject, Reply-To, In-Reply-To headers apply. Empty body (or only `GIT:` lines): not sent, headers still apply. |
| `--from=<address>`, `--sender=<address>` | Sender. Default `sendemail.from`, else prompts (suggesting your author identity). |
| `--reply-to=<address>` | Reply address, if not From. |
| `--in-reply-to=<identifier>` | Message-ID to reply to (e.g. keep a new series version in the old thread). Applies to the first email (all with `--no-thread`); others follow `--[no-]chain-reply-to`. Prompted if absent and no `--compose`. |
| `--[no-]outlook-id-fix` | Fetch the Message-ID Outlook servers substitute, so threads work. Default on for `smtp.office365.com` and `smtp-mail.outlook.com`. |
| `--subject=<string>` | First email's subject. Prompted if absent and no `--compose`. |
| `--to=<address>,...` | Main recipient (usually maintainer). Repeatable. Default `sendemail.to`; prompted if unset and no `--to-cmd`. |
| `--8bit-encoding=<encoding>` | Assumed encoding for undeclared non-ASCII. Default `sendemail.assume8bitEncoding`, else prompts. Unchecked. |
| `--compose-encoding=<encoding>` | `--compose` email encoding. Default `sendemail.composeEncoding`, else UTF-8. |
| `--transfer-encoding=(7bit\|8bit\|quoted-printable\|base64\|auto)` | `7bit` fails on non-ASCII; `quoted-printable` handles CRs, hard to read raw; `base64` safest, unreadable; `auto` (default): `8bit` if possible, else `quoted-printable`. Default `sendemail.transferEncoding`. |
| `--[no-]xmailer` | Add `X-Mailer:` (default on; `sendemail.xmailer`). |

### Sending

| Flag | What it does |
|------|--------------|
| `--envelope-sender=<address>` | Envelope sender (not visible From), e.g. the list-subscribed address. `auto` = From. Sendmail needs `-f` rights. Default `sendemail.envelopeSender`, else system decides. |
| `--sendmail-cmd=<command>` | Send via a sendmail-like command accepting `-i` (e.g. `msmtp`), shell-run if needed. Default `sendemail.sendmailCmd`; without it or `--smtp-server`, finds `sendmail` in `/usr/sbin`, `/usr/lib`, `$PATH`. |
| `--smtp-encryption=<encryption>` | `ssl`: SMTPS (port 465). `tls`: STARTTLS (default port 25; often 587). Else none. Default `sendemail.smtpEncryption`. |
| `--smtp-domain=<FQDN>` | HELO/EHLO host name (some servers check it against your IP). Guessed if unset. Default `sendemail.smtpDomain`. |
| `--smtp-auth=<mechanisms>` | Allowed login methods, space-separated (`"PLAIN LOGIN GSSAPI"`). `none` disables login. Default: all SASL-supported. |
| `--smtp-pass[=<password>]` | SMTP password (empty if no value). Overrides `sendemail.smtpPass`. Omit with `--smtp-user` to use a [credential helper](credential.md). |
| `--no-smtp-auth` | No login. = `--smtp-auth=none`. |
| `--smtp-server=<host>` | SMTP host or IP. Default `sendmail` if found, else `localhost`. A sendmail-like program path also works (old style, no args; prefer `--sendmail-cmd`). |
| `--smtp-server-port=<port>` | Port number or name (`submission`): usually 25, 587, 465. Also `sendemail.smtpServerPort`. |
| `--smtp-server-option=<option>` | Server option. Repeat for each. Default `sendemail.smtpServerOption`. |
| `--smtp-ssl` | Old `--smtp-encryption ssl`. |
| `--smtp-ssl-cert-path <path>` | CA certs: `c_rehash` folder or PEM file. Empty string disables verification. Default `sendemail.smtpSSLCertPath`, else library default. |
| `--smtp-ssl-client-cert <path>` | Client certificate. Key separate? Add `--smtp-ssl-client-key`. Default `sendemail.smtpSSLClientCert`. |
| `--smtp-ssl-client-key <path>` | Client certificate's private key. Default `sendemail.smtpSSLClientKey`. |
| `--smtp-user=<user>` | SMTP username. None (here or `sendemail.smtpUser`) = no login. |
| `--smtp-debug=(0\|1)` | `1` prints SMTP traffic. For TLS/login problems. |
| `--imap-sent-folder=<folder>` | Also copy sent mail to this IMAP folder (for providers like iCloud). `git imap-send --list` shows folders. Needs [`git imap-send`](imap-send.md) configured. |
| `--[no-]use-imap-only` | Only copy to `--imap-sent-folder`; don't send. For drafts. Default off; `sendemail.useImapOnly`. |
| `--batch-size=<num>` | Reconnect every `<num>` emails (per-connection caps). Default `sendemail.smtpBatchSize`. |
| `--relogin-delay=<int>` | Seconds before reconnecting. With `--batch-size`. Default `sendemail.smtpReloginDelay`. |

### Automating

| Flag | What it does |
|------|--------------|
| `--no-to`, `--no-cc`, `--no-bcc` | Drop config To/Cc/Bcc addresses. |
| `--no-identity` | Ignore `sendemail.identity`. |
| `--to-cmd=<command>` | Run `<command> <patch>` per patch; output lines are To addresses. Default `sendemail.toCmd`. |
| `--cc-cmd=<command>` | Same, for Cc. Default `sendemail.ccCmd`. |
| `--header-cmd=<command>` | Run per email; output lines added as headers. Beats `sendemail.headerCmd`. |
| `--no-header-cmd` | Run no header command. |
| `--[no-]chain-reply-to` | Each email replies to the previous (deep thread). Off: all reply to the first. Also `sendemail.chainReplyTo`. |
| `--identity=<identity>` | Prefer `sendemail.<identity>.*` over `sendemail.*`. Default `sendemail.identity`. |
| `--[no-]signed-off-by-cc`, `--[no-]signed-off-cc` | Cc `Signed-off-by:` and `Cc:` people. Default on; `sendemail.signedOffByCc`. |
| `--[no-]cc-cover` | Cc the cover letter's `Cc:` on all emails. Default off; `sendemail.ccCover`. |
| `--[no-]to-cover` | Same for its `To:`. Default off; `sendemail.toCover`. |
| `--suppress-cc=<category>` | Don't auto-Cc a group. Repeatable. `author`, `self`, `cc` (patch header `Cc:`), `bodycc` (message `Cc:`), `sob`, `misc-by` (`Acked-by:`, `Reviewed-by:`, `Tested-by:`, other `-by:`), `cccmd` (skip `--cc-cmd`), `body` (`sob` + `bodycc` + `misc-by`), `all`. Default `sendemail.suppressCc`. (`cc`, `bodycc`, `sob` never drop you; use `self`.) |
| `--[no-]suppress-from` | Don't Cc the From address. Default off; `sendemail.suppressFrom`. |
| `--[no-]thread` | Add `In-Reply-To`/`References` headers. Default on; `sendemail.thread`. Turn off if `format-patch` added them. |
| `--[no-]mailmap`, `--[no-]use-mailmap` | Map addresses via [mailmap](check-mailmap.md). Extra: `sendemail.mailmap.file`, `sendemail.mailmap.blob`. Default `sendemail.mailmap` (off). |

### Administering

| Flag | What it does |
|------|--------------|
| `--confirm=<mode>` | Confirm before sending: `always`, `never`, `cc` (if Cc's auto-added), `compose` (first email with `--compose`), `auto` (`cc` + `compose`). Default `sendemail.confirm`, else `auto` (`compose` if any suppress flag). |
| `--dry-run` | Do everything but send. |
| `--[no-]format-patch` | Ambiguous argument (file or commit): treat as commit (`--format-patch`) or file (`--no-format-patch`). Default: error. |
| `--quiet` | One line per email. |
| `--[no-]validate` | Check patches first: `sendemail-validate` hook, plus warn on lines over 998 characters unless encoding is `auto`, `base64`, or `quoted-printable`. Default on; `sendemail.validate`. |
| `--force` | Send despite failed checks. |
| `-v <n>` | Reroll count for `format-patch` (`[PATCH v<n>]`), when generating from commits. |

### Information

| Flag | What it does |
|------|--------------|
| `--dump-aliases` | Print sorted alias names from alias files and exit. See `sendemail.aliasesFile`. |
| `--translate-aliases` | Read aliases from stdin; print names and addresses, exit. |
| `-h` | Usage. |

## Settings with no flag

| Config | What it does |
|--------|--------------|
| `sendemail.multiEdit` | `true` (default): edit all files (`--annotate`, `--compose`) in one editor. `false`: one at a time. |
| `sendemail.aliasesFile`, `sendemail.aliasFileType` | Alias files and format: `mutt`, `mailrc`, `pine`, `elm`, `gnus`, `sendmail`. |
| `sendemail.forbidSendmailVariables` | Set to skip aborting when old `sendmail.*` config exists. |
