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
}

func NewMetaTree(t *tree.Tree) *MetaTree {
	size := len(t.Subtrees())

	m := &MetaTree{
		Tree:    t,
		meta:    make(map[*tree.Tree][3]string, size),
		unknown: make(map[*tree.Tree][3]string, size),
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
