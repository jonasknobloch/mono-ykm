package main

import "github.com/jonasknobloch/jinn/pkg/tree"

type Sample struct {
	t *tree.Tree
	e string
}

type Corpus []Sample

func NewCorpus() Corpus {
	return Corpus{}
}
