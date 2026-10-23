# git check-ref-format

Check that a name is a valid ref name; exit non-zero if not. For scripts.

A ref name must not:

- have a `/`-separated part that starts with `.` or ends with `.lock`
- be a single part with no `/` (unless `--allow-onelevel`)
- contain `..`, `@{`, `\`, a space, control characters, or any of `~ ^ : ? * [`
- start or end with `/`, or have `//` (unless `--normalize`)
- end with `.`, or be just `@`

**Usage:** `git check-ref-format [--normalize] [--[no-]allow-onelevel] [--refspec-pattern] <refname>`
or `git check-ref-format --branch <branch>`

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]allow-onelevel` | Allow names with no `/`. Off by default. |
| `--refspec-pattern` | Allow one `*`. |
| `--normalize` | Strip leading `/`, squash repeated `/`; print result if valid. |
| `--print` | Deprecated: `--normalize`. |
| `--branch <branch>` | Check as a branch name (can't start with `-`). Expands `@{-N}` and prints the result (a commit ID if that checkout was detached). |
