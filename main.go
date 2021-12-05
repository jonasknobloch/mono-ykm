package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"sort"
	"strings"
)

func main() {
	m := NewModel()

	m.InitTrees()
	m.InitWeights()

	tr := m.trees[702876].Tree
	f := strings.Split(m.trees[702977].Tree.Sentence(), " ")

	g := NewGraph(tr, f, m)

	fmt.Println(len(g.nodes))
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
