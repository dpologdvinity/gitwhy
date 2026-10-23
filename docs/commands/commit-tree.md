# git commit-tree

Low-level: make a commit object from a tree and print its ID. Moves no branch (use `git update-ref`). For normal commits use [`git commit`](commit.md).

Message from stdin unless `-m`/`-F`. Author and committer from config or `GIT_AUTHOR_*`/`GIT_COMMITTER_*`. No parents: root commit; two or more: merge.

**Usage:** `git commit-tree [-p <parent>]... [-S[<keyid>]] [-m <message>]... [-F <file>]... <tree>`

## Flags

| Flag | What it does |
|------|--------------|
| `<tree>` | Tree object ID to commit. |
| `-p <parent>` | Add a parent. Repeat for a merge. |
| `-m <message>` | Add a message paragraph. Repeatable; mixes with `-F` in order. |
| `-F <file>` | Read a message paragraph from `<file>` (`-` = stdin). Repeatable. |
| `-S[<keyid>]`, `--[no-]gpg-sign[=<keyid>]` | GPG-sign. Key defaults to committer; no space after `-S`. |

## Date formats

`GIT_AUTHOR_DATE` and `GIT_COMMITTER_DATE` accept:

| Format | Example |
|--------|---------|
| Git internal: seconds since 1970 plus UTC offset. Prefix `@` (required under 9 digits). | `@1112911993 +0200` |
| RFC 2822 | `Thu, 07 Apr 2005 22:13:13 +0200` |
| ISO 8601 (space for `T` OK). Date also as `YYYY.MM.DD`, `MM/DD/YYYY`, `DD.MM.YYYY`. | `2005-04-07T22:13:13` |
