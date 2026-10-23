# git cvsserver

Act as a CVS server so CVS clients (`cvs`, Eclipse, TortoiseCVS) can check out and commit. Each branch is a CVS module. Clients can't tag, branch, or merge.

Server side. Clients connect over SSH (`export CVS_SERVER="git cvsserver"`, then `cvs -d :ext:user@server/path/repo.git co <branch>`) or pserver (an `/etc/inetd.conf` line running `git-cvsserver pserver`). Each repo needs `gitcvs.enabled` and should be bare for `cvs commit`. Keeps an SQLite DB per branch, so users need write access to the git folder.

Often a separate package (`git-cvs` on Debian/Ubuntu). Needs Perl `DBD::SQLite`.

**Usage:** `git cvsserver [<options>] [pserver|server] [<directory>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `pserver` | Serve pserver protocol (from inetd). Anonymous read-only unless `gitcvs.authdb` lists accounts. |
| `server` | Serve over stdin/stdout (SSH). |
| `--base-path <path>` | Prefix client CVSROOT with `<path>`. Also `GIT_CVSSERVER_BASE_PATH`. |
| `--strict-paths` | Disallow subfolders of listed directories. |
| `--export-all` | Serve repos without `gitcvs.enabled`. Needs a `<directory>` list. |
| `-V`, `--version` | Print version. |
| `-h`, `-H`, `--help` | Print usage. |
| `<directory>...` | Allow only repos in these folders (default all). Also `GIT_CVSSERVER_ROOT` (one folder). |

If `GIT_CVSSERVER_BASE_PATH` or `GIT_CVSSERVER_ROOT` is set, don't pass the matching argument.
