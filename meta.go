package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
)

type Feature struct {
	n string
	r string
	t string
}

type MetaTree struct {
	Tree *tree.Tree
	meta map[*tree.Tree]Feature
}

func NewMetaTree(t *tree.Tree) *MetaTree {
	m := &MetaTree{
		Tree: t,
		meta: make(map[*tree.Tree]Feature, len(t.Subtrees())),
	}

	m.CollectFeatures()

	return m
}

func (mt *MetaTree) CollectFeatures() {
	var walk func(p, st *tree.Tree)
	walk = func(p, st *tree.Tree) {
		var f Feature

		f.n = nFeature(p, st)
		f.r = rFeature(st)
		f.t = tFeature(st)

		mt.Annotate(st, f)

		for _, c := range st.Children {
			walk(st, c)
		}
	}

	walk(nil, mt.Tree)
}

func (mt *MetaTree) Annotate(st *tree.Tree, f Feature) {
	mt.meta[st] = f
}

func (mt *MetaTree) Annotation(st *tree.Tree) (Feature, bool) {
	f, ok := mt.meta[st]
	return f, ok
}
