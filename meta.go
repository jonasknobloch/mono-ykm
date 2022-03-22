package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
)

type NodeFeature int

const (
	InsertionFeature NodeFeature = iota
	ReorderingFeature
	TranslationFeature
)

type MetaTree struct {
	Tree    *tree.Tree
	meta    map[*tree.Tree][3]string
	unknown map[*tree.Tree][3]string
	maxF    map[*tree.Tree]int
}

func NewMetaTree(t *tree.Tree) *MetaTree {
	size := len(t.Subtrees())

	m := &MetaTree{
		Tree:    t,
		meta:    make(map[*tree.Tree][3]string, size),
		unknown: make(map[*tree.Tree][3]string, size),
		maxF:    make(map[*tree.Tree]int, size),
	}

	return m
}

func (mt *MetaTree) CollectFeatures() {
	var walk func(p, st *tree.Tree)
	walk = func(p, st *tree.Tree) {
		mt.Annotate(st, [3]string{
			nFeature(p, st, false),
			rFeature(st, false),
			tFeature(st, false),
		}, false)

		mt.Annotate(st, [3]string{
			nFeature(p, st, true),
			rFeature(st, true),
			tFeature(st, true),
		}, true)

		for _, c := range st.Children {
			walk(st, c)
		}
	}

	walk(nil, mt.Tree)
}

func (mt *MetaTree) ComputeMaxFertility() {
	var walk func(st *tree.Tree) int
	walk = func(st *tree.Tree) int {
		phrasal := 0
		lexical := 0

		if Config.EnableInteriorInsertions && len(st.Children) != 0 {
			phrasal += 1
			lexical += 1
		}

		if Config.EnableTerminalInsertions && len(st.Children) == 0 {
			lexical += 1
		}

		if len(st.Children) == 0 {
			lexical += 1
		}

		if Config.EnablePhrasalTranslations && len(st.Children) != 0 {
			phrasal += Config.PhraseLengthLimit
		}

		for _, c := range st.Children {
			lexical += walk(c)
		}

		if phrasal > lexical {
			mt.maxF[st] = phrasal

			return phrasal
		}

		mt.maxF[st] = lexical

		return lexical
	}

	walk(mt.Tree)
}

func (mt *MetaTree) Annotate(st *tree.Tree, a [3]string, unknown bool) {
	if unknown {
		mt.unknown[st] = a
	} else {
		mt.meta[st] = a
	}
}

func (mt *MetaTree) Annotation(st *tree.Tree) ([3]string, bool) {
	a, ok := mt.meta[st]
	return a, ok
}

func (mt *MetaTree) Unknown(st *tree.Tree) ([3]string, bool) {
	u, ok := mt.unknown[st]
	return u, ok
}

func (mt *MetaTree) Feature(st *tree.Tree, nf NodeFeature) [2]string {
	a, ok := mt.Annotation(st)

	if !ok {
		panic("unknown feature")
	}

	u, ok := mt.Unknown(st)

	if !ok {
		panic("unknown feature")
	}

	return [2]string{a[nf], u[nf]}
}

func (mt *MetaTree) MaxFertility(st *tree.Tree) int {
	f, ok := mt.maxF[st]

	if !ok {
		panic("unknown subtree")
	}

	return f
}
