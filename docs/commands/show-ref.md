# git show-ref

List refs with their IDs (`<id> <ref>` per line), or check a ref exists. For scripts. Exits 1 if nothing matches.

Patterns match whole trailing name parts: `main` matches `refs/heads/main` and `refs/remotes/origin/main`, not `refs/heads/mymain`.

**Usage:** `git show-ref [--head] [-d] [-s[=<n>]] [--abbrev[=<n>]] [--branches] [--tags] [--] [<pattern>...]`
or `git show-ref --verify [-q] [-d] [-s[=<n>]] [--abbrev[=<n>]] [--] [<ref>...]`
or `git show-ref --exclude-existing[=<pattern>]`
or `git show-ref --exists <ref>`

## Flags

| Flag | What it does |
|------|--------------|
| `<pattern>...` | Only refs ending in these parts. |
| `--[no-]head` | Also show `HEAD`. |
| `--[no-]branches` | Branches only. With `--tags`: both. |
| `--heads` | Deprecated: `--branches`. |
| `--[no-]tags` | Tags only. With `--branches`: both. |
| `-d`, `--[no-]dereference` | Also show annotated tags' targets as `<ref>^{}`. |
| `-s`, `--[no-]hash[=<n>]` | IDs only (`<n>` shortens). |
| `--[no-]abbrev[=<n>]` | IDs at least `<n>` hex digits. |
| `--[no-]verify` | Arguments must be exact full ref names. Errors on missing unless `-q`. |
| `-q`, `--[no-]quiet` | No output, exit code only. Use with `--verify`. |
| `--[no-]exists` | Check one ref exists. Exit 0: yes. 2: missing. 1: error. |
| `--exclude-existing[=<pattern>]` | Filter stdin `[<anything> ]<ref>[^{}]` lines, printing refs that *don't* exist here. Skips refs not starting with `<pattern>`. |
| `--` | Rest are patterns or refs. |
