# git check-mailmap

Print each contact's canonical `Name <email>` per `.mailmap` (which maps old or misspelled identities to one), or the input unchanged if unmapped.

Contacts: `Name <user@host>`, `<user@host>`, or `user@host`. One output line each; name shown only if given or known to the mailmap.

**Usage:** `git check-mailmap [<flags>] <contact>...`

## Flags

| Flag | What it does |
|------|--------------|
| `<contact>...` | People to look up. |
| `--[no-]stdin` | Also read contacts from stdin, one per line. |
| `--[no-]mailmap-file=<file>` | Also read this mailmap; beats `.mailmap`, `mailmap.file`, `mailmap.blob`. |
| `--[no-]mailmap-blob=<blob>` | Same, from a repo blob (e.g. `HEAD:.mailmap`). `--mailmap-file` wins if both given. |
