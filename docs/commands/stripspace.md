# git stripspace

Clean stdin text like git cleans commit messages: strip trailing whitespace, squeeze blank-line runs, drop leading/trailing blank lines, add a final newline. All-whitespace input gives no output. For source files, use `git apply --whitespace=fix`.

**Usage:** `git stripspace [-s | -c] < <file>`

## Flags

| Flag | What it does |
|------|--------------|
| `-s`, `--strip-comments` | Also delete comment lines (`core.commentChar`, default `#`). |
| `-c`, `--comment-lines` | Instead, prefix each line with the comment char and a space (char only on empty lines). |
