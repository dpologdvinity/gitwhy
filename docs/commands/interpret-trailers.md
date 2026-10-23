# git interpret-trailers

Add or read trailers: `Key: value` lines ending a commit message (`Signed-off-by: Alice <a@x.com>`, `Co-authored-by: ...`). Reads `<file>`s or stdin, prints the result. Works on `git format-patch` output: `---` and the patch pass through untouched.

Existing trailers: the message's last block after a blank line that is all trailers, or 25%+ trailers including a git-made or configured one. The block must end the message or precede `---`. Keys: letters, digits, hyphens; spaces/tabs allowed before `:`. Values continue on lines starting with whitespace.

New trailers are written `key: value` after existing ones (or at the end after a blank line). By default skipped only if an identical trailer is adjacent.

Short names: `git config trailer.sign.key "Signed-off-by"` allows `--trailer "sign: Alice"`. Short name must be a case-insensitive prefix of the key. Other `trailer.*` config sets placement/action defaults and can auto-add trailers via a command (`trailer.<name>.cmd`).

**Usage:** `git interpret-trailers [<flags>] [(--trailer <key>[(=|:)<value>])...] [--parse] [<file>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `<file>...` | Messages to process. Default stdin. |
| `--[no-]trailer=<key>[(=\|:)<value>]` | Add a trailer. Repeatable. `<key>` may be a short name. `--no-trailer` clears the list. |
| `--[no-]in-place` | Edit files instead of printing. |
| `--[no-]trim-empty` | Drop trailers with blank values. |
| `--[no-]where=<placement>` | Placement for later `--trailer`s: `end` (default), `start`, `after` (after last same-key trailer), `before` (before first same-key). Overrides `trailer.where`; `--no-where` reverts to config. |
| `--[no-]if-exists=<action>` | When the key exists: `addIfDifferentNeighbor` (default; add unless same key+value adjacent), `addIfDifferent` (unless same key+value anywhere), `add`, `replace` (drop nearest same-key, then add), `doNothing`. Overrides `trailer.ifExists`; `--no-if-exists` reverts to config. |
| `--[no-]if-missing=<action>` | When the key is absent: `add` (default) or `doNothing`. Overrides `trailer.ifMissing`; `--no-if-missing` reverts to config. |
| `--[no-]only-trailers` | Print only trailers. |
| `--[no-]only-input` | Print only input trailers; add none. |
| `--[no-]unfold` | Join multi-line values onto one line. |
| `--parse` | `--only-trailers --only-input --unfold`: list input trailers. For scripts. |
| `--[no-]divider` | `---` ends the message (default). `--no-divider` for plain messages, not emails/patches. |
