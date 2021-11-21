package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"strings"
)

func DropPOSTags(tr *tree.Tree) {
	var walk func(pr, st *tree.Tree)
	walk = func(pr, st *tree.Tree) {
		if pr != nil && len(st.Children) == 1 && len(st.Children[0].Children) == 0 {
			for k, v := range pr.Children {
				if v == st {
					pr.Children[k] = st.Children[0]
					return
				}
			}
		}

		for _, c := range st.Children {
			walk(st, c)
		}
	}

	walk(nil, tr)
}

func CopyTree(tr *tree.Tree) (*tree.Tree, map[*tree.Tree]*tree.Tree) {
	m := make(map[*tree.Tree]*tree.Tree)

	var copyTree func(tr *tree.Tree) *tree.Tree
	copyTree = func(tr *tree.Tree) *tree.Tree {
		if tr == nil {
			return nil
		}

		var children []*tree.Tree

		if len(tr.Children) > 0 {
			children = make([]*tree.Tree, len(tr.Children))

			for k, v := range tr.Children {
				children[k] = copyTree(v)
			}
		}

		cp := &tree.Tree{
			Label:    tr.Label,
			Children: children,
		}

		m[cp] = tr
		return cp
	}

	return copyTree(tr), m
}

func TreeSentence(tr *tree.Tree) string {
	var sb strings.Builder

	for _, n := range tr.Leaves() {
		if n.Label == "" {
			continue
		}

		sb.WriteString(" ")
		sb.WriteString(n.Label)
	}

	if sb.Len() == 0 {
		return ""
	}

	return sb.String()[1:]
}
