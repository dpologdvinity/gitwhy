# git

The `git` command itself. Global flags go before the command name (`git -C ~/proj status`). `git help <command>` opens a command's manual.

`git <name>` runs a built-in, else `git-<name>` on `PATH`, else an alias.

**Usage:** `git [<global-flags>] <command> [<args>]`

## Flags

| Flag | What it does |
|------|--------------|
| `-v`, `--version` | Print version. Same as `git version`; takes its flags (e.g. `--build-options`). `--help` wins if both given. |
| `-h`, `--help` | Print usage and common commands. `-a`/`--all`: every command. With a command name: its manual page. |
| `-C <path>` | Run as if started in `<path>`. Repeatable; relative paths stack. `-C ""` stays put. Affects `--git-dir` and `--work-tree` paths. |
| `-c <name>=<value>` | Set config for this run. `-c foo.bar` means true; `-c foo.bar=` means empty. |
| `--config-env=<name>=<envvar>` | Like `-c`, value from env var `<envvar>` (keeps secrets off the command line). Errors if unset. |
| `--exec-path[=<path>]` | Set helper program folder; bare prints it. Also `GIT_EXEC_PATH`. |
| `--html-path` | Print HTML docs path. |
| `--man-path` | Print man page path. |
| `--info-path` | Print Info docs path. |
| `-p`, `--paginate` | Page output through `less` (or `$PAGER`) on a terminal. Overrides `pager.<cmd>`. |
| `-P`, `--no-pager` | Never page. |
| `--git-dir=<path>` | Path to `.git`. Disables parent search and assumes you're at the work-tree top (else add `--work-tree`). Also `GIT_DIR`. To run elsewhere, use `-C`. |
| `--work-tree=<path>` | Working tree top. Also `GIT_WORK_TREE`, `core.worktree`. |
| `--namespace=<name>` | Use a ref namespace. Also `GIT_NAMESPACE`. |
| `--bare` | Treat repo as bare. Sets `GIT_DIR` to current folder if unset. |
| `--no-replace-objects` | Ignore `git replace` refs. Also `GIT_NO_REPLACE_OBJECTS`. |
| `--no-lazy-fetch` | Partial clone: don't fetch missing objects on demand. Also `GIT_NO_LAZY_FETCH=1`. |
| `--no-optional-locks` | Skip optional locking work (e.g. `git status` index refresh). For background tools. Also `GIT_OPTIONAL_LOCKS=0`. |
| `--no-advice` | Hide all hints. |
| `--literal-pathspecs` | Paths are literal: no globs, no `:(magic)`. Also `GIT_LITERAL_PATHSPECS=1`. |
| `--glob-pathspecs` | All paths are globs; `:(literal)` opts out. Also `GIT_GLOB_PATHSPECS=1`. |
| `--noglob-pathspecs` | All paths literal; `:(glob)` opts in. Also `GIT_NOGLOB_PATHSPECS=1`. |
| `--icase-pathspecs` | Case-insensitive paths. Also `GIT_ICASE_PATHSPECS=1`. |
| `--list-cmds=<group>[,<group>...]` | Internal, may change: list commands by group (`builtins`, `parseopt`, `deprecated`, `main`, `others`, `list-<category>`, `nohelpers`, `alias`, `config`). |
| `--shallow-file=<file>` | Internal: read shallow info from `<file>` instead of `.git/shallow`. Value required (2.56+ checks; older crash). |
| `--attr-source=<tree-ish>` | Read `.gitattributes` from this tree. Also `GIT_ATTR_SOURCE`. |

## Environment variables

Booleans take `true`/`yes`/`on`/`1` or `false`/`no`/`off`/`0`.

| Variable | What it does |
|----------|--------------|
| `GIT_DIR` | `.git` path. Same as `--git-dir`. |
| `GIT_WORK_TREE` | Working tree top. Same as `--work-tree`. |
| `GIT_INDEX_FILE` | Index file instead of `.git/index`. |
| `GIT_OBJECT_DIRECTORY` | Object store instead of `.git/objects`. |
| `GIT_ALTERNATE_OBJECT_DIRECTORIES` | Extra read-only object stores, `:`-separated (`;` on Windows). |
| `GIT_COMMON_DIR` | Shared `.git` for worktrees; `HEAD`, index still from `GIT_DIR`. |
| `GIT_CEILING_DIRECTORIES` | `:`-separated absolute paths git won't search up into. |
| `GIT_DISCOVERY_ACROSS_FILESYSTEM` | Boolean: search for a repo across filesystems. |
| `GIT_DEFAULT_HASH` | New repo hash: `sha1` (default) or `sha256`. |
| `GIT_DEFAULT_REF_FORMAT` | New repo ref format: `files` (default) or `reftable`. |
| `GIT_AUTHOR_NAME`, `GIT_AUTHOR_EMAIL`, `GIT_AUTHOR_DATE` | Author for new commits/tags. Override `user.*`, `author.*`. |
| `GIT_COMMITTER_NAME`, `GIT_COMMITTER_EMAIL`, `GIT_COMMITTER_DATE` | Committer for new commits/tags. Override `user.*`, `committer.*`. |
| `EMAIL` | Fallback email. |
| `GIT_EDITOR` | Message editor. Beats `core.editor`, `VISUAL`, `EDITOR`. |
| `GIT_SEQUENCE_EDITOR` | Editor for `git rebase -i` todo list. |
| `GIT_PAGER` | Pager. Beats `core.pager`, `PAGER`. Empty or `cat` disables. |
| `GIT_EXTERNAL_DIFF` | Program replacing git's diff. |
| `GIT_SSH_COMMAND`, `GIT_SSH` | SSH command (with args) or program for `ssh://` remotes. |
| `GIT_SSH_VARIANT` | `GIT_SSH` flavor: `ssh`, `simple`, `plink`, `putty`, `tortoiseplink`. |
| `GIT_SSL_NO_VERIFY` | Skip HTTPS certificate checks. Unsafe. |
| `GIT_ASKPASS` | Password prompt program. Beats `core.askPass`, `SSH_ASKPASS`. |
| `GIT_TERMINAL_PROMPT` | Boolean: `0` disables terminal credential prompts. |
| `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_SYSTEM` | Replace `~/.gitconfig` and system config. `/dev/null` for none. |
| `GIT_CONFIG_NOSYSTEM` | Boolean: skip system config. |
| `GIT_CONFIG_COUNT`, `GIT_CONFIG_KEY_<n>`, `GIT_CONFIG_VALUE_<n>` | Set config from env, numbered from 0. Like `-c`. |
| `GIT_TRACE` | Trace commands run. `1`/`2` for stderr, absolute path for a file. |
| `GIT_TRACE_PACKET`, `GIT_TRACE_CURL`, `GIT_TRACE_SETUP`, `GIT_TRACE_PERFORMANCE` | Trace packets, HTTP, repo discovery, or timing. Values as `GIT_TRACE`. |
| `GIT_TRACE2`, `GIT_TRACE2_EVENT`, `GIT_TRACE2_PERF` | Structured trace: plain, JSON, or perf format. |
| `GIT_PROGRESS_DELAY` | Seconds before progress shows (default 2). |
| `GIT_MERGE_VERBOSITY` | Merge output level (0-5, default 2). |
| `GIT_REFLOG_ACTION` | Reflog message text; for scripts. |
| `GIT_NO_REPLACE_OBJECTS`, `GIT_NO_LAZY_FETCH`, `GIT_OPTIONAL_LOCKS` | Same as `--no-replace-objects`, `--no-lazy-fetch`, `--no-optional-locks`. |
| `GIT_LITERAL_PATHSPECS`, `GIT_GLOB_PATHSPECS`, `GIT_NOGLOB_PATHSPECS`, `GIT_ICASE_PATHSPECS` | Same as the matching `--*-pathspecs` flags. |
| `GIT_ALLOW_PROTOCOL` | `:`-separated allowed protocols (e.g. `https:ssh`). |
| `GIT_ADVICE` | Boolean: `0` hides hints, like `--no-advice`. |

Full list: [git-scm.com/docs/git#_environment_variables](https://git-scm.com/docs/git#_environment_variables).
