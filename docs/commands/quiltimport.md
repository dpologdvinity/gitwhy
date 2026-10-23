# git quiltimport

Apply a quilt patch series as commits on the current branch, one per patch, in order. Author comes from the patch, else `--author`, else you're asked. No subject: file name is used.

**Usage:** `git quiltimport [-n] [--author <author>] [--patches <dir>] [--series <file>] [--keep-non-patch]`

## Flags

| Flag | What it does |
|------|--------------|
| `-n`, `--[no-]dry-run` | Don't commit; check patches and warn about missing authors. |
| `--[no-]author "<name> <<email>>"` | Author for patches that lack one. |
| `--[no-]patches <dir>` | Patch folder. Default: `$QUILT_PATCHES` or `patches`. |
| `--[no-]series <file>` | Series file. Default: `$QUILT_SERIES` or `<patches>/series`. |
| `--[no-]keep-non-patch` | Strip only `[PATCH]` from subjects, keep other brackets (`git mailinfo -b`). |
