# git ls-remote

List a remote's refs and their commit IDs. Changes nothing locally.

Output: `<id> TAB <ref>` per ref. Annotated tags add a `<ref>^{}` line with the target's ID.

**Usage:** `git ls-remote [<flags>] [<repository> [<pattern>...]]`

## Flags

| Flag | What it does |
|------|--------------|
| `<repository>` | URL or remote name. Default: current branch's remote, or `origin`. |
| `<pattern>...` | Only refs whose tail matches a glob (`bar` matches `refs/heads/bar`, not `foobar`). |
| `-b`, `--[no-]branches` | Branches only. With `-t`: both. Old names: `--heads`, `-h` (alone, `-h` is help). |
| `-t`, `--[no-]tags` | Tags only. |
| `--[no-]refs` | Hide `^{}` lines and pseudorefs like `HEAD`. |
| `-q`, `--[no-]quiet` | Don't print the remote URL to stderr. |
| `--[no-]upload-pack=<exec>` | Path to `git-upload-pack` on the remote. |
| `--[no-]exit-code` | Exit 2 if no refs match. |
| `--[no-]get-url` | Print the rewritten remote URL; don't connect. |
| `--[no-]symref` | Show symbolic refs' targets (usually `HEAD`). |
| `--[no-]sort=<key>` | Sort by `<key>`; `-` for descending. `version:refname` sorts as versions. Object keys fail for unfetched objects. |
| `-o <option>`, `--[no-]server-option=<option>` | Send `<option>` to the server (protocol v2). Repeatable. Default: `remote.<name>.serverOption`. |
