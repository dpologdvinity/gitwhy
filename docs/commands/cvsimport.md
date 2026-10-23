# git cvsimport

Import a CVS repo into git, fresh or on top of an earlier import. Needs outdated `cvsps` 2.x (not 3+); for one-time moves, `cvs2git` or `cvs-fast-export` do better. Some histories import wrong (unstable timestamps, empty branches, several tags on one revision).

Don't commit on imported branches; work on your own and `git merge` new imports. `-r` keeps imported branches separate.

Often a separate package (`git-cvs` on Debian/Ubuntu).

**Usage:** `git cvsimport [-o <branch-for-HEAD>] [-h] [-v] [-d <CVSROOT>] [-A <author-conv-file>] [-p <options-for-cvsps>] [-P <file>] [-C <git-repository>] [-z <fuzz>] [-i] [-k] [-u] [-s <subst>] [-a] [-m] [-M <regex>] [-S <regex>] [-L <commit-limit>] [-r <remote>] [-R] [<CVS-module>]`

## Flags

| Flag | What it does |
|------|--------------|
| `-v` | Verbose. |
| `-d <CVSROOT>` | CVS repo: local path, `:local:`, `:ext:`, or `:pserver:`. Default `CVS/Root`, then `$CVSROOT`. |
| `<CVS-module>` | Module to import. Default `CVS/Repository`. |
| `-C <target-dir>` | Git repo to import into (created if missing). Default: current folder. |
| `-r <remote>` | Put CVS branches under `remotes/<remote>/<branch>`. |
| `-o <branch-for-HEAD>` | Branch for CVS's main line. Default `origin`, or `remotes/<remote>/master` with `-r`. `-o master` continues an old `cvs2git` import. |
| `-i` | Import only; don't touch working tree or index. |
| `-k` | Strip CVS keywords (`-kk`). Recommended; off by default. |
| `-u` | Turn `_` into `.` in tag and branch names. |
| `-s <subst>` | Replace `/` in branch names with `<subst>`. |
| `-p <options-for-cvsps>` | Extra comma-separated `cvsps` options. Not `-u` or `-A` (always on). |
| `-z <fuzz>` | Max seconds between file changes in one commit (cvsps default 300). |
| `-P <cvsps-output-file>` | Use saved `cvsps` output instead of running it. |
| `-m` | Detect merges from commit messages (built-in patterns). |
| `-M <regex>` | Custom merge pattern; `$1` captures the source branch. Escape `/`. Repeatable; add `-m` to keep built-ins. |
| `-S <regex>` | Skip matching paths. |
| `-a` | Include commits under 10 minutes old (skipped by default). |
| `-L <limit>` | Import at most this many commits. Memory-leak workaround. |
| `-A <author-conv-file>` | Map CVS users to names, emails, optional time zones: `exon=Andreas Ericsson <ae@op5.se>` or `... <email> America/Chicago`. Saved to `$GIT_DIR/cvs-authors` for reuse. Avoid if exporting back with `cvsexportcommit`. |
| `-R` | Write `$GIT_DIR/cvs-revisions` mapping each file's CVS revision to its git commit. Appended on later imports. |
| `-h` | Show usage. |
