package main

import (
	"gonum.org/v1/gonum/stat/combin"
)

// TODO import + export

type Model struct {
	n1 map[string][3]float64
	n2 map[string]float64
	r  map[string]map[string]float64
	t  map[string]map[string]float64
}

func NewModel() *Model {
	return &Model{
		n1: make(map[string][3]float64),
		n2: make(map[string]float64),
		r:  make(map[string]map[string]float64),
		t:  make(map[string]map[string]float64),
	}
}

func (m *Model) InitInsertionWeights(list map[string]int) {
	for word := range list {
		m.n2[word] = 1 / float64(len(list))
	}
}

func (m *Model) InitTranslationWeights(dictionary map[string]map[string]int) {
	for source, candidates := range dictionary {

		m.t[source] = make(map[string]float64, len(candidates)+1)

		for target := range candidates {
			m.t[source][target] = 1 / float64(len(candidates)+1)
		}

		m.t[source][""] = 1 / float64(len(candidates)+1)
	}
}

func (m *Model) PInsertion(insertion Insertion, terminal bool) float64 {
	if _, ok := m.n1[insertion.feature]; !ok {
		if terminal {
			m.n1[insertion.feature] = [3]float64{1, 0, 0}
		} else {
			m.n1[insertion.feature] = [3]float64{
				1 / float64(len(m.n2)*2+1),
				float64(len(m.n2)) / float64(len(m.n2)*2+1),
				float64(len(m.n2)) / float64(len(m.n2)*2+1),
			}
		}
	}

	p := float64(1)

	switch insertion.Position {
	case None:
		p *= m.n1[insertion.feature][0]
		break
	case Left:
		p *= m.n1[insertion.feature][1]
		p *= m.n2[insertion.Word]
		break
	case Right:
		p *= m.n1[insertion.feature][2]
		p *= m.n2[insertion.Word]
	}

	return p
}

func (m *Model) PReordering(reordering Reordering) float64 {
	if _, ok := m.r[reordering.feature]; !ok {
		m.r[reordering.feature] = make(map[string]float64)

		n := combin.NumPermutations(len(reordering.Reordering), len(reordering.Reordering))
		g := combin.NewPermutationGenerator(len(reordering.Reordering), len(reordering.Reordering))

		for g.Next() {
			m.r[reordering.feature][NewReordering(g.Permutation(nil), "").key] = 1 / float64(n)
		}
	}

	return m.r[reordering.feature][reordering.key]
}

func (m *Model) PTranslation(translation Translation) float64 {
	return m.t[translation.feature][translation.key]
}

func (m *Model) UpdateWeights(g *Graph) {
	// TODO update insertions

	sumR := float64(0)

	for feature, keys := range m.r {
		for key := range keys {
			sumR += g.ReorderingCount(key, feature)
		}
	}

	for feature, keys := range m.r {
		for key := range keys {
			count := g.ReorderingCount(key, feature)

			if count > 0 && sumR > 0 {
				m.r[feature][key] = count / sumR

			}
		}
	}

	sumT := float64(0)

	for feature, keys := range m.t {
		for key := range keys {
			sumT += g.TranslationCount(key, feature)
		}
	}

	for feature, keys := range m.t {
		for key := range keys {
			count := g.TranslationCount(key, feature)

			if count > 0 && sumT > 0 {
				m.r[feature][key] = count / sumT
			}
		}
	}
}
