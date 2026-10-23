# git mktag

Low-level: make an annotated tag object from stdin and print its ID. Unlike `git hash-object -t tag -w --stdin`, refuses tags failing strict `git fsck` checks (warnings count as errors; extra headers rejected). Creates no ref; use `git update-ref refs/tags/<name> <id>`. For normal tagging use `git tag -a`.

Input: four header lines, then optionally a blank line and a message (may include a GPG signature):

```
object <hash>
type <typename>
tag <tagname>
tagger <name> <<email>> <timestamp> <timezone>
```

Allow extra headers with `git -c fsck.extraHeaderEntry=ignore mktag`.

**Usage:** `git mktag <tag-file`

## Flags

| Flag | What it does |
|------|--------------|
| `--[no-]strict` | Use `git fsck --strict` checks. On by default. |
