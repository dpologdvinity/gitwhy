# git mailsplit

Split an mbox or Maildir into one file per email (`0001`, `0002`, ...). Mostly used by `git am`.

**Usage:** `git mailsplit [<flags>] -o<directory> [--] [(<mbox>|<Maildir>)...]`

No input given: reads stdin. A Maildir needs `cur`, `tmp`, `new`; messages sort by filename.

## Flags

| Flag | What it does |
|------|--------------|
| `-o<directory>` | Output folder. Required. |
| `-b` | Treat a file without a leading `From ` line as one message. |
| `-d<prec>` | Digits in names (default 4). |
| `-f<n>` | Skip the first `<n>` numbers (`-f3` starts at `0004`). |
| `--keep-cr` | Keep `\r` in `\r\n` endings. |
| `--mboxrd` | Input is mboxrd: undo `>From ` escaping. |
