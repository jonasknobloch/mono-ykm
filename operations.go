package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"gonum.org/v1/gonum/stat/combin"
	"strconv"
	"strings"
)

type Operation interface {
	Feature() string
	Key() string

	UnknownFeature() string
	UnknownKey() string
}

type InsertPosition string

const None InsertPosition = "n"
const Left InsertPosition = "l"
const Right InsertPosition = "r"

type Insertion struct {
	feature [2]string
	key     [2]string

	Position InsertPosition
	Word     string
}

func NewInsertion(pos InsertPosition, word string, feature [2]string) Insertion {
	key := func(word string) string {
		k := string(pos)

		if pos != None {
			k += " " + word
		}

		return k
	}

	n := Insertion{
		feature:  feature,
		Position: pos,
		Word:     word,
	}

	n.key = [2]string{key(word), key(UnknownToken)}

	return n
}

func (i Insertion) Feature() string {
	return i.feature[0]
}

func (i Insertion) Key() string {
	return i.key[0]
}

func (i Insertion) UnknownFeature() string {
	return i.feature[1]
}

func (i Insertion) UnknownKey() string {
	return i.key[1]
}

func Insertions(t *tree.Tree, d []string, f [2]string) []Operation {
	ops := make([]Operation, 0)

	if len(d) < t.Size()+1 {
		ops = append(ops, NewInsertion(None, "", f))
	}

	if len(d) > 0 {
		ops = append(ops, NewInsertion(Left, d[0], f))
		ops = append(ops, NewInsertion(Right, d[len(d)-1], f))
	}

	return ops
}

type Reordering struct {
	feature [2]string
	key     [2]string

	Reordering []int
}

func NewReordering(reordering []int, feature [2]string) Reordering {
	join := func(p []int) string {
		sb := strings.Builder{}

		for _, d := range p {
			sb.WriteString(" ")
			sb.WriteString(strconv.Itoa(d))
		}

		return sb.String()[1:]
	}

	r := Reordering{
		feature:    feature,
		Reordering: reordering,
	}

	key := join(reordering)

	r.key = [2]string{key, key}

	return r
}

func (r Reordering) Feature() string {
	return r.feature[0]
}

func (r Reordering) Key() string {
	return r.key[0]
}

func (r Reordering) UnknownFeature() string {
	return r.feature[1]
}

func (r Reordering) UnknownKey() string {
	return r.key[1]
}

func Reorderings(t *tree.Tree, f [2]string) []Operation {
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
	feature [2]string
	key     [2]string

	Word string
}

const NullToken = "$NULL$"

func NewTranslation(word string, feature [2]string) Translation {
	if word == "" {
		word = NullToken
	}

	t := Translation{
		feature: feature,
		Word:    word,
	}

	t.key = [2]string{word, UnknownToken}

	return t
}

func (t Translation) Feature() string {
	return t.feature[0]
}

func (t Translation) Key() string {
	return t.key[0]
}

func (t Translation) UnknownFeature() string {
	return t.feature[1]
}

func (t Translation) UnknownKey() string {
	return t.key[1]
}
