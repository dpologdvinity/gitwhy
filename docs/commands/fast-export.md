# git fast-export

Dump history (commits, files, tags, branches) as a text stream for [`git fast-import`](fast-import.md). For copying or script-editing history, or a readable `git bundle` alternative. `git fast-export --all | (cd /empty/repo && git fast-import)` makes a near-exact copy.

Can't export tags of trees.

**Usage:** `git fast-export [<flags>] [<revision-range>] [[--] <path>...]`

## Flags

Also takes the [commit selection options](../shared/revision-options.md), e.g. `--all`, `main~10..main`, `-- <path>`.

| Flag | What it does |
|------|--------------|
| `<revision-range>` | What to export. Files shared with the excluded boundary are still included (unless `--reference-excluded-parents`). |
| `[--] <path>...` | Only changes to these paths. |
| `--[no-]progress=<n>` | Emit a `progress` line every `<n>` objects. |
| `--[no-]signed-tags=<mode>` | Signed tags: `abort` (default), `strip`, `warn-strip`, `verbatim`, `warn-verbatim` (old name `warn`). Keep only if nothing will change. |
| `--[no-]signed-commits=<mode>` | Same for signed commits. Default `strip`. Written as `gpgsig <hash-algo> <format>`. Experimental. |
| `--[no-]tag-of-filtered-object=<mode>` | Tags of filtered-out objects: `abort` (default), `drop`, `rewrite` (point at nearest kept ancestor). |
| `-M`, `-C` | Emit renames (`-M`) and copies (`-C`). |
| `--[no-]reencode=<mode>` | Non-UTF-8 commits: `abort` (default), `yes` (convert to UTF-8), `no` (keep). |
| `--[no-]export-marks=<file>` | Write commit marks (`:<num> <id>`) to `<file>` at the end for incremental runs. Skipped if nothing new. Can be the `--import-marks` file. |
| `--[no-]import-marks=<file>` | Load marks first; already-marked objects aren't re-exported. |
| `--[no-]import-marks-if-exists=<file>` | Same, skipping a missing file. |
| `--[no-]mark-tags` | Mark tags too (needed for tags of tags, incremental tags). Off: many importers reject it. |
| `--[no-]fake-missing-tagger` | Invent a tagger for tagless old tags. |
| `--[no-]use-done-feature` | Wrap the stream in `feature done` ... `done`. |
| `--no-data` | Omit file contents; reference existing IDs. Target must have them. |
| `--data` | Include contents (default). |
| `--[no-]full-tree` | Write `deleteall` plus every file per commit, not just changes. |
| `--[no-]anonymize` | Replace names, paths, contents, messages, emails with deterministic fakes, keeping shape. For sharing bug repros. |
| `--anonymize-map=<from>[:<to>]` | With `--anonymize`: map token `<from>` to `<to>` (or keep it). Paths and refs split at `/`. |
| `--[no-]reference-excluded-parents` | Point the first commit of a range at its real parent ID instead of making a root. Target must have it. |
| `--[no-]show-original-ids` | Add `original-oid <id>` lines for filter scripts. |
| `--[no-]refspec=<refspec>` | Rename refs, e.g. `refs/heads/*:refs/heads/old/*`. Repeatable. |
