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

func (m *Model) InitTable(t map[string]map[string]*big.Float, d map[string]map[string]int) {
	for feature, keys := range d {
		t[feature] = make(map[string]*big.Float, len(keys))

		val := big.NewFloat(1 / float64(len(keys)))

		for key := range keys {
			t[feature][key] = new(big.Float).Copy(val)
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

	return new(big.Float)
}

func (m *Model) UpdateWeights(insertionCount, reorderingCount, translationCount *Count) error {
	update := func(p map[string]map[string]*big.Float, c *Count) error {
		for feature, keys := range p {
			if _, ok := c.val[feature]; !ok {
				continue
			}

			sum := c.Sum(feature)

			if sum.Cmp(new(big.Float)) == 0 {
				return errors.New("invalid counter sum for feature: " + feature)
			}

			for key := range keys {
				val := c.Get(feature, key)

				if val == nil {
					p[feature][key].SetFloat64(0)
					continue
				}

				p[feature][key].Quo(val, sum)
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
