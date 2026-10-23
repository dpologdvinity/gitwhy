# git subtree

Keep another project in a subdirectory, with or without its history, and move changes both ways: merge upstream updates in, or split the folder's changes out as a separate history to push. Unlike submodules, it's a plain folder: no `.gitmodules`, nothing special for cloners. Optional `contrib` add-on; may not be installed.

**Usage:** `git subtree <subcommand> -P <prefix> [<flags>] <args>`

Tip: commit subproject changes separately so split messages make sense.

## Flags

All subcommands. `-P` required.

| Flag | What it does |
|------|--------------|
| `-h`, `--help` | Help. |
| `-q`, `--quiet` | Less stderr output. |
| `-d`, `--debug` | More stderr output. |
| `-P`, `--[no-]prefix=<prefix>` | Subtree folder path. |
| `-S`, `--[no-]gpg-sign[=<key-id>]` | GPG-sign new commits. Key defaults to committer. |

## Subcommands

| Subcommand | What it does |
|------------|--------------|
| `add` | Create the folder from a commit or remote ref, joining histories with a commit. |
| `merge` | Merge newer subproject changes into the folder. |
| `split` | Make a history of only the folder (contents at root); print its tip ID. |
| `pull` | Fetch a remote ref, then `merge`. |
| `push` | `split`, then push to a remote ref. |

### git subtree add

`git subtree add -P <prefix> [<flags>] <local-commit>` or `git subtree add -P <prefix> [<flags>] <repository> <remote-ref>`

Example: `git subtree add --prefix=vendor/lib --squash https://example.com/lib.git main`.

| Flag | What it does |
|------|--------------|
| `--[no-]squash` | Import one commit of the current state, not full history. |
| `-m`, `--message=<message>` | Merge commit message. |

### git subtree merge

`git subtree merge -P <prefix> [<flags>] <local-commit> [<repository>]`. Keeps local folder changes; resolve conflicts as usual. If an earlier `--squash` merge used an annotated tag, it must exist locally; `<repository>` fetches it.

| Flag | What it does |
|------|--------------|
| `--[no-]squash` | Merge as one commit. Can go backwards (v2.5 to v2.4): folder becomes exactly that version. |
| `-m`, `--message=<message>` | Merge commit message. |

### git subtree split

`git subtree split -P <prefix> [<flags>] [<local-commit>] [<repository>]`. Default `HEAD`. Keeps only commits touching the folder. Deterministic: same input and flags give same IDs, so re-splits extend the last one. `<repository>`: fetch a needed tag (see `merge`).

Example: `git subtree split -P lib --annotate="(split) " -b lib-only`, then `git push <new-repo> lib-only:main`.

| Flag | What it does |
|------|--------------|
| `--[no-]annotate=<annotation>` | Prefix new commit messages. Keep it the same every split, or IDs change and merging breaks. |
| `-b`, `--branch=<branch>` | Create a new branch (must not exist) at the result. |
| `--[no-]ignore-joins` | Rebuild full history, ignoring past `--rejoin`s. Slow. |
| `--[no-]onto=<onto>` | Folder not made by `subtree add`: ID of the first imported subproject commit to build from. |
| `--[no-]rejoin` | Merge the split back in, speeding later splits and easing upstream merges. Commits then show twice in `git log`. Add `--squash` if you always squash. |
| `--[no-]squash` | With `--rejoin`: squash the rejoin merge. |
| `-m`, `--message=<message>` | With `--rejoin`: rejoin merge message. |

### git subtree pull

`git subtree pull -P <prefix> [<flags>] <repository> <remote-ref>`

| Flag | What it does |
|------|--------------|
| `--[no-]squash` | Merge as one commit. |
| `-m`, `--message=<message>` | Merge commit message. |

### git subtree push

`git subtree push -P <prefix> [<flags>] <repository> [+][<local-commit>:]<remote-ref>`. Splits `<local-commit>` (default `HEAD`) and pushes. Leading `+` ignored.

| Flag | What it does |
|------|--------------|
| `--[no-]annotate=<annotation>` | As in `split`. |
| `-b`, `--branch=<branch>` | As in `split`. |
| `--[no-]ignore-joins` | As in `split`. |
| `--[no-]onto=<onto>` | As in `split`. |
| `--[no-]rejoin` | As in `split`. |
| `--[no-]squash` | With `--rejoin`: squash the rejoin merge. |
| `-m`, `--message=<message>` | With `--rejoin`: rejoin merge message. |
