package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"sort"
)

func main() {
	m := NewModel()

	t := &tree.Tree{
		Label: "σ",
		Children: []*tree.Tree{
			{
				Label:    "α",
				Children: nil,
			},
			{
				Label: "γ",
				Children: []*tree.Tree{
					{
						Label:    "β",
						Children: nil,
					},
				},
			},
		},
	}

	mt := NewMetaTree(t)

	m.trees[1234] = mt
	m.trees2[t] = mt

	f := []string{"s", "b", "a"}

	nd := []string{"s", "a", "b"}

	td := map[string][]string{
		"α": {"s", "a", "b"},
		"β": {"s", "a", "b"},
	}

	m.InitWeights(nd, td)

	g := NewGraph(t, f, m)

	fmt.Println(len(g.nodes))
	fmt.Println(len(g.pruned))

	g.Draw()
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
