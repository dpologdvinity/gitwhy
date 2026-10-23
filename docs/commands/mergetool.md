# git mergetool

Open each conflicted file in a merge tool (vimdiff, meld, kdiff3, ...) to resolve it. Default: all conflicted files; a directory means all in it.

Leaves `<file>.orig` backups; delete them or set `mergetool.keepBackup=false`.

**Usage:** `git mergetool [<flags>] [<file>...]`

## Flags

| Flag | What it does |
|------|--------------|
| `-t <tool>`, `--tool=<tool>` | Tool, e.g. `vimdiff`, `meld`, `kdiff3`, `emerge`, `tortoisemerge`. Default: `merge.tool`, else guessed. |
| `--tool-help` | List tools and which are installed. |
| `-y`, `--no-prompt` | Don't ask before each file. Default when a tool is set. |
| `--prompt` | Ask before each file. |
| `-g`, `--gui` | Use `merge.guitool` (fallback `merge.tool`). |
| `--no-gui` | Use `merge.tool`, overriding `--gui` and `mergetool.guiDefault`. |
| `-O<orderfile>` | File order from glob patterns in `<orderfile>`. Overrides `diff.orderFile`; `-O/dev/null` cancels. |

## Custom tools

Set `mergetool.<tool>.cmd` to a shell command. Git provides:

| Variable | File |
|----------|------|
| `$BASE` | Temp copy of the common ancestor (if any). |
| `$LOCAL` | Temp copy of your version. |
| `$REMOTE` | Temp copy of the incoming version. |
| `$MERGED` | The real file; write the result here. |

`mergetool.<tool>.trustExitCode=true`: trust the tool's exit code; otherwise git checks if the file changed or asks. `mergetool.<tool>.path`: tool location if not on `PATH`.

## vimdiff

4 windows: `LOCAL`, `BASE`, `REMOTE` (top, read-only) and `MERGED` (bottom, edit). `:wq` saves; `:cq` aborts. `gvimdiff` and `nvimdiff` use gVim and Neovim.

Layout: `mergetool.vimdiff.layout` (or `mergetool.gvimdiff.layout`, `mergetool.nvimdiff.layout`). Separators, strongest first: `@` buffer to save (default `MERGED`), `+` tab, `/` horizontal split, `,` vertical split. Parentheses group. Default: `(LOCAL,BASE,REMOTE)/MERGED`.

Shortcuts: suffix `1`, `2`, or `3` (`vimdiff2`, `nvimdiff1`, ...).

| Suffix | Layout |
|--------|--------|
| `1` | `@LOCAL, REMOTE` |
| `2` | `LOCAL, MERGED, REMOTE` |
| `3` | `MERGED` |
