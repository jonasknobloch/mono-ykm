package main

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
	for feature, keys := range dictionary {
		m.n[feature] = make(map[string]float64, len(keys))

		for key := range keys {
			m.n[feature][key] = 1 / float64(len(keys))
		}
	}
}

func (m *Model) InitReorderingWeights(dictionary map[string]map[string]int) {
	for feature, keys := range dictionary {
		m.r[feature] = make(map[string]float64, len(keys))

		for key := range keys {
			m.r[feature][key] = 1 / float64(len(keys))
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
	return m.r[reordering.Feature()][reordering.Key()]
}

func (m *Model) PTranslation(translation Translation) float64 {
	return m.t[translation.feature][translation.key]
}

func (m *Model) UpdateWeights(insertionCount, reorderingCount, translationCount *Count) {
	update := func(p map[string]map[string]float64, c *Count) {
		for feature, keys := range p {
			for key := range keys {
				p[feature][key] = c.Get(feature, key) / c.Sum(feature)
			}
		}
	}

	update(m.n, insertionCount)
	update(m.r, reorderingCount)
	update(m.t, translationCount)
}
