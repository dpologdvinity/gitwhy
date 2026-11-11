# gitwhy: Next Phases Implementation Plan

> **Status:** done. Every phase (A–E) shipped in v0.7.0 instead of the per-phase versions below. Kept as a record of the plan; the checkboxes were not ticked during execution.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the post-v0.3.0 roadmap: hunk/line staging and a commit graph in the TUI, an optional shell hook so plain `git` commands get undo points, Windows/Git Bash support, and a Homebrew tap.

**Architecture:** gitwhy is a Go CLI in `~/git/gitwhy` that runs the real `git`. The engine (`internal/snapshot`, `preview`, `danger`, `recipe`, `explain`) has no UI. `internal/ui` is the huh-based CLI/wizard; `internal/tui` is the Bubble Tea full-screen home that returns an `Action` the UI runs. New work follows the same split: pure logic in small packages with unit tests, thin UI on top.

**Tech Stack:** Go (module `github.com/dpologdvinity/gitwhy`, `go 1.24` in go.mod, toolchain at `~/.local/go/bin`), `charm.land/bubbletea/v2`, `charm.land/huh/v2`, `charm.land/lipgloss/v2`, `github.com/charmbracelet/x/ansi`, `gopkg.in/yaml.v3`. Tools: `golangci-lint` v2.15, `goreleaser` v2.19, `vhs` (all via Linuxbrew), `gh`.

**Spec:** `docs/superpowers/specs/2026-10-22-gitwhy-v1-design.md` (v1 design; its "TUI phase" and "Later" lists are the source of this roadmap). The user later said the TUI does not need to look like lazygit; the existing home screen (`internal/tui/tui.go`) is the design to extend.

## Roadmap (do in this order; each phase ends shippable)

| Phase | Tasks | Ships as |
|---|---|---|
| A. TUI hunk & line staging | 1–3 | v0.4.0 |
| B. TUI commit graph | 4–5 | v0.5.0 |
| C. Shell hook: undo points for plain `git` | 6–8 | v0.6.0 |
| D. Windows / Git Bash | 9–10 | v0.7.0 |
| E. Homebrew tap (needs a token from the user) | 11 | with the next tag |

Stop after any phase if the user says so. Not in this plan (ask the user first; outward-facing): an upstream PR to `atotto/clipboard`, promotion posts.

## Global Constraints

- No AI, LLM, network service or API key at runtime.
- Every repo change gitwhy makes goes through a snapshot (`snapshot.Take` + `snapshot.Finish`) **except** staging/unstaging, which is always reversible and runs directly (existing rule in `internal/tui/tui.go` `quick`).
- git floor 2.38 (`checkGit` in `cmd/gitwhy/main.go`). Do not use newer git features without a fallback.
- `go.mod` keeps `go 1.24` and the `replace github.com/atotto/clipboard => ./third_party/clipboard` line.
- Never `git push` or push tags unless the user asked for it in the current session. Committing locally is fine.
- Keep CPU and cost low: run the narrowest tests for each task (`go test ./internal/<pkg>/ -run <Name>`). Run the full suite (`go test ./...`, ~90 s) and `golangci-lint run ./...` once at the end of each phase. At most 2 subagents at a time.
- Put Go on PATH in every shell: `export PATH=$HOME/.local/go/bin:$PATH`.
- Code style: match the surrounding code; one-line comments only for the non-obvious why; `gofmt`. Lint must report `0 issues.`
- Docs/README text: plain English, fewest words.

## Review Focus

1. **Binary files, renames, mode-only changes and untracked files in hunk view**: no hunks to show. Expect a one-line message and no crash. Test in Task 1 (`TestParseBinaryAndModeOnly`) and Task 3 (`TestHunkModeUntrackedFile`).
2. **Partial line staging of a hunk whose unselected lines are deletions**: the patch must keep them as context, or `git apply` rejects it. Test in Task 1 (`TestPatchLinesForwardKeepsUnselectedDeletions`).
3. **Hook in a non-repo folder, a bare repo, or with global options (`git -C x ...`)**: `git` must run normally with no snapshot and no output from gitwhy. Test in Task 6 (`TestPreSkips`).
4. **Hook when gitwhy itself errors or isn't on PATH**: the user's git command must still run and keep its exit code. Test in Task 7 (`TestHookScriptKeepsExitCode`).
5. **Graph with merges and `--all`**: connector-only lines must not be selectable, and actions on commits not in HEAD's history must be refused (except copying). Test in Task 4 (`TestParseGraphSkipsConnectors`) and Task 5 (`TestGraphActionsOutsideHistory`).

---

## Phase A: TUI hunk & line staging

### Task 1: `internal/hunk` — parse a diff and build partial patches

**Files:**
- Create: `internal/hunk/hunk.go`
- Test: `internal/hunk/hunk_test.go`

**Interfaces:**
- Produces:
  - `type Hunk struct { Header string; Lines []string }`: `Header` is the full `@@ ... @@ context` line; `Lines` are the body lines, each starting with `' '`, `'+'`, `'-'` or `'\'` (for "\ No newline at end of file").
  - `type FileDiff struct { Header []string; Hunks []Hunk }`: `Header` is every line before the first `@@` (`diff --git`, `index`, `---`, `+++`, mode lines).
  - `func Parse(diff string) (*FileDiff, error)`: parses the output of `git diff -- <one path>`. Empty input gives `&FileDiff{}` with 0 hunks and no error. Input with a `Binary files` line, or with no `@@`, gives 0 hunks and no error.
  - `func (fd *FileDiff) Patch(h int, sel []bool, reverse bool) string`: a patch for hunk `h` only. `sel` has one entry per `Hunk.Lines` element, true means include that change line; `nil` means the whole hunk. The result is meant for `git apply --cached --recount` (reverse: add `-R`).
  - `func (h Hunk) Changes() (added, removed int)`

- [ ] **Step 1: Write the failing tests** in `internal/hunk/hunk_test.go`. They use `testrepo.New(t)` from `internal/testrepo`, so every patch is checked by real git.

```go
func TestParseTwoHunks(t *testing.T)              // file with 20 lines; edit line 2 and line 18 -> len(Hunks)==2, Header has "+++ b/f.txt"
func TestPatchWholeHunkStagesOnlyThatHunk(t *testing.T)
// Patch(0, nil, false) applied with `git apply --cached --recount` ->
// `git diff --cached` shows only the line-2 change; `git diff` still shows the line-18 change.
func TestPatchLinesForwardKeepsUnselectedDeletions(t *testing.T)
// a hunk that replaces line 2 ("-old","+new"); select only "+new" -> after apply,
// `git show :f.txt` contains both "old" and "new" (the deletion was kept as context).
func TestPatchReverseUnstagesHunk(t *testing.T)
// stage both hunks, Parse(`git diff --cached`), Patch(1, nil, true) applied with
// `git apply --cached -R --recount` -> only the line-2 change remains staged.
func TestParseBinaryAndModeOnly(t *testing.T)
// binary file diff and a chmod-only diff -> 0 hunks, nil error.
```

- [ ] **Step 2: Run** `go test ./internal/hunk/` and expect a compile failure (package doesn't exist).

- [ ] **Step 3: Implement** `Parse`, `Patch` and `Changes` in `internal/hunk/hunk.go`. `Patch` algorithm (must be exactly this):

```
out = fd.Header lines (only "diff --git", "index", "---", "+++", "old mode", "new mode", "new file mode", "deleted file mode")
out += hunk.Header
for i, line in hunk.Lines:
    keep := sel == nil || sel[i]
    switch line[0]:
    case ' ', '\\': emit line
    case '+': if keep { emit } else if reverse { emit " "+line[1:] }   // forward: drop
    case '-': if keep { emit } else if !reverse { emit " "+line[1:] }  // reverse: drop
end with "\n"
```

`--recount` makes git ignore the counts in the `@@` header, so the header is copied as-is.

- [ ] **Step 4: Run** `go test ./internal/hunk/` and expect PASS.

- [ ] **Step 5: Commit** with `git add internal/hunk && git commit -m "Add hunk package: parse diffs, build partial patches"`.

### Task 2: TUI hunk mode (stage/unstage whole hunks)

**Files:**
- Modify: `internal/tui/tui.go` (model struct, `key`, `tabKey`, `render`, `keys`)
- Create: `internal/tui/hunks.go` (hunk-mode state, loading, keys, rendering)
- Test: `internal/tui/hunks_test.go`

**Interfaces:**
- Consumes: `hunk.Parse`, `(*FileDiff).Patch`, `Hunk.Changes` (Task 1). Existing `model.repo *git.Repo`, `model.files []file`, `file.staged()/unstaged()/untracked()/conflict()`, `model.load()`, `model.msg`.
- Produces (unexported, same package):
  - New `model` fields: `hunks *hunkView` (nil when not in hunk mode).
  - `type hunkView struct { path string; staged bool; fd *hunk.FileDiff; cur int; lineMode bool; line int; sel []bool }`
  - `func (m *model) openHunks(path string, staged bool) error`: loads `git diff [--cached] -- path` into `fd`; error if it has 0 hunks.
  - `func (m *model) hunkKey(k string) (tea.Model, tea.Cmd)`: handles all keys while `m.hunks != nil`.
  - `func (m *model) applyHunk(sel []bool) error`: forward: `git apply --cached --recount` (stdin = patch). Unstage (`staged==true`): `git apply --cached -R --recount`. Then reload: if the file has no hunks left on this side, leave hunk mode.

Behaviour (Changes tab):
- `enter` on a non-conflict, tracked file opens hunk mode. Show the unstaged side if `f.unstaged()`, else the staged side. On an untracked file, set `m.msg = "New file: press space to stage it whole"` and stay in file mode.
- In hunk mode: `j/k` move between hunks, `space` stages (or unstages) the selected hunk, `S` switches staged/unstaged side (if that side has hunks), `esc` or `q` goes back to the file list (`q` must not quit the TUI here), `v` enters line mode (Task 3).
- Render: the left pane lists hunks as `@@ -a,b +c,d @@ ctx  +N −M`. The right pane shows the selected hunk's lines, colored like `detailView`. The title line shows `path — unstaged changes` or `path — staged changes`. Footer keys: `space stage hunk · S staged/unstaged · v pick lines · esc back` (with "unstage" when `staged`).

- [ ] **Step 1: Write failing tests** in `internal/tui/hunks_test.go`. Build the model directly the way `tui_test.go` does (`m := &model{repo: r.Repo}; m.load()`), then drive it with `m.key("enter")` etc.

```go
func TestHunkModeStagesOneHunk(t *testing.T)
// 20-line file committed, edit lines 2 and 18; enter, space ->
// `git diff --cached --stat` shows f.txt; `git diff -U0` still has the line-18 change; m.hunks.fd has 1 hunk left.
func TestHunkModeUnstageSide(t *testing.T)
// stage whole file, enter (opens staged side since no unstaged changes), space -> nothing staged; m.hunks == nil (back to files).
func TestHunkModeUntrackedFile(t *testing.T)
// untracked file selected, enter -> m.hunks == nil and m.msg contains "New file".
func TestHunkModeQGoesBack(t *testing.T)
// in hunk mode, key("q") returns m.action.Kind != "quit" and m.hunks == nil.
```

- [ ] **Step 2: Run** `go test ./internal/tui/ -run HunkMode` and expect FAIL.
- [ ] **Step 3: Implement** in `internal/tui/hunks.go` and wire it in `tui.go`. At the top of `key`, add `if m.hunks != nil { return m.hunkKey(k) }`. In `render`, use the hunk panes when `m.hunks != nil`. Run `git apply` with `m.repo.OutputIn(strings.NewReader(patch), "apply", "--cached", "--recount", ...)`.
- [ ] **Step 4: Run** `go test ./internal/tui/` and expect PASS.
- [ ] **Step 5: Commit** with `git commit -m "TUI: stage and unstage single hunks"`.

### Task 3: Line mode, docs, release prep

**Files:**
- Modify: `internal/tui/hunks.go`, `internal/tui/hunks_test.go`, `README.md` (home screen section)

**Interfaces:**
- Consumes: `hunkView` (Task 2), `(*FileDiff).Patch(h, sel, reverse)` (Task 1).

Behaviour: `v` turns on line mode for the selected hunk, with `sel` all false and `line` on the first `+`/`-` line. `j/k` move between change lines only (skip context). `space` toggles `sel[line]`. `enter` applies the selected lines via `applyHunk(sel)`; with none selected, `m.msg = "Pick lines with space first"`. `esc` leaves line mode. Render selected lines with a `▌` marker and the cursor line reversed.

- [ ] **Step 1: Write the failing test**

```go
func TestLineModeStagesPickedLine(t *testing.T)
// hunk adds "+a" and "+b" after line 5; enter, v, space (selects first change line "+a"), enter ->
// `git show :f.txt` contains "a" and not "b"; `git diff` shows only "+b".
```

- [ ] **Step 2: Run** `go test ./internal/tui/ -run LineMode` and expect FAIL.
- [ ] **Step 3: Implement** line mode.
- [ ] **Step 4: Run** `go test ./internal/tui/` and expect PASS.
- [ ] **Step 5: README:** in "The home screen" add one sentence: "`enter` on a changed file shows its hunks: `space` stages one, `v` picks single lines."
- [ ] **Step 6: Phase check:** run `go test ./... && golangci-lint run ./...` and expect all `ok` and `0 issues.`
- [ ] **Step 7: Manual check with vhs** (one run, keep it short). Write `/tmp/hunk.tape`: make a temp repo with a 20-line file and 2 edits, run `gitwhy`, `Enter`, `Space`, `Screenshot /tmp/hunk.png`, `q`, `q`. Paths in vhs `Output`/`Screenshot` must be relative to the tape's folder. Look at the PNG and confirm one hunk is staged.
- [ ] **Step 8: Commit** with `git commit -am "TUI: stage single lines; docs"`. Ask the user before tagging `v0.4.0` and pushing.

---

## Phase B: TUI commit graph

### Task 4: `internal/graph` — parse `git log --graph`

**Files:**
- Create: `internal/graph/graph.go`, `internal/graph/graph_test.go`

**Interfaces:**
- Produces:
  - `type Row struct { Graph string; Commit *Commit }`: `Commit` is nil for connector-only lines.
  - `type Commit struct { ID, Short, Subject, When, Author, Refs string }`
  - `func Load(r *git.Repo, all bool, n int) ([]Row, error)`: runs `git log --graph --color=never -n <n> --format=%x00%H%x00%h%x00%s%x00%cr%x00%an%x00%D [--all] --`. Each output line splits at the first `\x00`. The part before it is `Graph` (trailing spaces trimmed). If a field part exists, it is a commit row; otherwise the whole line is `Graph` with a nil `Commit`.
  - `func Parse(out string) []Row`: the pure part of `Load`, for tests.

- [ ] **Step 1: Failing tests**

```go
func TestParseGraphSkipsConnectors(t *testing.T)
// literal input with "* \x00id1\x00a\x00s1\x00x\x00y\x00HEAD -> main", "|\\ ", "| * \x00id2...", "|/ ", "* \x00id3..." ->
// 5 rows; rows[1].Commit == nil; rows[2].Graph == "| *"; rows[0].Commit.Refs == "HEAD -> main".
func TestLoadWithMerge(t *testing.T)
// testrepo with a --no-ff merge -> Load(r, false, 50) has 3+ commit rows and at least one connector row.
```

- [ ] **Step 2: Run** `go test ./internal/graph/` and expect FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run** and expect PASS.
- [ ] **Step 5: Commit** with `git commit -m "Add graph package: parse git log --graph"`.

### Task 5: Show the graph in the Commits tab

**Files:**
- Modify: `internal/tui/tui.go` (replace the `commits []commit` loading in `load`; `list`; `tabKey` commits branch; `count`; `keys[tabCommits]`)
- Test: `internal/tui/tui_test.go` (update `TestCommitKeysPresetCommit`), new `internal/tui/graph_test.go`

**Interfaces:**
- Consumes: `graph.Load`, `graph.Row`, `graph.Commit` (Task 4).
- Produces: `model.rows []graph.Row`, `model.all bool` (`a` toggles all branches), and `func (m *model) commitAt(i int) *graph.Commit`, which returns the i-th row whose `Commit != nil`. The commit cursor counts commit rows only, and `count(tabCommits)` is the number of commit rows. The old `commit` struct and `m.commits` go away; update every use.

Behaviour:
- Draw each row as `Graph` (each char colored by its column index, cycling colors 2–6), then `short subject (refs)`. Connector rows are drawn but never selected; the cursor skips them. Scrolling must keep the cursor visible, counting all rows.
- `a` toggles `--all` and reloads.
- If a commit is not in HEAD's history (`git merge-base --is-ancestor <id> HEAD` fails), keys `r d f s x m o` set `m.msg = "That commit isn't on your branch. Switch to its branch first, or p to copy it here."` and do nothing.
- New key `p` (any commit): `run("copy-commit", recipe.Answers{"from": id, "commit": id})`.
- Footer keys for the tab: add `p copy here · a all branches`.

- [ ] **Step 1: Failing tests** in `internal/tui/graph_test.go`:

```go
func TestGraphActionsOutsideHistory(t *testing.T)
// branch "other" with a commit not on main; m.all=true; load; move the cursor to that commit;
// tabKey("r") -> ok==false and m.msg contains "isn't on your branch"; tabKey("p") -> Recipe=="copy-commit".
func TestGraphCursorSkipsConnectors(t *testing.T)
// repo with a merge; count(tabCommits) == number of commit rows; commitAt(k) != nil for every k < count.
```

- [ ] **Step 2: Run** `go test ./internal/tui/ -run Graph` and expect FAIL.
- [ ] **Step 3: Implement**, and update `TestCommitKeysPresetCommit` to use `commitAt(0).ID`.
- [ ] **Step 4: Run** `go test ./internal/tui/` and expect PASS.
- [ ] **Step 5: Phase check:** `go test ./... && golangci-lint run ./...`. Do one vhs screenshot of a repo with a merge, and confirm the graph lines up with the commits.
- [ ] **Step 6: Commit** and update the README (one line: "Commits tab draws the branch graph; `a` shows all branches, `p` copies a commit here."). Ask before tagging `v0.5.0`.

---

## Phase C: Shell hook (undo points for plain `git`)

Design: `gitwhy hook init bash` prints a shell function that wraps `git`. Before each git command it calls `gitwhy __pre <args>`. That call takes a snapshot only for risky commands and prints the snapshot ID. After git exits, it calls `gitwhy __post <id>` (`snapshot.Finish`). `gitwhy undo` / `undo --list` then work for plain git commands. The hook works in bash and zsh; fish is not supported (say so in the help).

### Task 6: `__pre` / `__post` commands

**Files:**
- Create: `internal/hook/hook.go`, `internal/hook/hook_test.go`
- Modify: `internal/snapshot/snapshot.go` (add `Get`), `cmd/gitwhy/main.go` (route `__pre`, `__post` before `checkGit` and before loading docs, so they're fast)

**Interfaces:**
- Consumes: `danger.Warnings(args []string) []string`, `preview.Preview(repo, args).Save`, `snapshot.Take(r, label, extra)`, `snapshot.Finish(r, s)`, `git.Open(dir)`.
- Produces:
  - `func snapshot.Get(r *git.Repo, id string) (*Snapshot, error)`: reads `<gitdir>/gitwhy/snapshots/<id>/meta.json`.
  - `func hook.Pre(cwd string, args []string) string`: returns the snapshot ID, or `""` when no snapshot was taken. It never returns an error and never prints.
  - `func hook.Post(cwd, id string)`: calls `snapshot.Finish`; it ignores errors.
  - `func hook.Risky(args []string) bool`: true when `len(danger.Warnings(args)) > 0`, but false when every warning is about the remote (warning text contains "gitwhy can't undo this", or `args[0] == "push"`).

`Pre` rules:
- Return `""` for an empty `args`, or when `args[0]` starts with `-` (global options like `-C`, `-c`, `--git-dir`).
- Return `""` when `GITWHY_HOOK_OFF=1`.
- Return `""` when `git.Open(cwd)` fails (no repo, or a bare repo).
- Return `""` when `!Risky(args)`.
- Otherwise take a snapshot with label `"git " + strings.Join(args, " ")` and `extra = preview.Preview(repo, args).Save`.

- [ ] **Step 1: Failing tests** in `internal/hook/hook_test.go`:

```go
func TestPreSkips(t *testing.T)
// Pre(nonRepoDir, ["reset","--hard"])=="" ; Pre(repo, ["-C","x","reset","--hard"])=="" ;
// Pre(repo, ["status"])=="" ; Pre(repo, ["push","--force"])=="" ; with t.Setenv("GITWHY_HOOK_OFF","1") -> "".
func TestPreThenUndo(t *testing.T)
// repo with a dirty file; id := Pre(dir, ["reset","--hard"]); run `git reset --hard`; Post(dir, id);
// s, _ := snapshot.Get(r, id); snapshot.Restore(r, s) -> the dirty content is back.
```

- [ ] **Step 2: Run** `go test ./internal/hook/` and expect FAIL.
- [ ] **Step 3: Implement** `snapshot.Get`, `hook.Pre`, `hook.Post` and `hook.Risky`. In `main.go`, before the flag loop: `if len(args) >= 1 && args[0] == "__pre" { wd, _ := os.Getwd(); fmt.Print(hook.Pre(wd, args[1:])); return nil }`; likewise `__post <id>`.
- [ ] **Step 4: Run** `go test ./internal/hook/ ./internal/snapshot/ -run 'Pre|Get'` and expect PASS.
- [ ] **Step 5: Measure** the cost on a non-risky command: `time ~/.local/bin/gitwhy __pre status` in any repo must be under 50 ms. If it isn't, stop importing the docs/recipes path for `__pre` (check that `main.go` returns before `ui.New`).
- [ ] **Step 6: Commit** with `git commit -m "Add hook pre/post: snapshots for risky plain git commands"`.

### Task 7: `gitwhy hook init|install|uninstall`

**Files:**
- Modify: `internal/hook/hook.go` (script text and rc-file editing), `cmd/gitwhy/main.go` (route `hook`), `internal/hook/hook_test.go`

**Interfaces:**
- Produces:
  - `func hook.Script(shell string) (string, error)`: `shell` is `"bash"` or `"zsh"`; anything else returns an error saying `only bash and zsh are supported`.
  - `func hook.Install(rcPath, shell string) (changed bool, err error)`: appends the block between `# >>> gitwhy hook >>>` and `# <<< gitwhy hook <<<`, containing `eval "$(gitwhy hook init <shell>)"`. Idempotent.
  - `func hook.Uninstall(rcPath string) (changed bool, err error)`: removes that block only.

The script must be exactly this function (identical in bash and zsh):

```sh
git() {
  local __gw_id __gw_rc
  __gw_id=$(command gitwhy __pre "$@" 2>/dev/null)
  command git "$@"
  __gw_rc=$?
  if [ -n "$__gw_id" ]; then command gitwhy __post "$__gw_id" >/dev/null 2>&1; fi
  return $__gw_rc
}
```

- [ ] **Step 1: Failing tests**

```go
func TestHookScriptKeepsExitCode(t *testing.T)
// write Script("bash") + "\ngit rev-parse --verify nope; echo rc=$?" to a file; run `bash <file>` with PATH lacking gitwhy
// -> output contains "rc=128" (git's own exit code, even though gitwhy isn't found).
func TestInstallIdempotentAndUninstall(t *testing.T)
// temp rc with "export A=1\n"; Install twice -> marker appears once; Uninstall -> file == "export A=1\n".
```

- [ ] **Step 2: Run** and expect FAIL.
- [ ] **Step 3: Implement.** CLI:
  - `gitwhy hook init <shell>` prints the script.
  - `gitwhy hook install`: shell from `$SHELL` (`bash` uses `~/.bashrc`, `zsh` uses `~/.zshrc`). It asks with `confirm("Add the gitwhy hook to <rc>?", true)`, then prints "Open a new terminal (or run: source <rc>)".
  - `gitwhy hook uninstall` removes the block.
  - Add these three lines to `usage` in `main.go`.
- [ ] **Step 4: Run** `go test ./internal/hook/` and expect PASS.
- [ ] **Step 5: Commit** with `git commit -m "Add gitwhy hook init/install/uninstall"`.

### Task 8: End-to-end check and docs

**Files:**
- Modify: `cmd/gitwhy/e2e_test.go`, `README.md`

- [ ] **Step 1: Failing e2e test** using the existing `bin`, `newRepo`, `drive`, `yes`, `expect` helpers:

```go
func TestHookUndoesPlainGit(t *testing.T)
// script := output of exec.Command(bin, "hook", "init", "bash");
// run bash -c "<script>; cd <repo>; echo dirty > f1.txt; git reset --hard" with PATH = dir(bin):$PATH and env(...)
// then drive(t, r.dir, []string{"undo"}, yes) -> r.read("f1.txt") == "dirty\n".
```

- [ ] **Step 2: Run** `go test ./cmd/gitwhy/ -run Hook` and expect FAIL until main routes `hook` (Task 7). Then expect PASS.
- [ ] **Step 3: README:** add a section "Undo for plain git (optional)" with `gitwhy hook install`. Say what it does in two sentences: it snapshots before risky commands only, costs about 30 ms otherwise, and works in bash/zsh. Give `GITWHY_HOOK_OFF=1` to pause it.
- [ ] **Step 4: Phase check:** `go test ./... && golangci-lint run ./...`.
- [ ] **Step 5: Commit.** Ask before tagging `v0.6.0`.

---

## Phase D: Windows / Git Bash

There's no Windows machine locally; GitHub Actions `windows-latest` (it has Git for Windows and Git Bash) is the test bed.

### Task 9: Windows CI and builds

**Files:**
- Modify: `.github/workflows/ci.yml`, `.goreleaser.yaml`

- [ ] **Step 1:** In `ci.yml`, make the job a matrix: `runs-on: ${{ matrix.os }}` with `os: [ubuntu-latest, windows-latest]`. Keep the apt step only for `if: runner.os == 'Linux'`. On Windows, `git-lfs` ships with Git for Windows. Install filter-repo with `pip install git-filter-repo`, or let those tests skip (they already skip when the tool is missing). Run golangci-lint only on Linux.
- [ ] **Step 2:** In `.goreleaser.yaml`: `goos: [linux, darwin, windows]`; add `format_overrides: [{goos: windows, formats: [zip]}]` under archives. Run `goreleaser check` and expect `1 configuration file(s) validated`.
- [ ] **Step 3:** Commit, then ask the user to push (CI only runs on GitHub). After the push, read failures with `gh run view <id> --log-failed | tail -80`.
- [ ] **Step 4: Commit** with `git commit -m "CI: test on Windows; release Windows zip"`.

### Task 10: Fix Windows failures

Expected trouble spots, in order of likelihood. Fix only what CI shows.
- **`/dev/null` in tests:** `internal/testrepo` sets `core.hooksPath=/dev/null`, and `internal/tui` diffs untracked files against `/dev/null`. Git for Windows accepts `/dev/null` in `git diff --no-index`. For `core.hooksPath`, use an empty `t.TempDir()` instead.
- **Path separators:** git prints `/` paths; `filepath.Join(r.Top, filepath.FromSlash(p))` is already used in `snapshot`. Search for other `r.Top + "/"` style joins: `grep -rn 'Top + "/' internal`.
- **`todo.Editor`:** git runs `GIT_SEQUENCE_EDITOR` through its bundled `sh`, so single-quote quoting works. A Windows path like `C:\Users\...` must become `C:/Users/...` for sh. Apply `filepath.ToSlash` to `self` and `plan` in `todo.Editor`.
- **e2e tests that use `sh -c` or `test ! -e`:** run them on Windows only if `sh` is on PATH (`exec.LookPath("sh")`), otherwise `t.Skip`.
- **Line endings:** CI may check out with `core.autocrlf=true`. Docs parse fine with `\r\n` only if `kb.Parse` trims `\r`. If kb tests fail, add `strings.TrimRight(line, "\r")` in `kb.Parse`.

- [ ] **Step 1:** For each failure: write or keep the failing test, fix, run that test locally on Linux (must still pass), and commit with `git commit -m "windows: <what>"`.
- [ ] **Step 2:** Repeat the push/CI cycle (ask before each push) until CI is green on both OSes.
- [ ] **Step 3: README:** change platform lines to "Linux, macOS (untested), Windows (Git Bash or PowerShell; CI-tested)". Add a PowerShell install line only if `install.ps1` exists; don't write one in this plan.
- [ ] **Step 4:** Ask before tagging `v0.7.0`.

---

## Phase E: Homebrew tap

### Task 11: Publish a tap via goreleaser

Needs the user. Ask them to:
1. Create the public repo `dpologdvinity/homebrew-tap`, or allow you to run `gh repo create dpologdvinity/homebrew-tap --public --description "Homebrew formulae"`.
2. Create a fine-grained token with Contents read/write on that repo only, then run `gh secret set HOMEBREW_TAP_TOKEN -R dpologdvinity/gitwhy` themselves. Never handle the token value yourself.

**Files:**
- Modify: `.goreleaser.yaml`, `.github/workflows/release.yml`, `README.md`

- [ ] **Step 1:** Check which section goreleaser 2.19 wants (`brews` is deprecated in favor of `homebrew_casks` in v2.10+): run `goreleaser check` after adding it. Use `homebrew_casks` with `repository: {owner: dpologdvinity, name: homebrew-tap, token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"}`, `homepage: https://github.com/dpologdvinity/gitwhy`, `description: "Do anything with git without looking it up, with previews and undo"`, `license: MIT`. Linux casks need `binaries: [gitwhy]`; follow `goreleaser check` warnings until it says validated.
- [ ] **Step 2:** In `release.yml`, add `HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}` to the goreleaser step env.
- [ ] **Step 3: README** install section: add `brew install dpologdvinity/tap/gitwhy`.
- [ ] **Step 4:** Commit. On the next tag the user approves, confirm with `gh api repos/dpologdvinity/homebrew-tap/contents` that a file for gitwhy appeared, and run `brew install dpologdvinity/tap/gitwhy && gitwhy version`.

---

## After every phase

- Update the agent's project memory "Status" line with the new version and what shipped.
- Rebuild the local binary: `go build -o ~/.local/bin/gitwhy ./cmd/gitwhy`.
