package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/msrpc"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"net/url"
)

// TODO import + export

type Model struct {
	trees  map[int]*MetaTree
	trees2 map[*tree.Tree]*MetaTree
	n1     map[string][3]float64
	n2     map[string]float64
	r      map[string]map[string]float64
	t      map[string]map[string]float64
}

var parser *corenlp.Client

func init() {
	u, _ := url.Parse("http://localhost:9000")

	c, err := corenlp.NewClient(u, corenlp.Properties{
		Annotators:   corenlp.Annotators{corenlp.ParserAnnotator},
		OutputFormat: corenlp.FormatJSON,
	})

	if err != nil {
		panic(err)
	}

	parser = c
}

func parse(e string) (*MetaTree, error) {
	doc, err := parser.Annotate(e)

	if err != nil {
		return nil, err
	}

	dec := tree.NewDecoder()
	p := doc.Sentences[0].Parse

	tr, err := dec.Decode(p)

	if err != nil {
		return nil, err
	}

	return NewMetaTree(tr), nil
}

func NewModel() *Model {
	return &Model{
		trees:  make(map[int]*MetaTree),
		trees2: make(map[*tree.Tree]*MetaTree),
		n1:     make(map[string][3]float64),
		n2:     make(map[string]float64),
		r:      make(map[string]map[string]float64),
		t:      make(map[string]map[string]float64),
	}
}

func (m *Model) InitTrees() {
	i, err := msrpc.NewIterator("msr_paraphrase_dummy.txt")

	if err != nil {
		panic(err)
	}

	limit := 1
	counter := 0

	for i.Next() && counter < limit {
		s := i.Sample()

		if !s.Quality {
			continue
		}

		fmt.Println(s.ID1, s.ID2)

		if mt, err := parse(s.String1); err == nil {
			m.trees[s.ID1] = mt
			m.trees2[mt.Tree] = mt
		} else {
			fmt.Println(err)
		}

		if mt, err := parse(s.String2); err == nil {
			m.trees[s.ID2] = mt
			m.trees2[mt.Tree] = mt
		} else {
			fmt.Println(err)
		}

		counter++
	}
}

func (m *Model) InitWeights(nd []string, td map[string][]string) {
	for _, mt := range m.trees {
		mt.Tree.Walk(func(st *tree.Tree) {
			a, ok := mt.Annotation(st)

			if !ok {
				return
			}

			if _, ok := m.n1[a[InsertionFeature]]; !ok {
				ops := Insertions(st, nd, a[InsertionFeature])

				var cN, cL, cR float64 = 0, 0, 0

				for _, op := range ops {
					insertion := op.(Insertion)

					switch insertion.Position {
					case Left:
						cL++
					case Right:
						cR++
					default:
						cN++
					}
				}

				m.n1[a[InsertionFeature]] = [3]float64{
					cN / float64(len(ops)),
					cL / float64(len(ops)),
					cR / float64(len(ops)),
				}
			}

			for _, w := range nd {
				m.n2[w] = 1 / float64(len(nd))
			}

			if _, ok := m.r[a[ReorderingFeature]]; !ok && a[ReorderingFeature] != "" {
				ops := Reorderings(st, a[ReorderingFeature])

				m.r[a[ReorderingFeature]] = make(map[string]float64, len(ops))
				for _, op := range ops {
					m.r[a[ReorderingFeature]][op.Key()] = 1 / float64(len(ops))
				}
			}

			if _, ok := m.t[a[TranslationFeature]]; !ok && a[TranslationFeature] != "" {
				m.t[a[TranslationFeature]] = make(map[string]float64, len(td[a[TranslationFeature]])+1)

				m.t[a[TranslationFeature]][""] = 1 / float64(len(td[a[TranslationFeature]])+1)

				for _, w := range td[a[TranslationFeature]] {
					m.t[a[TranslationFeature]][w] = 1 / float64(len(td[a[TranslationFeature]])+1)
				}
			}
		})
	}
}
