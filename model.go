package main

import (
	"errors"
	"fmt"
	"math/big"
)

type Model struct {
	n map[string]map[string]*big.Float
	r map[string]map[string]*big.Float
	t map[string]map[string]*big.Float
}

func NewModel() *Model {
	return &Model{
		n: make(map[string]map[string]*big.Float),
		r: make(map[string]map[string]*big.Float),
		t: make(map[string]map[string]*big.Float),
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
	if len(m.Table(op)) == 0 {
		return big.NewFloat(0.1)
	}

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

	return new(big.Float)
}

func (m *Model) UpdateWeights(insertionCount, reorderingCount, translationCount *Count) error {
	update := func(p map[string]map[string]*big.Float, c *Count) error {
		for feature, keys := range c.val {
			sum := c.Sum(feature)

			if sum.Cmp(new(big.Float)) == 0 {
				return errors.New("invalid counter sum for feature: " + feature)
			}

			if _, ok := p[feature]; !ok {
				p[feature] = make(map[string]*big.Float, len(c.val[feature]))
			}

			for key := range keys {
				if _, ok := p[feature][key]; !ok {
					p[feature][key] = new(big.Float)
				}

				p[feature][key].Quo(c.Get(feature, key), sum)
			}
		}

		return nil
	}

	if err := update(m.n, insertionCount); err != nil {
		return fmt.Errorf("insertion: %w", err)
	}

	if err := update(m.r, reorderingCount); err != nil {
		return fmt.Errorf("reordering: %w", err)
	}

	if err := update(m.t, translationCount); err != nil {
		return fmt.Errorf("translation: %w", err)
	}

	return nil
}
