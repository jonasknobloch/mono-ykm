package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"testing"
)

func TestNewMetaTree(t *testing.T) {
	dec := tree.NewDecoder()
	tr, _ := dec.Decode("(FOO (BAR a) (BAZ b))")

	mt := NewMetaTree(tr)

	mt.Tree.Walk(func(st *tree.Tree) {
		if a, ok := mt.Annotation(st); ok {
			fmt.Printf("n: \"%s\", r: \"%s\", t: \"%s\"\n", a[InsertionFeature], a[ReorderingFeature], a[TranslationFeature])
		}
	})

	// TODO implement
}
