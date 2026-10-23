# git diagnose

Zip up machine, git, and repo details for debugging. Unlike [`git bugreport`](bugreport.md), focuses on repo size and layout.

Default contents: `git version --build-options`, repo root path, free disk space, every pack's name and size (including alternates), loose object counts.

**Usage:** `git diagnose [<flags>]`

## Flags

| Flag | What it does |
|------|--------------|
| `-o <path>`, `--[no-]output-directory <path>` | Write the zip to `<path>`. Default: current folder. |
| `-s <format>`, `--[no-]suffix <format>` | Name it `git-diagnostics-<suffix>`, suffix = current time in this `strftime` format. |
| `--mode=(stats\|all)` | `stats` (default): the above. `all`: also copies `.git`, `.git/hooks`, `.git/info`, `.git/logs`, `.git/objects/info`; may be enough to rebuild your repo, so share carefully. |
