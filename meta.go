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
	Tree *tree.Tree
	meta map[*tree.Tree][3]string
}

func NewMetaTree(t *tree.Tree) *MetaTree {
	m := &MetaTree{
		Tree: t,
		meta: make(map[*tree.Tree][3]string, len(t.Subtrees())),
	}

	return m
}

func (mt *MetaTree) CollectFeatures(replaceUnknownTokens bool) {
	var walk func(p, st *tree.Tree)
	walk = func(p, st *tree.Tree) {
		mt.Annotate(st, [3]string{
			nFeature(p, st, replaceUnknownTokens),
			rFeature(st, replaceUnknownTokens),
			tFeature(st, replaceUnknownTokens),
		})

		for _, c := range st.Children {
			walk(st, c)
		}
	}

	walk(nil, mt.Tree)
}

func (mt *MetaTree) Annotate(st *tree.Tree, a [3]string) {
	mt.meta[st] = a
}

func (mt *MetaTree) Annotation(st *tree.Tree) ([3]string, bool) {
	a, ok := mt.meta[st]
	return a, ok
}
