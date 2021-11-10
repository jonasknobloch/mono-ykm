package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"net/url"
	"sort"
)

func main() {
	u, _ := url.Parse("http://localhost:9000")

	c, _ := corenlp.NewClient(u, corenlp.Properties{
		Annotators:   corenlp.Annotators{corenlp.ParserAnnotator},
		OutputFormat: corenlp.FormatJSON,
	})

	d, _ := c.Annotate("first we came to the tall palm trees")

	p := d.Sentences[0].Parse

	dec := tree.NewDecoder()
	t, _ := dec.Decode(p)

	fmt.Println(t.Leaves())

	Insert(t.SubtreeAtPosition([]int{0, 2, 1, 1, 1}), false, "rather")
	Reorder(t.Children[0], []int{2, 0, 1})
	Translate([2]*tree.Tree{nil, t.Leaves()[5]}, "big")

	fmt.Println(t.Leaves())
}

// Insert adds a new child node with the given label
func Insert(t *tree.Tree, x bool, n string) {
	if n == "" {
		return
	}

	c := &tree.Tree{Label: n, Children: nil}

	if x {
		t.Children = append(t.Children, c)
	} else {
		t.Children = append([]*tree.Tree{c}, t.Children...)
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
func Translate(t [2]*tree.Tree, s string) {
	if len(t[1].Children) > 0 {
		panic("not a leaf")
	}

	if s != "" {
		t[1].Label = s
		return
	}

	for k, c := range t[0].Children {
		if c != t[1] {
			continue
		}

		t[1].Children = append(t[1].Children[:k], t[1].Children[k+1:]...)
	}
}
