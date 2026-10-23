#!/usr/bin/env python3
"""Build gitwhy/docs/README.md: every command grouped by git's own categories, with its one-line summary."""
import os, subprocess

DOCS = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "docs")
CMD_DIR = os.path.join(DOCS, "commands")

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

placed = {c for cs in groups.values() for c in cs}
other = sorted(f[:-3] for f in os.listdir(CMD_DIR) if f.endswith(".md") and f[:-3] not in placed and f != "git.md")
if other:
    groups["Low-level Commands / Internal Helpers"] += other

def summary(cmd):
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
