package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"strings"
)

func nFeature(p, st *tree.Tree, replaceLeafs bool) string {
	var feature string
	var label string

	if replaceLeafs && len(st.Children) == 0 {
		label = UnknownToken
	} else {
		label = st.Label
	}

	if p == nil {
		feature = "ROOT " + label
	} else {
		feature = p.Label + " " + label
	}

	return feature
}

func rFeature(st *tree.Tree, replaceLeafs bool) string {
	if len(st.Children) == 0 {
		return ""
	}

	var feature string

	for _, c := range st.Children {
		feature += " "

		if replaceLeafs && len(c.Children) == 0 {
			feature += UnknownToken
		} else {
			feature += c.Label
		}
	}

	return feature[1:]
}

func tFeature(st *tree.Tree, replaceLeafs bool) string {
	var sb strings.Builder

	for i, leaf := range st.Leaves() {
		if i > 0 {
			sb.WriteString(" ")
		}

		if replaceLeafs {
			sb.WriteString(UnknownToken)
		} else {
			sb.WriteString(leaf.Label)
		}
	}

	return sb.String()
}
