# gitweb

Web UI for git repos (branches, commits, files, diffs, blame, search, feeds). A Perl CGI script run by your web server; no command-line flags. Quickest start: [`git instaweb`](instaweb.md).

Often a separate package (`gitweb` on Debian/Ubuntu).

## Setup

Settings go in `gitweb_config.perl` or `/etc/gitweb.conf` (or `GITWEB_CONFIG`); see `man gitweb.conf`.

| Setting | What it does |
|---------|--------------|
| `$projectroot` | Folder of repos to show (default `/pub/git`). Bare repos work best. |
| `$projects_list` | File listing repos to show: `<path> <owner>` per line, URL-encoded (`+` for space). The TXT link (`project_index`) generates one. |
| `$strict_export` | Only allow listed repos, not just hide others. |
| `$export_ok` | Only show repos containing this file, e.g. `git-daemon-export-ok`. |
| `$export_auth_hook` | Perl function given a repo path; true allows it. |
| `$feature{'<name>'}{'default'} = [1]` | Enable a feature, e.g. `blame` or `pathinfo` (URLs without `gitweb.cgi`). |

Per repo, in its git folder:

| File or config | What it does |
|----------------|--------------|
| `README.html` | HTML on the summary page. Only if `$prevent_xss` is off. |
| `description` or `gitweb.description` | One-line description. File wins. |
| `category` or `gitweb.category` | Group name, if grouping is on. File wins. |
| `cloneurl` or `gitweb.url` (repeatable) | Clone URLs, one per line. File wins. |
| `gitweb.owner` | Owner. Default: list file, or folder owner's name. |

## URLs

`.../gitweb.cgi/<repo>/<action>/<revision>:/<path>?<arguments>`. Two-sided actions: `<rev-from>:/<path-from>..<rev-to>:/<path-to>`. Revision defaults to `HEAD`. Action defaults: `project_list` (no repo), `summary` (repo), `tree` (path).

| Action | What it shows |
|--------|---------------|
| `project_list` | All repos. |
| `summary` | Repo overview. |
| `heads`, `remotes` | Local or remote-tracking branches (`remotes` off by default). |
| `tags` | All tags. |
| `tree`, `blob` | Folder listing or file at a revision. |
| `blob_plain` | Raw file. |
| `blobdiff` | One file's diff between revisions. |
| `blame`, `blame_incremental` | Who last changed each line. Off by default (slow). Incremental loads via JavaScript. |
| `commit`, `commitdiff` | Commit details, or its changes. |
| `patch` | Commit as an email patch for `git am`. |
| `tag` | One annotated tag. |
| `log`, `shortlog` | Branch history; `shortlog` is one line per commit. |
| `history` | One file or folder's history. |
| `rss`, `atom` | Recent-commit feed. |
