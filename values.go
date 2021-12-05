package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"gonum.org/v1/gonum/stat/combin"
	"strconv"
	"strings"
)

func nValues(k, l int, f []string) []string {
	ns := make([]string, 0)
	ns = append(ns, "n")

	return ns

	// TODO implement

	// if l == 0 {
	// 	return ns
	// }
	//
	// ns = append(ns, "l"+f[k])
	// ns = append(ns, "r"+f[k+l-1])
	//
	// return ns
}

func rValues(t *tree.Tree) []string {
	rs := make([]string, 0)

	if len(t.Children) == 0 {
		return rs
	}

	join := func(p []int) string {
		sb := strings.Builder{}

		for _, d := range p {
			sb.WriteString(" ")
			sb.WriteString(strconv.Itoa(d))
		}

		return sb.String()[1:]
	}

	g := combin.NewPermutationGenerator(len(t.Children), len(t.Children))

	for g.Next() {
		rs = append(rs, join(g.Permutation(nil)))
	}

	return rs
}

func tValues() []string {
	panic("tValues not implemented")

	// TODO implement

	return nil
}
