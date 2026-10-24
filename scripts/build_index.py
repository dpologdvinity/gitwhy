#!/usr/bin/env python3
"""Build docs/README.md: every command grouped by git's own categories, with its one-line summary.

Usage: python3 scripts/build_index.py  (no arguments; run it from anywhere)

Needs Python 3 and an installed git: the groups come from `git help -a`, so a different git
version can move or add commands. Each summary is the first sentence of docs/commands/<cmd>.md.
Documented commands that git does not list are added to "Internal helpers". Prints the commands
that have no doc file. Reads no environment variables.
"""
import os, subprocess

# Paths are found from this file's location, so the current folder does not matter.
DOCS = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "docs")
CMD_DIR = os.path.join(DOCS, "commands")

# git's category headings from `git help -a`, mapped to the plainer headings we show. The order
# here is the order of the sections in the index.
NAMES = {
    "Main Porcelain Commands": "Everyday commands",
    "Ancillary Commands / Manipulators": "Change settings and repo data",
    "Ancillary Commands / Interrogators": "Inspect the repo",
    "Interacting with Others": "Work with other systems (SVN, Perforce, email, CVS)",
    "Low-level Commands / Manipulators": "Low-level: change objects and refs",
    "Low-level Commands / Interrogators": "Low-level: read objects and refs",
    "Low-level Commands / Syncing Repositories": "Low-level: transfer data between repos",
    "Low-level Commands / Internal Helpers": "Internal helpers (git runs these for you)",
    "External commands": "Separate tools",
    "Command aliases": "Old names and aliases",
}
# not installed locally; categories from git's command-list.txt
EXTRA = {
    "Main Porcelain Commands": ["citool", "gitk", "gui", "scalar"],
    "Ancillary Commands / Interrogators": ["gitweb"],
    "Interacting with Others": ["archimport", "cvsexportcommit", "cvsimport", "cvsserver", "p4", "send-email", "svn"],
    "Low-level Commands / Internal Helpers": ["sh-i18n", "sh-setup"],
    "Command aliases": ["fsck-objects", "init-db", "stage", "pickaxe"],
}
SKIP = {"remote-codecommit"}  # third-party remote helper, not part of git

# Read `git help -a`: a heading line starts at column 0, and the command lines under it are
# indented, with the command name first.
groups, cur = {}, None
for line in subprocess.run(["git", "help", "-a"], capture_output=True, text=True).stdout.splitlines():
    if line and not line.startswith(" "):
        # "Command aliases" lists the user's own aliases, not git's; skip it
        cur = line if line in NAMES and line != "Command aliases" else None
        if cur: groups[cur] = []
    elif cur and line.startswith("   ") and line.split():
        if line.split()[0] not in SKIP:
            groups[cur].append(line.split()[0])
for g, cs in EXTRA.items():
    groups.setdefault(g, []).extend(cs)

# Documented commands that no group lists still get a row. git.md is the global options page,
# linked from the header instead.
placed = {c for cs in groups.values() for c in cs}
other = sorted(f[:-3] for f in os.listdir(CMD_DIR) if f.endswith(".md") and f[:-3] not in placed and f != "git.md")
if other:
    groups["Low-level Commands / Internal Helpers"] += other

def summary(cmd):
    """Return the first sentence of docs/commands/<cmd>.md, ending in one full stop.

    The first non-blank line that is not a heading is used. Relative links are rewritten so they
    work from docs/README.md. Returns None when the doc file does not exist, and "" when it has
    no text line.
    """
    p = os.path.join(CMD_DIR, cmd + ".md")
    if not os.path.exists(p):
        return None
    for line in open(p):
        line = line.strip()
        if line and not line.startswith("#"):
            s = line.split(". ")[0].rstrip(".") + "."
            # summaries come from commands/, so re-root their relative links at docs/
            return s.replace("](../shared/", "](shared/").replace("](", "](commands/").replace("](commands/shared/", "](shared/").replace("](commands/http", "](http")
    return ""

# Write one markdown table per group, commands sorted and without duplicates. Commands that git
# lists but we have not documented are left out and reported at the end.
out = ["# gitwhy: every git command, in plain English", "",
       "One file per command, with every flag. Global options (`-C`, `-c`, `--no-pager`, ...): [`git`](commands/git.md).",
       "Shared option sets: [diff options](shared/diff-options.md), [revision options](shared/revision-options.md).", ""]
missing = []
for g in NAMES:
    cs = groups.get(g)
    if not cs:
        continue
    out += [f"## {NAMES[g]}", "", "| Command | What it does |", "|---------|--------------|"]
    for c in sorted(set(cs)):
        s = summary(c)
        if s is None:
            missing.append(c); continue
        out.append(f"| [`{c}`](commands/{c}.md) | {s} |")
    out.append("")
open(os.path.join(DOCS, "README.md"), "w").write("\n".join(out))
print("index written; commands without a doc file:", " ".join(missing) or "none")
