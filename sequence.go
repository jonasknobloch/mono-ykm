package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"gonum.org/v1/gonum/stat/combin"
	"strings"
)

type nodeOperation struct {
	n string
	r []int
	t string
}

type nodeFeature struct {
	n [2]string
	r string
	t string
}

type operations map[*tree.Tree]nodeOperation
type features map[*tree.Tree]nodeFeature

type sequence struct {
	s operations
	f features
}

func Prod(s []int) int {
	p := 1

	for _, f := range s {
		p *= f
	}

	return p
}

func sequences(tr *tree.Tree, e2 string, td, nd []string) (sqs []sequence) {
	nOps := make(map[*tree.Tree]struct {
		r [][]int
		t []string
		n []string
	})

	subtrees := tr.Subtrees()
	dimensions := make([]int, len(subtrees))

	for i, st := range subtrees {
		r, t, n := NodeOperations(st, td, nd)

		nOps[st] = struct {
			r [][]int
			t []string
			n []string
		}{r: r, t: t, n: n}

		dimensions[i] = Prod([]int{len(r), len(t), len(n)})
	}

	sg := NewSubscriptGenerator(dimensions)

	counter := 0

	max := Prod(dimensions) - 1

	for sg.Next() {
		sub := sg.Subscript()
		ops := make(operations, len(sub))

		// fmt.Printf("%d / %d -> %v", counter, max, sub)

		for i, idx := range sub {
			st := subtrees[i]
			dim := []int{len(nOps[st].r), len(nOps[st].t), len(nOps[st].n)}
			sub := make([]int, len(dim))
			SubFor(sub, idx, dim)

			ops[st] = nodeOperation{
				r: nOps[st].r[sub[0]],
				t: nOps[st].t[sub[1]],
				n: nOps[st].n[sub[2]],
			}
		}

		tt, fm := transformTree(tr, ops)

		if s := TreeSentence(tt); s == e2 {
			fmt.Printf("%d / %d -> %v %s\n", counter, max, sub, s)
		}

		sqs = append(sqs, sequence{s: ops, f: fm})

		counter++
	}

	return
}

func NodeOperations(tr *tree.Tree, td, nd []string) (r [][]int, t []string, n []string) {
	if len(tr.Children) == 0 {
		t = td
	} else {
		t = []string{""} // no leaf = no translation
	}

	if len(tr.Children) > 0 {
		n = nd
	} else {
		n = []string{""} // no interior node = no insert
	}

	if len(tr.Children) > 0 {
		r = combin.Permutations(len(tr.Children), len(tr.Children))
	} else {
		r = [][]int{{}} // no children = no reordering
	}

	return
}

func transformTree(tr *tree.Tree, ops operations) (*tree.Tree, features) {
	cp, m := CopyTree(tr)

	fm := make(map[*tree.Tree]nodeFeature, len(ops))

	nF := func(pred, st *tree.Tree) [2]string {
		if pred == nil {
			return [2]string{}
		}

		// TODO handle ROOT node

		return [2]string{
			pred.Label,
			st.Label,
		}
	}

	rF := func(st *tree.Tree) string {
		if len(st.Children) < 2 {
			return ""
		}

		sb := strings.Builder{}

		for _, c := range st.Children {
			sb.WriteString(" ")
			sb.WriteString(c.Label)
		}

		return sb.String()[1:]
	}

	tF := func(st *tree.Tree) string {
		if len(st.Children) != 0 {
			return ""
		}

		return st.Label
	}

	var walk func(pred, st *tree.Tree)
	walk = func(pred, st *tree.Tree) {
		if _, ok := m[st]; !ok {
			return
		}

		if len(st.Children) > 0 {
			Reorder(st, ops[m[st]].r)
			Insert(st, ops[m[st]].n)
		}

		if len(st.Children) == 0 {
			Translate(st, ops[m[st]].t)
		}

		fm[m[st]] = struct {
			n [2]string
			r string
			t string
		}{n: nF(pred, st), r: rF(st), t: tF(st)}

		for _, c := range st.Children {
			walk(st, c)
		}
	}

	walk(nil, cp)
	return cp, fm
}
