#!/usr/bin/env python3
"""Report flags in the installed git's -h/man output that are missing from gitwhy docs.

Usage: check_coverage.py [command...]  (default: all commands)

Checks every main git command, plus git-filter-repo, git-lfs and the global options ("_git",
documented in docs/commands/git.md). Needs Python 3, an installed git, and `man` with git's man
pages; without man only the `-h` output is checked. Results depend on the installed git version.
Sets MANWIDTH=110 for man so option lines wrap the same way every time.

Exits 1 when a command has no doc file at all. Missing flags are only printed, not an error,
because a newer git can add flags before the docs catch up.
"""
import os, re, sys, subprocess

# Paths are found from this file's location, so the current folder does not matter.
DOCS = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "docs")
cmds = subprocess.run(["git", "--list-cmds=main,nohelpers"], capture_output=True, text=True).stdout.split()
cmds += ["filter-repo", "lfs", "_git"]

# A flag such as -n or --dry-run, not part of a longer word such as "re-run".
FLAG = re.compile(r"(?<![\w-])(--?[A-Za-z0-9][\w-]*)")
# flag lines: -h output lines, and man option headers (indented 7 spaces, start with -)
HDR = re.compile(r"^(?:\s{4}|\s{7})(-[^\s].*)$")

def flags_in_source(text):
    """Return the set of flags named on the option lines of git's help text.

    Only the part of a line before a double space is read, so flags mentioned in the description
    are not counted. --no-x is stored as --x, so either form in the docs counts.
    """
    found = set()
    for line in text.splitlines():
        m = HDR.match(line)
        if m:
            for f in FLAG.findall(m.group(1).split("  ")[0]):
                found.add(f.replace("--no-", "--", 1) if f.startswith("--no-") else f)
    return found

def source_text(cmd):
    """Return `git <cmd> -h` plus the man page, from the installed git.

    "_git" gives `git -h` and the git(1) man page. "lfs" gives every git-lfs* man page, since
    git-lfs has one page per subcommand. stdin is closed so no command can wait for input.
    """
    run = lambda args, **kw: subprocess.run(args, capture_output=True, text=True, stdin=subprocess.DEVNULL, **kw)
    if cmd == "_git":
        return run(["git", "-h"]).stdout + run(["man", "-P", "cat", "git"], env={**os.environ, "MANWIDTH": "110"}).stdout
    if cmd == "lfs":
        mandir = os.path.dirname(run(["man", "-w", "git-lfs"]).stdout.strip())
        pages = sorted({f.split(".")[0] for f in os.listdir(mandir) if f.startswith("git-lfs")}) if mandir else []
        return "".join(run(["man", "-P", "cat", p], env={**os.environ, "MANWIDTH": "110"}).stdout for p in pages)
    # Depending on the git version and command, -h usage goes to stdout or stderr, so keep both.
    h = run(["git", cmd, "-h"])
    m = run(["man", "-P", "cat", "git-" + cmd], env={**os.environ, "MANWIDTH": "110"})
    return h.stdout + h.stderr + m.stdout

def doc_text(name):
    """Return the text of a command's doc file plus every shared option page it links to.

    Returns None when docs/commands/<name>.md does not exist.
    """
    p = os.path.join(DOCS, "commands", ("git" if name == "_git" else name) + ".md")
    if not os.path.exists(p):
        return None
    text = open(p).read()
    for shared in re.findall(r"\((?:\.\./)?shared/([\w-]+\.md)\)", text):
        sp = os.path.join(DOCS, "shared", shared)
        if os.path.exists(sp):
            text += open(sp).read()
    return text

def normalized(text):
    """Return every flag in the docs text, with --no-x stored as --x like flags_in_source."""
    # accept --[no-]x, --no-x, and --x forms
    return set(f.replace("--no-", "--", 1) for f in FLAG.findall(text.replace("[no-]", "")))

# Compare each command's flags with its docs. A bare "-" or "--" is not a flag to document.
missing_files, total, gaps = [], 0, {}
only = set(sys.argv[1:])
for c in cmds:
    if only and c not in only:
        continue
    doc = doc_text(c)
    if doc is None:
        missing_files.append(c); continue
    want = flags_in_source(source_text(c))
    have = normalized(doc)
    miss = sorted(f for f in want if f not in have and f not in ("-", "--"))
    total += len(want)
    if miss:
        gaps[c] = miss

print(f"commands without doc: {len(missing_files)} {' '.join(missing_files)}")
print(f"source flags checked: {total}; commands with gaps: {len(gaps)}")
for c, m in gaps.items():
    print(f"  {c}: {' '.join(m)}")
sys.exit(1 if missing_files else 0)
