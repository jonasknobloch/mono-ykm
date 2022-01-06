package main

import "github.com/jonasknobloch/jinn/pkg/tree"

func nFeature(p, t *tree.Tree) string {
	var f string

	// TODO handle ROOT label

	if p == nil {
		f = "TOP " + t.Label
	} else {
		f = p.Label + " " + t.Label
	}

	return f
}

func rFeature(t *tree.Tree) string {
	if len(t.Children) == 0 {
		return ""
	}

	var f string

	for _, c := range t.Children {
		f += " " + c.Label
	}

	return f[1:]
}

func tFeature(t *tree.Tree) string {
	if len(t.Children) != 0 {
		return ""
	}

	return t.Label
}
