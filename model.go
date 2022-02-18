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

func (m *Model) InitTable(t map[string]map[string]*big.Float, d map[string]map[string]int) {
	for feature, keys := range d {
		t[feature] = make(map[string]*big.Float, len(keys))

		val := big.NewFloat(1 / float64(len(keys)))

		for key := range keys {
			t[feature][key] = val
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
	if p, ok := m.Table(op)[op.Feature()][op.Key()]; ok {
		return p
	}

	if p, ok := m.Table(op)[op.Feature()][op.UnknownKey()]; ok {
		return p
	}

	if p, ok := m.Table(op)[op.UnknownFeature()][op.Key()]; ok {
		return p
	}

	if p, ok := m.Table(op)[op.UnknownFeature()][op.UnknownKey()]; ok {
		return p
	}

	return big.NewFloat(0)
}

func (m *Model) UpdateWeights(insertionCount, reorderingCount, translationCount *Count) {
	update := func(p map[string]map[string]*big.Float, c *Count) {
		for feature, keys := range p {
			if _, ok := c.val[feature]; !ok {
				continue
			}

			sum := c.Sum(feature)

			if sum.Cmp(big.NewFloat(0)) == 0 {
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
