# git archimport

Import a GNU Arch project into git via `tla` (must be installed). First run in an empty folder; rerun with the same arguments to pull new commits. Arch merges become git merges where possible.

Obsolete; not shipped by Debian, Ubuntu, or Fedora.

**Usage:** `git archimport [-h] [-v] [-o] [-a] [-f] [-T] [-D <depth>] [-t <tempdir>] <archive>/<branch>[:<git-branch>]...`

`:<git-branch>` sets the git branch name (e.g. `:master`).

## Flags

| Flag | What it does |
|------|--------------|
| `-h` | Show usage. |
| `-v` | Verbose. |
| `-T` | Tag every imported commit with its Arch name. |
| `-f` | Faster; can't handle folder renames or permission changes. |
| `-o` | Old branch naming (`category--branch`). Names given on the command line still win. |
| `-D <depth>` | Also import merged-in branches, this many levels deep. Use >1 if patch logs were pruned. |
| `-a` | Auto-register archives from mirrors.sourcecontrol.net. Handy with `-D`. |
| `-t <tempdir>` | Temp folder. |
| `<archive>/<branch>` | Arch branch, as `tla log` accepts. |
