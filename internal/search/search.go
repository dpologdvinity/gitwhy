// Package search ranks recipes, commands and flags against a few words.
package search

import (
	"sort"
	"strings"
	"unicode"

	"github.com/dpologdvinity/gitwhy/internal/kb"
	"github.com/dpologdvinity/gitwhy/internal/recipe"
)

// Kind of result.
type Kind int

const (
	Recipe Kind = iota
	Command
	Flag
)

// Result is one search hit.
type Result struct {
	Kind   Kind
	Recipe *recipe.Recipe
	Cmd    string // command name for Command and Flag
	Flag   kb.Flag
	Title  string
	Detail string
	score  float64
}

type doc struct {
	r      Result
	title  []string // words weighted higher
	body   []string
	weight float64
}

// Index is a searchable set of documents.
type Index struct{ docs []doc }

// New builds an index over recipes, commands and their flags.
func New(k *kb.KB, rs []*recipe.Recipe) *Index {
	ix := &Index{}
	for _, r := range rs {
		ix.docs = append(ix.docs, doc{
			r:      Result{Kind: Recipe, Recipe: r, Title: r.Title, Detail: r.Category},
			title:  words(r.Title + " " + strings.Join(r.Aliases, " ")),
			body:   words(r.Category + " " + r.Explain),
			weight: 1.6,
		})
	}
	for _, name := range k.Names() {
		c := k.Commands[name]
		ix.docs = append(ix.docs, doc{
			r:      Result{Kind: Command, Cmd: name, Title: "git " + name, Detail: c.Summary},
			title:  words(name + " " + strings.ReplaceAll(name, "-", " ")),
			body:   words(c.Summary),
			weight: 1.0,
		})
		for _, f := range c.Flags {
			ix.docs = append(ix.docs, doc{
				r:      Result{Kind: Flag, Cmd: name, Flag: f, Title: "git " + name + " " + f.Display, Detail: f.Desc},
				title:  words(name + " " + strings.Join(f.Spellings, " ")),
				body:   words(f.Desc),
				weight: 0.5,
			})
		}
	}
	return ix
}

// Search returns up to n results, best first.
func (ix *Index) Search(query string, n int) []Result {
	q := words(query)
	if len(q) == 0 {
		return nil
	}
	var out []Result
	for _, d := range ix.docs {
		total, matched := 0.0, 0
		for _, w := range q {
			s := max(best(w, d.title)*1.5, best(w, d.body))
			if s > 0 {
				matched++
			}
			total += s
		}
		if matched == 0 {
			continue
		}
		// Prefer documents that match every word.
		cover := float64(matched) / float64(len(q))
		r := d.r
		r.score = total * cover * cover * d.weight
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score > out[j].score })
	if len(out) > n {
		out = out[:n]
	}
	return out
}

// best scores one query word against a document's words.
func best(q string, ws []string) float64 {
	b := 0.0
	for _, w := range ws {
		var s float64
		switch {
		case w == q:
			s = 3
		case strings.HasPrefix(w, q) && len(q) >= 2:
			s = 2
		case len(q) >= 4 && strings.Contains(w, q):
			s = 1
		case len(q) >= 4 && len(w) >= 4 && closeEnough(q, w):
			s = 1 // typo
		}
		if s > b {
			b = s
		}
	}
	return b
}

var stop = map[string]bool{"a": true, "an": true, "the": true, "to": true, "my": true, "i": true, "how": true, "do": true, "of": true, "in": true, "on": true, "it": true, "git": true, "want": true, "and": true, "is": true, "for": true}

func words(s string) []string {
	f := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-'
	})
	var out []string
	for _, w := range f {
		w = strings.Trim(w, "-")
		if w == "" || stop[w] {
			continue
		}
		out = append(out, w)
		if strings.Contains(w, "-") { // "cherry-pick" also matches "cherry pick"
			out = append(out, strings.Split(w, "-")...)
		}
	}
	return out
}

// closeEnough reports an edit distance of at most 1 (one typo) between a and b.
func closeEnough(a, b string) bool {
	if a == b {
		return true
	}
	la, lb := len(a), len(b)
	if la-lb > 1 || lb-la > 1 {
		return false
	}
	i, j, edits := 0, 0, 0
	for i < la && j < lb {
		if a[i] == b[j] {
			i++
			j++
			continue
		}
		edits++
		if edits > 1 {
			return false
		}
		switch {
		case la > lb:
			i++
		case lb > la:
			j++
		default:
			// Swapped neighbors ("comit" vs "commti") count as one edit.
			if i+1 < la && j+1 < lb && a[i] == b[j+1] && a[i+1] == b[j] {
				i += 2
				j += 2
				continue
			}
			i++
			j++
		}
	}
	return edits+(la-i)+(lb-j) <= 1
}
