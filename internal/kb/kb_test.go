package kb

import (
	"os"
	"testing"
)

func load(t *testing.T) *KB {
	t.Helper()
	k, err := Load(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func TestLoadRealDocs(t *testing.T) {
	k := load(t)
	if n := len(k.Commands); n < 180 {
		t.Fatalf("got %d commands, want >= 180", n)
	}
	for _, name := range k.Names() {
		c := k.Commands[name]
		if c.Summary == "" {
			t.Errorf("%s: empty summary", name)
		}
	}
	if len(k.Shared["diff-options"]) < 50 || len(k.Shared["revision-options"]) < 50 {
		t.Errorf("shared pages too small: %d %d", len(k.Shared["diff-options"]), len(k.Shared["revision-options"]))
	}
}

func TestLookup(t *testing.T) {
	k := load(t)
	cases := []struct{ cmd, sub, arg string }{
		{"reset", "", "--soft"},
		{"reset", "", "--hard"},
		{"commit", "", "--amend"},
		{"commit", "", "--no-verify"},
		{"log", "", "--oneline"}, // shared revision options
		{"diff", "", "--stat"},   // shared diff options
		{"stash", "push", "-u"},
		{"push", "", "--force-with-lease=main"},
		{"lfs", "track", "--lockable"},
	}
	for _, c := range cases {
		if _, ok := k.Lookup(c.cmd, c.sub, c.arg); !ok {
			t.Errorf("Lookup(%s %s %s) not found", c.cmd, c.sub, c.arg)
		}
	}
	if s := k.Commands["stash"].Sub("pop"); s == nil || s.Summary == "" {
		t.Errorf("stash pop subcommand missing or no summary")
	}
}

func TestParseFlagForms(t *testing.T) {
	f, ok := parseFlag("`--[no-]verify`", "x")
	if !ok || len(f.Spellings) != 2 || f.Spellings[0] != "--verify" || f.Spellings[1] != "--no-verify" {
		t.Errorf("negatable: %+v", f)
	}
	f, _ = parseFlag("`-m <msg>`, `--message=<msg>`", "x")
	if f.Spellings[0] != "-m" || f.Spellings[1] != "--message" || f.Value != "<msg>" || f.Optional {
		t.Errorf("value: %+v", f)
	}
	f, _ = parseFlag("`-S[<key-id>]`, `--[no-]gpg-sign[=<key-id>]`", "x")
	if f.Spellings[0] != "-S" || f.Spellings[1] != "--gpg-sign" || f.Value != "<key-id>" || !f.Optional {
		t.Errorf("short optional: %+v", f)
	}
	f, _ = parseFlag("`--color[=<when>]`", "x")
	if f.Spellings[0] != "--color" || f.Value != "<when>" || !f.Optional {
		t.Errorf("optional: %+v", f)
	}
}

// TestSharedLinksOnOneLine proves that a line linking two shared option pages records both,
// so "git log -p" finds the diff options and "git format-patch" keeps the revision options.
func TestSharedLinksOnOneLine(t *testing.T) {
	k := load(t)
	for _, name := range []string{"log", "show", "whatchanged", "gitk", "format-patch"} {
		c := k.Commands[name]
		if c == nil {
			t.Errorf("%s: missing", name)
			continue
		}
		if !contains(c.Shared, "diff-options") || !contains(c.Shared, "revision-options") {
			t.Errorf("%s: shared = %v, want both diff-options and revision-options", name, c.Shared)
		}
	}
	if _, ok := k.Lookup("log", "", "-p"); !ok {
		t.Error("Lookup(log -p) not found")
	}
}
