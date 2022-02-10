package main

import (
	"fmt"
	"math/big"
)

type Model struct {
	n map[string]map[string]*big.Float
	r map[string]map[string]*big.Float
	t map[string]map[string]*big.Float
}

const ErrorStrategyIgnore = "ignore"
const ErrorStrategyKeep = "keep"
const ErrorStrategyReset = "reset"

func NewModel() *Model {
	return &Model{
		n: make(map[string]map[string]*big.Float),
		r: make(map[string]map[string]*big.Float),
		t: make(map[string]map[string]*big.Float),
	}
}

func (m *Model) InitInsertionWeights(dictionary map[string]map[string]int) {
	for feature, keys := range dictionary {
		m.n[feature] = make(map[string]*big.Float, len(keys))

		for key := range keys {
			m.n[feature][key] = big.NewFloat(1 / float64(len(keys)))
		}
	}
}

func (m *Model) InitReorderingWeights(dictionary map[string]map[string]int) {
	for feature, keys := range dictionary {
		m.r[feature] = make(map[string]*big.Float, len(keys))

		for key := range keys {
			m.r[feature][key] = big.NewFloat(1 / float64(len(keys)))
		}
	}
}

func (m *Model) InitTranslationWeights(dictionary map[string]map[string]int) {
	for feature, keys := range dictionary {
		m.t[feature] = make(map[string]*big.Float, len(keys))

		for key := range keys {
			m.t[feature][key] = big.NewFloat(1 / float64(len(keys)))
		}
	}
}

func (m *Model) Table(op Operation) map[string]map[string]*big.Float {
	switch op.(type) {
	case Insertion:
		return m.n
	case Reordering:
		return m.r
	case Translation:
		return m.t
	default:
		panic("unexpected operation type")
	}
}

func (m *Model) Probability(op Operation) *big.Float {
	return m.Table(op)[op.Feature()][op.Key()]
}

func (m *Model) UpdateWeights(insertionCount, reorderingCount, translationCount *Count) {
	update := func(p map[string]map[string]*big.Float, c *Count) {
		for feature, keys := range p {
			if _, ok := c.val[feature]; !ok {
				continue
			}

			sum := c.Sum(feature)

			if sum == big.NewFloat(0) {
				fmt.Printf("Invalid probability distribution for %s\n", feature)

				var val *big.Float

				switch Config.ModelErrorStrategy {
				case ErrorStrategyIgnore:
					val = big.NewFloat(0)
				case ErrorStrategyReset:
					val = big.NewFloat(1 / float64(c.Size(feature)))
				case ErrorStrategyKeep:
					continue
				}

				for key := range keys {
					p[feature][key].Copy(val)
				}

				continue
			}

			for key := range keys {
				val := c.Get(feature, key)

				if val == nil {
					p[feature][key] = big.NewFloat(0)
					continue
				}

				p[feature][key].Quo(val, sum)
			}
		}
	}

	update(m.n, insertionCount)
	update(m.r, reorderingCount)
	update(m.t, translationCount)
}
