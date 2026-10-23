# gitwhy research notes

Researched 2026-10-22. Local machine: git 2.55.0, git-lfs 3.4.1.

## 1. Git version and command list

- Latest git release: **2.56.0**. It is the latest tag on github.com/git/git, and git-scm.com/docs/git says "last updated in 2.56.0". GitHub Blog highlights post is dated 2026-09-28.
- New in 2.56, per GitHub Blog: `git history drop` and `git refs create|update|delete|rename`. These are subcommands, not new top-level commands. The blog does not make clear which `git refs` subcommands are new in 2.56.
- 2.55 added `git history fixup`. The `git history` command is still marked experimental.
- git-scm.com/docs/git lists 160 commands. Local `git --list-cmds=main,nohelpers` gives 164.

### Category layout on git-scm.com/docs/git (useful for grouping in gitwhy)

Main porcelain; Ancillary manipulators; Ancillary interrogators; Interacting with others; Low-level manipulation; Low-level interrogation; Syncing repositories; Syncing helpers; Internal helpers.

git-scm.com/docs (the reference index) uses a beginner-friendly grouping instead: Setup and Config, Getting and Creating Projects, Basic Snapshotting, Branching and Merging, Sharing and Updating, Inspection and Comparison, Patching, Debugging, Email, External Systems, Administration, Server Admin, Plumbing.

### On git-scm.com, missing locally (14)

`archimport citool cvsexportcommit cvsimport cvsserver gitk gitweb gui p4 scalar send-email sh-i18n sh-setup svn`

- Most of these come in separate distro packages: git-gui (`gui`, `citool`), gitk, gitweb, git-svn, git-cvs, git-email (`send-email`). `archimport` and `p4` are often not packaged at all. Likely cause: the Debian/Ubuntu package split (not verified one by one).
- `sh-i18n` and `sh-setup` are shell script libraries. `nohelpers` filters them out.
- `scalar` is a separate binary. It is not installed under `git-core` here.

### Local, not in the git-scm.com command list (18)

`fsck-objects init-db merge-octopus merge-ours merge-recursive merge-recursive-ours merge-recursive-theirs merge-resolve merge-subtree pickaxe remote-ext remote-fd remote-ftp remote-ftps remote-http remote-https stage subtree`

- Aliases or legacy names: `fsck-objects` (= fsck), `init-db` (= init), `stage` (= add), `pickaxe` (= blame, old builtin alias).
- Merge strategy backends: the `merge-*` entries. Users reach them through `git merge -s <strategy>`.
- Remote helpers: the `remote-*` entries. Git runs them internally for URL schemes. Man pages exist for `remote-ext` and `remote-fd`.
- `subtree` is from contrib and is shipped by Debian. It has no entry in the official command list.
- Suggestion: gitwhy should map aliases to their main command and mark the rest as "internal / you rarely type this".

## 2. git-scm.com cheat sheet (input for "recipes")

Structure: 17 sections. Each item is a goal phrase followed by commands, with "OR" for alternatives. Some sections have before/after diagrams.

1. Getting Started: start a repo, clone.
2. Prepare to Commit: add a file, add all, `add -p`, `mv`, `rm`, `rm --cached`, unstage one/all (`reset` OR `restore --staged`), `status`.
3. Make Commits: `commit`, `-m`, `-am`.
4. Move Between Branches: switch, create, list, sort by recent, delete, force delete.
5. Diff Staged/Unstaged: `diff HEAD`, `diff --staged`, `diff`.
6. Diff Commits: `show`, diff two commits, diff one file, `--stat`.
7. Ways to refer to a commit: branch, tag, ID, remote branch, `HEAD`, `HEAD~3`/`HEAD^^^`. Reference table only, no commands.
8. Discard Your Changes: restore a file, restore staged+worktree, `reset --hard`, `clean`, `stash`.
9. Edit History: undo the last commit (`reset HEAD^`), squash (`rebase -i`), undo a failed rebase (`reflog` then `reset --hard`), `commit --amend`.
10. Code Archaeology: `log`, `--graph`, `--oneline`, `log <file>`, `--follow`, `log -G`, `blame`.
11. Combine Diverged Branches: rebase, merge, squash merge, fast-forward, cherry-pick. Has diagrams.
12. Restore an Old File: `checkout <commit> <file>` OR `restore --source`.
13. Add a Remote: `remote add`.
14. Push: push to origin, push to tracking branch, `-u`, `--force-with-lease`, `--tags`.
15. Pull: `fetch`, `pull --rebase`, `pull`.
16. Configure Git: `config`, `--global`, alias, `man git-config`.
17. Important Files: `.git/config`, `~/.gitconfig`, `.gitignore`.

- Gaps that gitwhy recipes could fill: stash pop/list, tags (create/delete), resolving merge conflicts, `revert` vs `reset`, recovering a deleted branch, `.gitignore` for files already tracked, `bisect`, `worktree`, `git lfs`.
- The page itself has no explicit license statement. Treat it as reference only and write recipes in our own words.

## 3. Licensing

- **git (code + Documentation/)**: GPL-2.0 only (COPYING, Linus's preamble: "v2, not v2.2 or v3.x"). Documentation/ has no separate license file. Debian's copyright file lists `Files: *` as GPL-2 (local `/usr/share/doc/git/copyright`). There is no explicit "docs license" statement upstream, so this is an inference: the docs are covered by GPL-2.0.
- **git-scm.com site**: site source is MIT. Graphical assets are CC BY 3.0 (git/git-scm.com README). The reference pages on the site are rendered from git's own GPL-2.0 Documentation.
- **Pro Git book** (linked from the site): CC BY-NC-SA 3.0. The NonCommercial and ShareAlike terms make it a bad source to copy from.
- **What this means for gitwhy**:
  - Original short descriptions written in our own words are our own work. GPL does not apply to them. Command names, flag names and facts about behavior are not copyrightable.
  - Do not paste or lightly paraphrase sentences from git man pages. Copied text would make that data a derivative of GPL-2.0 material. Then the data (and arguably the bundle) would need GPL-2.0-compatible licensing.
  - Do not copy cheat-sheet or Pro Git wording.
  - Reading the local man pages at runtime (for example, a "show the official text" link) is fine, because we do not redistribute them.
  - Keep a note in the repo that the descriptions are original. This is not legal advice.
- **Others, for reference**: tldr pages are CC BY 4.0 (reuse is allowed with attribution). explainshell is GPL-3.0. lazygit is MIT. git-lfs is MIT.

## 4. Competitors / prior art

| Tool | What it does | How gitwhy differs | Stars |
|---|---|---|---|
| lazygit | TUI for doing git operations | It runs git; it does not teach or explain flags | 83.4k |
| gitui | Rust TUI git client | Same as lazygit | 22.6k |
| tig | Text-mode browser for history/diffs | Viewer, no explanations | 13.4k |
| GitKraken | Commercial GUI client | Proprietary GUI, not offline per-flag docs | n/a |
| tldr-pages | Community example cheatsheets for many CLIs, including git subcommands | A few examples per command, not every flag; not git-specific; no "why" | 64.0k |
| explainshell | Web app that matches command-line args to man page text | Uses raw man text, not plain English; web-hosted; GPL-3.0 | 14.3k |
| git-extras | Extra git subcommands (summary, changelog, ...) | Adds commands, does not explain them | 18.1k |
| jj (Jujutsu) | Git-compatible VCS with `jj undo` / an operation log | Replaces the workflow; does not explain git | 32.2k |
| gitless | Simplified VCS on top of git | Different VCS; low activity | 2.0k |
| Learn Git Branching | Interactive browser tutorial with visualizations | Tutorial, not a lookup tool; web | 34.2k |
| ohshitgit.com | Plain-English "I messed up, how do I fix it" recipes | Closest in tone; a short static page, not a CLI; repo has no license (NOASSERTION) | 1.6k (ksylor/ohshitgit) |
| git-flight-rules | Big guide of "what to do when X" | Recipe content for goal-to-commands; docs repo, not a tool | 42.6k |
| ugit (Bhupesh-V) | CLI to undo the last git command (20+ scenarios) | Undo only | 1.5k |
| git-undo (zhangela etc.) | Small undo helpers | Undo only, small | 51 and less |
| git-explain (nazarli-shabnam) | Suggests add/commit, optional AI | Despite the name, not an explainer | 43 |
| git-onboard (jarmstrong158) | Beginner setup walkthrough explaining commands | Setup only, tiny | 0 |

- None of the tools found is an offline, no-AI CLI/TUI that explains every git command and flag in plain English. tldr and explainshell come closest on lookup. ohshitgit and git-flight-rules come closest on recipes. This is a search result, not an exhaustive survey.

### Name "gitwhy" availability

- **GitHub**: taken several times (22 repos named gitwhy). Most similar is a cluster of "git blame tells you who, gitwhy tells you why" code-provenance tools: mehrtam/gitwhy (6 stars), adityakumarroy67/gitwhy (Gemini), surajsrivastav/gitwhy (Go module), Wladastic/GitWhy. hexapode/git-why (22 stars) and Sameer988/git-why are also code-provenance tools.
- **npm**: `gitwhy` is free. `git-why` is taken (0.1.4, AI git-history explainer).
- **PyPI**: `gitwhy` is free. `git-why` is taken (0.1.0, code-explainer from git history).
- **crates.io**: `gitwhy` and `git-why` are both free.
- **Homebrew core**: `gitwhy` and `git-why` are both free (formula and cask).
- **Go**: `github.com/surajsrivastav/gitwhy` exists on pkg.go.dev. Go module paths are namespaced by owner, so this only matters for search confusion.
- Risk: the name is crowded and it already means "why does this line of code exist" (a blame/provenance meaning). Consider a distinct name, or own it clearly with a tagline like "explains git commands".

## 5. Git LFS

- Latest: **v3.8.1**, released 2026-10-22 (security fixes for 4 CVEs). Before it: v3.8.0 (2026-08-28), v3.7.1 (2025-10-17). The local 3.4.1 is several minors behind.
- Official man pages in `docs/man` on the main branch. The subcommands are:
  `checkout clean clone completion dedup env ext fetch filter-process fsck install lock locks logs ls-files merge-driver migrate pointer post-checkout post-commit post-merge pre-push prune pull push smudge standalone-file status track uninstall unlock untrack update`
  Upstream also has `git-lfs-config.adoc` (section 5, `.lfsconfig`) and `git-lfs-faq.adoc` (section 7). These are not subcommands.
- Local `/usr/share/man/man1/git-lfs-*`: the same 32 subcommands plus `git-lfs.1`. There is no difference in the subcommand set. `git-lfs-config.5` is in man5. `git-lfs-faq` is not installed locally, because it was added after 3.4.1 (inferred).
- Not checked: flag-level changes between 3.4.1 and 3.8.1. Read the CHANGELOG before writing flag descriptions.
- Internal or hook subcommands that gitwhy should mark as "git runs this, not you": `clean smudge filter-process merge-driver post-checkout post-commit post-merge pre-push standalone-file`.

## Unconfirmed

- The exact cause of each "missing locally" command (assumed to be the distro package split).
- Whether git upstream explicitly states a license for Documentation/ (inferred GPL-2.0 from COPYING and Debian's copyright file).
- Star counts are from the GitHub API on 2026-10-22. GitKraken has no public repo.
- Competitor search was GitHub plus known tools only. App stores, VS Code extensions and paid courses were not surveyed.

## Sources

- https://git-scm.com/docs/git
- https://git-scm.com/docs
- https://git-scm.com/cheat-sheet
- https://github.com/git/git (tags; COPYING: https://raw.githubusercontent.com/git/git/master/COPYING)
- https://github.blog/open-source/git/highlights-from-git-2-56/
- https://github.blog/open-source/git/highlights-from-git-2-55/
- https://github.com/git/git-scm.com (README "License" section)
- https://github.com/progit/progit2 (LICENSE.asc)
- /usr/share/doc/git/copyright (local Debian package)
- https://github.com/git-lfs/git-lfs/releases
- https://github.com/git-lfs/git-lfs/tree/main/docs/man
- https://github.com/jesseduffield/lazygit, https://github.com/gitui-org/gitui, https://github.com/jonas/tig
- https://github.com/tldr-pages/tldr, https://github.com/idank/explainshell, https://github.com/tj/git-extras
- https://github.com/jj-vcs/jj, https://github.com/gitless-vcs/gitless, https://github.com/pcottle/learnGitBranching
- https://ohshitgit.com, https://github.com/ksylor/ohshitgit, https://github.com/k88hudson/git-flight-rules
- https://github.com/Bhupesh-V/ugit, https://github.com/zhangela/git-undo
- https://www.gitkraken.com
- https://github.com/search?q=gitwhy, https://www.npmjs.com/package/git-why, https://pypi.org/project/git-why/
- https://crates.io/api/v1/crates/gitwhy, https://formulae.brew.sh/api/formula/gitwhy.json, https://pkg.go.dev/search?q=gitwhy
