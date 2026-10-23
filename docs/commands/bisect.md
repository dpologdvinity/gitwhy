# git bisect

Binary-search history for the commit that introduced a bug. Mark one bad and one good commit; git checks out the midpoint, you mark it, repeat until git names the first bad commit (`refs/bisect/bad`). `git bisect run` automates it.

Works for any change: use `old`/`new` or your own terms instead of `good`/`bad`. Don't mix the two pairs in one session.

Example:

```
git bisect start
git bisect bad              # current commit is broken
git bisect good v2.6.13     # this one worked
# test each commit git picks, then: git bisect good / git bisect bad
git bisect reset            # done: go back
```

**Usage:** `git bisect <subcommand> [<flags>] [<args>]`

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `start` | Begin a session, optionally with bad and good commits. |
| `bad`, `new`, `<term-new>` | Mark a commit as having the change. |
| `good`, `old`, `<term-old>` | Mark commits as not having it. |
| `terms` | Print the terms in use. |
| `skip` | Mark commits untestable; git picks another. |
| `next` | Check out the next commit to test. |
| `reset` | End the session and go back. |
| `visualize`, `view` | Show remaining suspects. |
| `log` | Print marks so far as a replayable script. |
| `replay` | Re-run a saved `log` file. |
| `run` | Test each commit automatically with a script. |
| `help` | Print usage. |

### git bisect start

`git bisect start [<flags>] [<bad> [<good>...]] [--] [<pathspec>...]`. Clears any old session.

| Flag | What it does |
|------|--------------|
| `<bad> [<good>...]` | Bad commit, then known-good commits. More good = fewer tests. |
| `[--] <pathspec>...` | Only consider commits touching these paths. |
| `--term-new=<term>`, `--term-bad=<term>` | Your word for "has the change" (e.g. `slow`). Not a subcommand name. |
| `--term-old=<term>`, `--term-good=<term>` | Your word for "before the change" (e.g. `fast`). |
| `--no-checkout` | Don't check out test commits; just move `BISECT_HEAD`. Automatic in bare repos. |
| `--first-parent` | Follow only first parents of merges, so a merge is blamed, not commits inside it. |
| `--reset-when-found[=<where>]` | (2.56+) End the session once found. `<where>`: `original` (default, go back) or `found` (stay). Not with `--no-checkout`. |

### git bisect bad / new

`git bisect (bad | new | <term-new>) [<rev>]`. Mark `<rev>` (default: current) as having the change. No flags.

### git bisect good / old

`git bisect (good | old | <term-old>) [<rev>...]`. Mark commits (default: current) as not having it. No flags.

### git bisect terms

`git bisect terms [<flag>]`. No flag: print both.

| Flag | What it does |
|------|--------------|
| `--term-good`, `--term-old` | Print only the "before" term. |
| `--term-bad`, `--term-new` | Print only the "after" term. |

### git bisect skip

`git bisect skip [(<rev> | <range>)...]`. Skip untestable commits (default: current). `v2.5..v2.6` skips after `v2.5` through `v2.6`. Skipping near the culprit may leave the answer ambiguous. No flags.

### git bisect next

No flags.

### git bisect reset

`git bisect reset [<commit>]`. Return to the pre-`start` commit, or `<commit>` (`bisect/bad` for the culprit, `HEAD` to stay). No flags.

### git bisect visualize / view

`git bisect visualize [<log-options>]`. Opens `gitk` in a graphical session, else `git log`. Takes `git log` flags like `-p`, `--stat`.

### git bisect log

No flags. Fix a wrong mark: save the log, edit it, `git bisect reset`, `git bisect replay file`.

### git bisect replay

`git bisect replay <logfile>`. No flags.

### git bisect run

`git bisect run [--reset-when-found[=<where>]] <cmd> [<arg>...]`. Run `<cmd>` on each commit; exit code marks it. Keep the script outside the repo.

- `0`: good/old.
- `1`–`127` except `125`: bad/new.
- `125`: skip.
- Anything else (e.g. `255`): stop bisecting.

Example: `git bisect start HEAD v1.2 -- && git bisect run sh -c "make || exit 125; ./test.sh"`. A script that patches code must undo it before exiting.

| Flag | What it does |
|------|--------------|
| `--reset-when-found[=<where>]` | (2.56+) As on `start`: end once found. `<where>`: `original` (default) or `found`. |

### git bisect help

No flags. Same as `git bisect -h`.
