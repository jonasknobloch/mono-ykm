package main

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

type Model struct {
	n map[string]map[string]*big.Float
	r map[string]map[string]*big.Float
	t map[string]map[string]*big.Float

	l map[string]map[string]*big.Float
	f map[string]map[string]*big.Float
}

func NewModel() *Model {
	return &Model{
		n: make(map[string]map[string]*big.Float),
		r: make(map[string]map[string]*big.Float),
		t: make(map[string]map[string]*big.Float),

		l: make(map[string]map[string]*big.Float),
		f: make(map[string]map[string]*big.Float),
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
	probability := func(table map[string]map[string]*big.Float, features, keys []string) *big.Float {
		if len(table) == 0 {
			return big.NewFloat(0.1)
		}

		for _, feature := range features {
			for _, key := range keys {
				if p, ok := table[feature][key]; ok {
					return p
				}
			}
		}

		return new(big.Float)
	}

	// TODO fertility decomposition

	features := []string{op.Feature(), op.UnknownFeature()}
	keys := []string{op.Key(), op.UnknownKey()}

	p := probability(m.Table(op), features, keys)

	if translation, ok := op.(Translation); ok {
		key := strconv.Itoa(translation.Fertility[1])
		fertility := probability(m.f, features, []string{key})

		if translation.Fertility[1] == 0 {
			return fertility
		}

		p = new(big.Float).Mul(fertility, p)
	}

	return p
}

func (m *Model) Lambda(feature string, terminal bool) (*big.Float, *big.Float) {
	var lambda *big.Float
	var kappa *big.Float

	if terminal {
		lambda = big.NewFloat(1)
	} else if len(m.l) == 0 {
		lambda = big.NewFloat(0.5)
	} else {
		if _, ok := m.l[feature]; !ok {
			lambda = new(big.Float)
		} else {
			lambda = m.l[feature][LambdaKey]
			kappa = m.l[feature][KappaKey]
		}
	}

	if kappa == nil {
		kappa = big.NewFloat(1)
		kappa.Sub(kappa, lambda)
	}

	return lambda, kappa
}

func (m *Model) UpdateWeights(insertionCount, reorderingCount, translationCount, lambdaCount, fertilityCount *Count) error {
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

	if err := update(m.l, lambdaCount); err != nil {
		return fmt.Errorf("lambda: %w", err)
	}

	if err := update(m.f, fertilityCount); err != nil {
		return fmt.Errorf("fertility: %w", err)
	}

	return nil
}
