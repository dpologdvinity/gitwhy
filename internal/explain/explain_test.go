package explain

import (
	"os"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/kb"
)

func TestExplain(t *testing.T) {
	k, err := kb.Load(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		cmd  string
		want []string // substrings of each part's meaning, "" = anything
	}{
		{"reset --soft HEAD~1", []string{"", "", "1 commit(s) before"}},
		{"commit -m hello", []string{"", "", "value for -m"}},
		{"clean -fdx", []string{"", "", "", ""}},
		{"rebase -i origin/main", []string{"", "", ""}},
		{"diff main...feature", []string{"", "since they split"}},
		{"checkout --ours -- a.txt", []string{"", "", "end of options", "file or folder"}},
		{"stash pop stash@{1}", []string{"", "", "stash number 1"}},
		{"commit --amend --no-edit", []string{"", "", "Turns off --edit"}},
		{"bisect run sh -c make test", []string{"", "", "test command"}},
	}
	for _, c := range cases {
		parts := Explain(k, nil, strings.Fields(c.cmd))
		if len(parts) != len(c.want) {
			t.Errorf("%q: got %d parts %+v", c.cmd, len(parts), parts)
			continue
		}
		for i, w := range c.want {
			if strings.Contains(parts[i].Meaning, "not in gitwhy's docs") || parts[i].Meaning == "" {
				t.Errorf("%q part %q: no meaning (%q)", c.cmd, parts[i].Token, parts[i].Meaning)
			}
			if w != "" && !strings.Contains(parts[i].Meaning, w) {
				t.Errorf("%q part %q = %q, want %q", c.cmd, parts[i].Token, parts[i].Meaning, w)
			}
		}
	}
}
