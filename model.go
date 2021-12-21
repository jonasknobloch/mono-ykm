package main

import (
	"gonum.org/v1/gonum/stat/combin"
)

// TODO import + export

type Model struct {
	n map[string]map[string]float64
	r map[string]map[string]float64
	t map[string]map[string]float64
}

func NewModel() *Model {
	return &Model{
		n: make(map[string]map[string]float64),
		r: make(map[string]map[string]float64),
		t: make(map[string]map[string]float64),
	}
}

func (m *Model) InitInsertionWeights(dictionary map[string]map[string]int) {
	for key, keys := range dictionary {
		m.n[key] = make(map[string]float64, len(keys))

		for value := range keys {
			m.n[key][value] = 1 / float64(len(keys))
		}
	}
}

func (m *Model) InitTranslationWeights(dictionary map[string]map[string]int) {
	for feature, keys := range dictionary {
		m.t[feature] = make(map[string]float64, len(keys))

		for key := range keys {
			m.t[feature][key] = 1 / float64(len(keys))
		}
	}
}

func (m *Model) PInsertion(insertion Insertion) float64 {
	return m.n[insertion.Feature()][insertion.Key()]
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
	sumN := float64(0)

	for feature, keys := range m.n {
		for key := range keys {
			sumN += g.InsertionCount(key, feature)
		}
	}

	for feature, keys := range m.n {
		for key := range keys {
			count := g.InsertionCount(key, feature)

			if count > 0 && sumN > 0 {
				m.n[feature][key] = count / sumN
			}
		}
	}

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
				m.t[feature][key] = count / sumT
			}
		}
	}
}
