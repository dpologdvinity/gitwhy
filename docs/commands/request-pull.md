# git request-pull

Print a pull request message: branch description, commit summary since `<start>`, and where to pull from. Push to `<URL>` first.

Example: `git request-pull v1.0 https://git.ko.xz/project master` asks for changes from `v1.0` to your `master`.

**Usage:** `git request-pull [-p] <start> <URL> [<end>]`

## Flags

| Flag | What it does |
|------|--------------|
| `-p` | Include the full diff. |
| `<start>` | Commit upstream already has. |
| `<URL>` | Public repo to pull from. |
| `<end>` | Last commit to include (default `HEAD`). Use `<local>:<remote>` if the remote branch name differs, e.g. `master:for-linus`. |
