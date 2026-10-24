// Package kb parses the docs/ markdown into commands and flags.
package kb

import (
	"bufio"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Flag is one row of a flag table.
type Flag struct {
	Display   string   // first cell as written, without backticks: "-a, --all"
	Spellings []string // every accepted spelling: "-a", "--all", "--no-all"
	Value     string   // value placeholder, e.g. "<who>"; "" if none
	Optional  bool     // value may be left out ("--x[=<v>]")
	Desc      string
}

// Sub is a subcommand with its own flags.
type Sub struct {
	Name    string
	Summary string
	Usage   []string
	Flags   []Flag
}

// Command is one docs/commands/<name>.md file.
type Command struct {
	Name    string
	Summary string // first paragraph
	Usage   []string
	Flags   []Flag
	Subs    []*Sub
	Shared  []string // shared option pages it links to: "diff-options", "revision-options"
}

// Sub returns the named subcommand or nil.
func (c *Command) Sub(name string) *Sub {
	for _, s := range c.Subs {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// KB is the whole knowledge base.
type KB struct {
	Commands map[string]*Command
	Shared   map[string][]Flag // "diff-options" -> flags
}

// Names returns command names sorted.
func (k *KB) Names() []string {
	names := make([]string, 0, len(k.Commands))
	for n := range k.Commands {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// FlagsFor returns every flag a command (or subcommand, if sub != "") accepts, including shared pages.
func (k *KB) FlagsFor(cmd, sub string) []Flag {
	c := k.Commands[cmd]
	if c == nil {
		return nil
	}
	var flags []Flag
	if s := c.Sub(sub); s != nil {
		flags = append(flags, s.Flags...)
	}
	flags = append(flags, c.Flags...)
	for _, sh := range c.Shared {
		flags = append(flags, k.Shared[sh]...)
	}
	return flags
}

// Lookup finds the flag matching an argument like "--soft", "-q" or "--author=bob".
func (k *KB) Lookup(cmd, sub, arg string) (Flag, bool) {
	name := arg
	if i := strings.IndexByte(arg, '='); i > 0 {
		name = arg[:i]
	}
	for _, f := range k.FlagsFor(cmd, sub) {
		for _, s := range f.Spellings {
			if s == name {
				return f, true
			}
		}
	}
	return Flag{}, false
}

// Load parses docs/commands/*.md and docs/shared/*.md from fsys.
func Load(fsys fs.FS) (*KB, error) {
	k := &KB{Commands: map[string]*Command{}, Shared: map[string][]Flag{}}
	cmds, err := fs.Glob(fsys, "docs/commands/*.md")
	if err != nil {
		return nil, err
	}
	for _, p := range cmds {
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c := Parse(string(data))
		if c.Name == "" {
			c.Name = strings.TrimSuffix(path.Base(p), ".md")
		}
		k.Commands[c.Name] = c
	}
	shared, _ := fs.Glob(fsys, "docs/shared/*.md")
	for _, p := range shared {
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, err
		}
		k.Shared[strings.TrimSuffix(path.Base(p), ".md")] = Parse(string(data)).Flags
	}
	return k, nil
}

// Patterns for Parse: codeRe finds `code` spans; sharedRe finds links to a shared options page
// and captures its name ("diff-options").
var (
	codeRe   = regexp.MustCompile("`([^`]+)`")
	sharedRe = regexp.MustCompile(`\]\((?:\.\./)?shared/([\w-]+)\.md`)
)

// Parse parses one command file. Name is "" for files without "# git <name>".
func Parse(md string) *Command {
	c := &Command{}
	var sub *Sub        // current "### git <cmd> <sub>" section
	inSubTable := false // inside the "## Subcommands" table
	summaryDone := false
	var para []string

	sc := bufio.NewScanner(strings.NewReader(md))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	inFence := false
	for sc.Scan() {
		line := sc.Text()
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		switch {
		case strings.HasPrefix(line, "# "):
			title := strings.TrimSpace(line[2:])
			if strings.HasPrefix(title, "git ") {
				c.Name = strings.TrimSpace(title[4:])
			} else if title == "git" {
				c.Name = "git"
			}
			continue
		case strings.HasPrefix(line, "### "):
			h := strings.TrimSpace(line[4:])
			inSubTable = false
			if name, ok := subName(c.Name, h); ok {
				sub = c.Sub(name)
				if sub == nil {
					sub = &Sub{Name: name}
					c.Subs = append(c.Subs, sub)
				}
			}
			continue
		case strings.HasPrefix(line, "## "):
			h := strings.ToLower(strings.TrimSpace(line[3:]))
			inSubTable = strings.HasPrefix(h, "subcommands")
			sub = nil
			continue
		case strings.HasPrefix(line, "#### "):
			continue
		}

		for _, m := range sharedRe.FindAllStringSubmatch(line, -1) {
			if !contains(c.Shared, m[1]) {
				c.Shared = append(c.Shared, m[1])
			}
		}

		if !summaryDone {
			if trim == "" {
				if len(para) > 0 {
					c.Summary = strings.Join(para, " ")
					summaryDone = true
				}
			} else if !strings.HasPrefix(trim, "|") && !strings.HasPrefix(trim, "**Usage") {
				para = append(para, trim)
				continue
			}
		}

		if strings.HasPrefix(trim, "**Usage:**") || (strings.HasPrefix(trim, "- `git ") && !strings.HasPrefix(trim, "|")) {
			for _, u := range codeRe.FindAllStringSubmatch(trim, -1) {
				if strings.HasPrefix(u[1], "git ") {
					if sub != nil {
						sub.Usage = append(sub.Usage, u[1])
					} else {
						c.Usage = append(c.Usage, u[1])
					}
				}
			}
			continue
		}

		if !strings.HasPrefix(trim, "| `") {
			continue
		}
		cells := splitRow(trim)
		if len(cells) < 2 {
			continue
		}
		first, desc := cells[0], cells[1]
		if inSubTable {
			if m := codeRe.FindStringSubmatch(first); m != nil {
				name := strings.Fields(m[1])[0]
				if s := c.Sub(name); s == nil {
					c.Subs = append(c.Subs, &Sub{Name: name, Summary: desc})
				} else if s.Summary == "" {
					s.Summary = desc
				}
			}
			continue
		}
		f, ok := parseFlag(first, desc)
		if !ok {
			continue
		}
		if sub != nil {
			sub.Flags = append(sub.Flags, f)
		} else {
			c.Flags = append(c.Flags, f)
		}
	}
	if !summaryDone && len(para) > 0 {
		c.Summary = strings.Join(para, " ")
	}
	return c
}

// subName extracts "push" from "git stash push" when the command is "stash".
func subName(cmd, heading string) (string, bool) {
	h := strings.Trim(heading, "` ")
	prefix := "git " + cmd + " "
	if cmd == "" || !strings.HasPrefix(h, prefix) {
		return "", false
	}
	rest := strings.Fields(h[len(prefix):])
	if len(rest) == 0 {
		return "", false
	}
	return rest[0], true
}

// splitRow splits a markdown table row on unescaped pipes.
func splitRow(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	var cells []string
	var cur strings.Builder
	for i := 0; i < len(row); i++ {
		if row[i] == '\\' && i+1 < len(row) && row[i+1] == '|' {
			cur.WriteByte('|')
			i++
			continue
		}
		if row[i] == '|' {
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteByte(row[i])
	}
	return append(cells, strings.TrimSpace(cur.String()))
}

// parseFlag reads a first cell like "`-a`, `--all`" or "`--[no-]verify`" or "`--author=<who>`".
func parseFlag(cell, desc string) (Flag, bool) {
	codes := codeRe.FindAllStringSubmatch(cell, -1)
	if len(codes) == 0 || !strings.HasPrefix(codes[0][1], "-") {
		return Flag{}, false
	}
	f := Flag{Desc: desc}
	var display []string
	for _, cm := range codes {
		spell := strings.TrimSpace(cm[1])
		display = append(display, spell)
		// "-a <x>", "--x=<v>", "--x[=<v>]"
		head, value := spell, ""
		if i := strings.Index(spell, "[<"); i > 0 && !strings.Contains(spell[:i], "=") { // "-S[<key-id>]"
			head, value, f.Optional = spell[:i], strings.Trim(spell[i:], "[]"), true
		} else if i := strings.Index(spell, "[="); i > 0 {
			head, value, f.Optional = spell[:i], strings.TrimSuffix(spell[i+2:], "]"), true
		} else if i := strings.IndexAny(spell, " ="); i > 0 {
			head, value = spell[:i], spell[i+1:]
		}
		if value != "" && f.Value == "" {
			f.Value = value
		}
		if strings.Contains(head, "[no-]") {
			f.Spellings = append(f.Spellings, strings.Replace(head, "[no-]", "", 1), strings.Replace(head, "[no-]", "no-", 1))
		} else {
			f.Spellings = append(f.Spellings, head)
		}
	}
	f.Display = strings.Join(display, ", ")
	return f, true
}

// contains reports whether xs has x.
func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
