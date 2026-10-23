# git read-tree

Low-level: load a tree into the index. Files untouched unless `-u`. With `-m`, merges 2 or 3 trees into the index, resolving trivial cases and leaving conflicts unmerged. `git checkout` and `git merge` do this for you.

**Usage:** `git read-tree [(-m [--trivial] [--aggressive] | --reset | --prefix=<prefix>) [-u | -i]] [<flags>] (--empty | <tree-ish1> [<tree-ish2> [<tree-ish3>]])`

## Flags

| Flag | What it does |
|------|--------------|
| `<tree-ish>` | One to three trees or commits. See [Merging](#merging). |
| `-m` | Merge. Refuses if unmerged entries remain. |
| `--[no-]reset` | Like `-m`, but drop unmerged entries. With `-u`, overwrites local changes and untracked files. |
| `-u` | Update working files after a successful merge. |
| `-i` | Skip the working-tree check. For merging into a temp index (`GIT_INDEX_FILE`). |
| `-n`, `--[no-]dry-run` | Check for failure; change nothing. |
| `-v`, `--[no-]verbose` | Checkout progress. |
| `--[no-]trivial` | Three trees: merge only if no file needs a content merge. |
| `--[no-]aggressive` | Three trees: also resolve one-side delete (delete), both delete (delete), identical adds (add). |
| `--prefix=<prefix>/` | Add the tree under `<prefix>/` to the current index. Won't overwrite entries. For subtree merges. |
| `--index-output=<file>` | Write result to `<file>` (index stays locked). Must be on the index's filesystem, in a writable folder. |
| `--exclude-per-directory=<gitignore>` | Obsolete. Only `.gitignore`, only with `-u`. |
| `--[no-]recurse-submodules[=<checkout>]` | Update active submodules to recorded commits (detached `HEAD`). |
| `--no-sparse-checkout` | Ignore sparse checkout even if `core.sparseCheckout`. |
| `--sparse-checkout` | Apply sparse checkout (default). |
| `--[no-]empty` | Empty the index. |
| `-q`, `--[no-]quiet` | No feedback. |
| `--[no-]debug-unpack` | Print internal unpack details. |

## Merging

With `-m`, tree count picks the merge type.

**One tree** (`read-tree -m <tree>`): plain read that keeps cached stat info for unchanged files, so `git checkout-index -f -u -a` rewrites only real changes and `git diff-files` shows none falsely.

**Two trees** (`read-tree -m <H> <M>`): fast-forward the index from current `<H>` to newer `<M>`, keeping local changes; fails if any would be lost. Per path (I = index):

| | I | H | M | Result |
|-|---|---|---|--------|
| 1 | none | none | exists | use M |
| 2 | none | exists | none | remove from index |
| 3 | none | exists | H == M | use M if the index is empty (first checkout), else keep index |
| 3 | none | exists | H != M | fail |
| 4-7 | exists | none | none, or same as I | keep index |
| 8-9 | exists | none | differs from I | fail |
| 10 | clean, same as H | exists | none | remove from index |
| 11-13 | other | exists | none | fail |
| 14-15 | any | H == M | | keep index |
| 16-17 | differs from both | H != M | | fail |
| 18-19 | same as M | H != M | | keep index |
| 20 | clean, same as H | H != M | | use M |
| 21 | dirty, same as H | H != M | | fail |

"Clean": working file matches index. In "keep index" cases with `-u`, stale working files stay. Afterwards `git diff-index --cached <M>` shows carried-over local changes (except cases 18-19).

**Three trees** (`read-tree -m <base> <ours> <theirs>`): fills stages 1 (base), 2 (ours), 3 (theirs). Collapses to stage 0 when ours equals theirs, or only one side changed (take it). Others stay unmerged; resolve with `git merge-index git-merge-one-file -a`. `git write-tree` refuses until resolved. Existing index entries must match `<ours>`. Fails before changing anything if it would overwrite local changes; others are left alone. Full manual merge:

```sh
git read-tree -m -u $(git merge-base $JC $LT) $JC $LT
git merge-index git-merge-one-file -a
echo "Merge with Linus" | git commit-tree $(git write-tree) -p $JC -p $LT
```

## Sparse checkout

Prefer [`git sparse-checkout`](sparse-checkout.md). With `core.sparseCheckout` on, `read-tree` reads `$GIT_DIR/info/sparse-checkout` (gitignore-style patterns). Matching entries, or those with files on disk, are checked out; the rest get skip-worktree and are removed from disk. Patterns can exclude: `/*` then `!unwanted`. To restore all files, set it to `/*` and rerun `read-tree` before disabling.
