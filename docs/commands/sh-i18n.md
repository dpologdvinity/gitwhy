# git sh-i18n

Not a command: a shell library git's scripts source for translated messages. Wraps GNU gettext; falls back to plain output without it.

**Usage (inside a script):** `. "$(git --exec-path)/git-sh-i18n"`

## Functions

No flags. Defines:

| Function | What it does |
|----------|--------------|
| `gettext` | Print a message, translated if possible. |
| `eval_gettext` | Same, also fills in `$variables`. |
