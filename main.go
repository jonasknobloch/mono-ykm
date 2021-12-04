package main

import (
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"net/url"
	"sort"
	"strings"
)

func main() {
	u, _ := url.Parse("http://localhost:9000")

	c, _ := corenlp.NewClient(u, corenlp.Properties{
		Annotators:   corenlp.Annotators{corenlp.ParserAnnotator},
		OutputFormat: corenlp.FormatJSON,
	})

	e := "I love dogs"
	f := "I really like dogs"

	doc, _ := c.Annotate(e)

	p := doc.Sentences[0].Parse

	dec := tree.NewDecoder()
	tr, _ := dec.Decode(p)

	NewGraph(tr, strings.Split(f, " "))
}

// Insert adds a new child node with the given label
func Insert(t *tree.Tree, n string) {
	if n == "" {
		return
	}

	if n == "n" {
		return
	}

	c := &tree.Tree{Label: n[1:], Children: nil}

	if n[:1] == "l" {
		t.Children = append([]*tree.Tree{c}, t.Children...)
	}

	if n[:1] == "r" {
		t.Children = append(t.Children, c)
	}
}

// Reorder applies the given permutation
func Reorder(t *tree.Tree, p []int) {
	if len(t.Children) != len(p) {
		panic("unsuitable permutation")
	}

	if sort.IntsAreSorted(p) {
		return
	}

	c := make([]*tree.Tree, len(t.Children))

	for k, v := range p {
		c[v] = t.Children[k]
	}

	t.Children = c
}

// Translate replaces the given leaf's label
func Translate(t *tree.Tree, s string) {
	if len(t.Children) > 0 {
		panic("not a leaf")
	}

	t.Label = s
	return
}
