# git cvsexportcommit

Apply one git commit to a CVS checkout. Checks CVS files are up to date first; commits to CVS only with `-c`. For a merge, name the parent to diff against.

Often a separate package (`git-cvs` on Debian/Ubuntu). Needs `cvs`.

**Usage:** `git cvsexportcommit [-h] [-u] [-v] [-c] [-P] [-p] [-a] [-d <cvsroot>] [-w <cvs-workdir>] [-W] [-f] [-m <msgprefix>] [<parent-commit>] <commit-id>`

Run from the CVS checkout root with `GIT_DIR` set, or from the git repo with `-w`.

## Flags

| Flag | What it does |
|------|--------------|
| `-h` | Show usage. |
| `-c` | Commit to CVS if the patch applied cleanly. |
| `-p` | Strict patching (`patch --fuzz=0`). |
| `-a` | Add Author (and Committer, if different) lines to the CVS message. |
| `-d <cvsroot>` | Use this CVSROOT (like `cvs -d`). |
| `-f` | Proceed even if CVS files aren't up to date. |
| `-P` | Allow a parent commit that isn't a direct parent. |
| `-m <msgprefix>` | Prefix the commit message. |
| `-u` | Update affected files from CVS first. |
| `-k` | Undo CVS keyword expansion (`$Revision: 1.2$` to `$Revision$`) before patching. |
| `-w <cvs-workdir>` | CVS checkout path. No `GIT_DIR` needed inside the git repo. Default `cvsexportcommit.cvsdir`. |
| `-W` | Current folder is both git and CVS checkout. Checks out the parent commit there first. |
| `-v` | Verbose. |
