package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
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
