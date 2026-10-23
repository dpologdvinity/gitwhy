# git rev-parse

Low-level, for scripts. Turns names like `HEAD~2` or `main` into object IDs, answers repo questions (top directory? inside a repo?), and parses shell-script options. Most modes need a repository.

**Usage:** `git rev-parse [<flags>] [<arg>...]`, `git rev-parse --parseopt [<flags>] -- [<args>...]`, `git rev-parse --sq-quote [<arg>...]`

Common uses:

| Command | Prints |
|---------|--------|
| `git rev-parse HEAD` | Full ID of the current commit. |
| `git rev-parse --short HEAD` | Short ID. |
| `git rev-parse --abbrev-ref HEAD` | Current branch name (`HEAD` if detached). |
| `git rev-parse --show-toplevel` | Path of the repo's top directory. |
| `git rev-parse --verify --end-of-options "$REV^{commit}"` | ID if `$REV` is a real commit; error otherwise. |

## Flags

### Modes (must come first)

| Flag | What it does |
|------|--------------|
| `--parseopt` | Parse a shell script's options (see [--parseopt](#--parseopt)). Works outside a repo. |
| `--sq-quote` | Only shell-quote the arguments into one line. Works outside a repo. |

### For --parseopt

| Flag | What it does |
|------|--------------|
| `--keep-dashdash` | Print the first `--` instead of dropping it. |
| `--stop-at-non-option` | Stop at the first non-option (for scripts with subcommands). |
| `--stuck-long` | Print long forms with values attached (`--foo=bar`). |

### Filtering

Sorts arguments into "for `git rev-list`" (revisions and revision flags) and "other".

| Flag | What it does |
|------|--------------|
| `--revs-only` | Print only `git rev-list` arguments. |
| `--no-revs` | Print only non-`git rev-list` arguments. |
| `--flags` | Print only flags (starting with `-`). |
| `--no-flags` | Print only non-flags. |

### Output

| Flag | What it does |
|------|--------------|
| `--default <arg>` | Use `<arg>` if no arguments given. |
| `--prefix <arg>` | Act as if run from subdirectory `<arg>`; resolve and print relative paths from there. |
| `--verify` | Exactly one argument naming an object; print its ID or fail. Add `^{commit}` (or another type) to check type, `^{object}` to check existence. |
| `-q`, `--quiet` | With `--verify`: fail silently (non-zero exit). |
| `--end-of-options` | Everything after is an argument, not a flag. Use before untrusted input. |
| `--sq` | Print one shell-quoted line. Arguments still interpreted (unlike `--sq-quote`). |
| `--short[=<length>]` | `--verify`, printing a unique short ID of at least `<length>` chars (min 4; default `core.abbrev`). |
| `--not` | Toggle `^` on each printed object name. |
| `--abbrev-ref[=(strict\|loose)]` | Short unambiguous ref name (`main`, not `refs/heads/main`). Default mode from `core.warnAmbiguousRefs`. |
| `--symbolic` | Print names close to how you typed them, not IDs. |
| `--symbolic-full-name` | Full ref names (`refs/heads/main`). Skips non-refs. |
| `--output-object-format=(sha1\|sha256\|storage)` | Print IDs in this hash format, translating if the repo supports both. |

### Objects

| Flag | What it does |
|------|--------------|
| `--all` | Every ref under `refs/`. |
| `--branches[=<pattern>]` | All branches, or those matching a glob. No `?`, `*`, `[` = prefix match. |
| `--tags[=<pattern>]` | All tags, or matching. |
| `--remotes[=<pattern>]` | All remote-tracking branches, or matching. |
| `--glob=<pattern>` | Refs matching a glob (`refs/` added if missing). |
| `--exclude=<glob-pattern>` | Skip matching refs in the next `--all`, `--branches`, `--tags`, `--remotes`, or `--glob`. Repeatable. Omit `refs/heads/` etc. for `--branches`/`--tags`/`--remotes`; include `refs/` for `--glob`/`--all`. |
| `--exclude-hidden=(fetch\|receive\|uploadpack)` | Skip refs hidden by `<service>.hideRefs` (plus `transfer.hideRefs`) in the next `--all` or `--glob`. |
| `--disambiguate=<prefix>` | List every object whose ID starts with `<prefix>` (4+ hex digits). |
| `--bisect` | `refs/bisect/bad` plus negated `refs/bisect/good-*`: the range left to test. |

### Files and repository info

| Flag | What it does |
|------|--------------|
| `--local-env-vars` | List repo-specific `GIT_*` variable names (`GIT_DIR`, not `GIT_EDITOR`). |
| `--path-format=(absolute\|relative)` | Path style for later path flags, until the next `--path-format`. |
| `--git-dir` | Path to `.git` (or `$GIT_DIR`). Fails outside a repo. Uses `--path-format`. |
| `--git-common-dir` | `$GIT_COMMON_DIR` if set, else git dir. With worktrees, the shared main `.git`. Uses `--path-format`. |
| `--resolve-git-dir <path>` | Check `<path>` is a repo or `.git` file pointing to one; print real repo path. Uses `--path-format`. |
| `--git-path <path>` | Path of `$GIT_DIR/<path>`, honoring `GIT_OBJECT_DIRECTORY`, `GIT_INDEX_FILE`, etc. Uses `--path-format`. |
| `--show-toplevel` | Absolute path of the working tree's top. Fails with no working tree. Uses `--path-format`. |
| `--show-superproject-working-tree` | Parent repo's top if this is a submodule; else nothing. Uses `--path-format`. |
| `--shared-index-path` | Shared index path in split-index mode; else empty. Uses `--path-format`. |
| `--absolute-git-dir` | `--git-dir`, always absolute. |
| `--is-inside-git-dir` | `true` if inside `.git`. |
| `--is-inside-work-tree` | `true` if inside the working tree. |
| `--is-bare-repository` | `true` if bare. |
| `--is-shallow-repository` | `true` if shallow (partial history). |
| `--show-cdup` | Path up to the top directory (`../../`), or empty. |
| `--show-prefix` | Path from the top directory down to here. |
| `--show-object-format[=(storage\|input\|output\|compat)]` | Hash algorithm (`sha1` or `sha256`). Default `storage`. `compat` prints an empty line if none. |
| `--show-ref-format` | Ref storage format: `files` or `reftable`. |

### Other

| Flag | What it does |
|------|--------------|
| `--since=<date>`, `--after=<date>` | Print matching `--max-age=<timestamp>` for `git rev-list`. |
| `--until=<date>`, `--before=<date>` | Print the matching `--min-age=<timestamp>`. |
| `<arg>...` | Flags and arguments to parse. |

## Naming commits and objects

Works in every git command that takes a commit or object.

| Syntax | Means |
|--------|-------|
| `dae86e1950b1...`, `dae86e` | Object ID, full or a unique prefix. |
| `v1.7.4.2-679-g3bee7fb` | `git describe` output. |
| `main`, `heads/main`, `refs/heads/main` | A ref. Short names are tried as: `.git/<name>`, `refs/<name>`, `refs/tags/<name>`, `refs/heads/<name>`, `refs/remotes/<name>`, `refs/remotes/<name>/HEAD`. |
| `HEAD` | The commit you're on. |
| `FETCH_HEAD`, `ORIG_HEAD`, `MERGE_HEAD`, `REBASE_HEAD`, `REVERT_HEAD`, `CHERRY_PICK_HEAD`, `BISECT_HEAD`, `AUTO_MERGE` | Last fetched; before the last big HEAD move (merge, rebase, reset, am); being merged; where rebase stopped; being reverted; being cherry-picked; being tested by `bisect --no-checkout`; tree ort wrote during a conflicted merge. |
| `@` | `HEAD`. |
| `<ref>@{<date>}` | Where the local ref pointed then (`main@{yesterday}`, `HEAD@{5 minutes ago}`). Needs a reflog. |
| `<ref>@{<n>}` | Where the ref pointed `<n>` changes ago (`main@{1}`). Needs a reflog. |
| `@{<n>}` | Same, for the current branch. |
| `@{-<n>}` | `<n>`th previous branch or commit checked out (`@{-1}` = last). |
| `<branch>@{upstream}`, `@{u}` | Branch it tracks (`origin/main`). Case-insensitive. |
| `<branch>@{push}`, `@{push}` | Where `git push` would send it. Same as `@{u}` unless push and pull targets differ. |
| `<rev>^`, `<rev>^<n>` | First (or `<n>`th) parent. `<rev>^0` = the commit itself. |
| `<rev>~`, `<rev>~<n>` | `<n>` generations back via first parents. `HEAD~3` = `HEAD^^^`. |
| `<rev>^{<type>}` | Peel tags etc. to `<type>` (`commit`, `tree`, `blob`, `tag`, `object`). `^{object}` checks existence; `^{tag}` checks it's a tag. |
| `<rev>^{}` | Peel tags to a non-tag. |
| `<rev>^{/<regex>}` | Newest commit reachable from `<rev>` with matching message. |
| `:/<regex>` | Newest commit reachable from any ref with matching message. `:/!-foo` = doesn't match; `:/!!` = literal `!`. |
| `<rev>:<path>` | File or directory at `<path>` in that commit. `./`, `../` are relative to the current directory. |
| `:<n>:<path>`, `:<path>` | File in the index at stage `<n>` (default 0). During a merge: 1 = base, 2 = ours, 3 = theirs. |

Parents, where B and C are A's parents (B first):

```
G   H   I   J
 \ /     \ /
  D   E   F
   \  |  / \
    \ | /   |
     |/    |
      B     C
       \   /
        \ /
         A
```

`B = A^ = A~1`, `C = A^2`, `D = A^^ = A~2`, `E = B^2`, `F = B^3`, `G = A~3`, `H = D^2 = A~2^2`, `I = F^`, `J = F^2`.

## Ranges

| Syntax | Selects |
|--------|---------|
| `<rev>` | `<rev>` and all its ancestors. |
| `^<rev>` | Exclude `<rev>` and its ancestors. |
| `A..B` | Reachable from B but not A. Missing side = `HEAD`: `origin..` = my commits not in origin. |
| `A...B` | Reachable from either, not both. Missing side = `HEAD`. |
| `<rev>^@` | `<rev>`'s parents and ancestors, not `<rev>`. |
| `<rev>^!` | Only `<rev>`, excluding its parents. |
| `<rev>^-<n>` | `<rev>^<n>..<rev>` (default `<n>` = 1). For a merge, `M^-` = commits it brought in. |

Ranges combine: `A..B C..D` = reachable from B or D, but from neither A nor C.

## --parseopt

Parse a shell script's options like git built-ins do. Feed an option spec on stdin; it prints normalized arguments for `eval`. Bad option: prints usage, exits 129. `-h`/`--help`: exits 129, or 0 in 2.56+. Quote the result.

```sh
OPTS_SPEC="\
my-script [<options>] <args>...
my-script does foo and bar!
--
h,help!   show the help
foo       some nifty option --foo
bar=      some cool option --bar with an argument
baz=arg   another option --baz with a named argument
qux?path  qux takes an optional path argument
 An option group header
C?        option C with an optional argument"
eval "$(echo "$OPTS_SPEC" | git rev-parse --parseopt -- "$@" || echo exit $?)"
```

Spec: usage lines, a line with `--`, then one line per option: `<short>,<long><flags><arg-hint> <help>`.

| Flag char | Meaning |
|-----------|---------|
| `=` | Takes a value. |
| `?` | Takes an optional value (pair with `--stuck-long`). |
| `*` | Hide from `-h` (shown by `--help-all`). |
| `!` | No `--no-` form. |

`<arg-hint>` names the value in help. Lines that don't match (start with a space to force it) become group headers.

## --sq-quote

`args=$(git rev-parse --sq-quote "$@")` gives a safely quoted string for `eval`.
