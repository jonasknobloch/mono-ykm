package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"gonum.org/v1/gonum/stat/combin"
	"strconv"
	"strings"
)

type Operation interface {
	Key() string
	Feature() string
}

type InsertPosition string

const None InsertPosition = "n"
const Left InsertPosition = "l"
const Right InsertPosition = "r"

type Insertion struct {
	key      string
	feature  string
	Position InsertPosition
	Word     string
}

func NewInsertion(pos InsertPosition, w, f string) Insertion {
	k := string(pos)

	if pos != None {
		k += " " + w
	}

	return Insertion{
		key:      k,
		feature:  f,
		Position: pos,
		Word:     w,
	}
}

func (i Insertion) Key() string {
	return i.key
}

func (i Insertion) Feature() string {
	return i.feature
}

func Insertions(t *tree.Tree, d []string, f string) []Operation {
	ops := make([]Operation, 0)

	ops = append(ops, NewInsertion(None, "", f))

	if len(t.Children) == 0 {
		return ops
	}

	for _, w := range d {
		ops = append(ops, NewInsertion(Left, w, f))
		ops = append(ops, NewInsertion(Right, w, f))
	}

	return ops
}

type Reordering struct {
	key        string
	Reordering []int
	feature    string
}

func NewReordering(p []int, f string) Reordering {
	join := func(p []int) string {
		sb := strings.Builder{}

		for _, d := range p {
			sb.WriteString(" ")
			sb.WriteString(strconv.Itoa(d))
		}

		return sb.String()[1:]
	}

	return Reordering{
		feature:    f[1:],
		key:        join(p),
		Reordering: p,
	}
}

func (r Reordering) Key() string {
	return r.key
}

func (r Reordering) Feature() string {
	return r.feature
}

func Reorderings(t *tree.Tree, f string) []Operation {
	ops := make([]Operation, 0)

	if len(t.Children) == 0 {
		return ops
	}

	g := combin.NewPermutationGenerator(len(t.Children), len(t.Children))

	for g.Next() {
		ops = append(ops, NewReordering(g.Permutation(nil), f))
	}

	return ops
}

type Translation struct {
	key     string
	feature string
	Word    string
}

func NewTranslation(w, f string) Translation {
	return Translation{
		key:     w,
		feature: f,
		Word:    w,
	}
}

func (t Translation) Key() string {
	return t.key
}

func (t Translation) Feature() string {
	return t.feature
}

func Translations(t *tree.Tree, d []string, f string) []Operation {
	ops := make([]Operation, 0)

	if len(t.Children) != 0 {
		return ops
	}

	ops = append(ops, NewTranslation("", f))

	for _, w := range d {
		ops = append(ops, NewTranslation(w, f))
	}

	return ops
}
