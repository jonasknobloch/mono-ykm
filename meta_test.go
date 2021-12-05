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
		if f, ok := mt.Annotation(st); ok {
			fmt.Printf("n: \"%s\", r: \"%s\", t: \"%s\"\n", f.n, f.r, f.t)
		}
	})

	// TODO implement
}
