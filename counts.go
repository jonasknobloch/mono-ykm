package main

import (
	"math/big"
)

func (g *Graph) InsideWeightsInterior(n *Node, filter ...string) *big.Float {
	if n.nType != MajorNode {
		panic("not a major node")
	}

	if len(n.tree.Children) == 0 {
		panic("not an interior node node")
	}

	sumI := big.NewFloat(0)

	for _, i := range g.succ[n] {
		if len(filter) > 0 && filter[0] != "" && i.n.Key() != filter[0] {
			continue
		}

		sumR := big.NewFloat(0)

		for _, r := range g.succ[i] {
			if len(filter) > 1 && filter[1] != "" && r.r.Key() != filter[1] {
				continue
			}

			sumP := big.NewFloat(0)

			for _, p := range g.succ[r] {
				prod := big.NewFloat(1)

				for _, m := range g.succ[p] {
					prod.Mul(prod, g.Beta(m))
				}

				sumP.Add(sumP, prod)
			}

			sumR.Add(sumR, sumP.Mul(sumP, g.edges[[2]*Node{i, r}]))
		}

		sumI.Add(sumI, sumR.Mul(sumR, g.edges[[2]*Node{n, i}]))
	}

	return sumI
}

func (g *Graph) InsideWeightsTerminal(n *Node, filter ...string) *big.Float {
	if n.nType != MajorNode {
		panic("not a major node")
	}

	if len(n.tree.Children) > 0 {
		panic("not a terminal node")
	}

	sumI := big.NewFloat(0)

	for _, i := range g.succ[n] {
		if len(filter) > 0 && filter[0] != "" && i.n.Key() != filter[0] {
			continue
		}

		sumT := big.NewFloat(0)

		for _, t := range g.succ[i] {
			if len(filter) > 1 && filter[1] != "" && t.t.Key() != filter[1] {
				continue
			}

			sumT.Add(sumT, g.edges[[2]*Node{i, t}])
		}

		sumI.Add(sumI, sumT.Mul(sumT, g.edges[[2]*Node{n, i}]))
	}

	return sumI
}

func (g *Graph) InsertionCount(feature, key string) (*big.Float, bool) {
	sum := big.NewFloat(0)

	var ms []*Node
	var ok bool

	if ms, ok = g.insertions[feature][key]; !ok {
		return sum, ok
	}

	for _, m := range ms {
		prod := big.NewFloat(1)

		prod.Mul(prod, g.pAlpha[m])

		if len(m.tree.Children) == 0 {
			prod.Mul(prod, g.InsideWeightsTerminal(m, key))
		} else {
			prod.Mul(prod, g.InsideWeightsInterior(m, key))
		}

		prod.Quo(prod, g.Beta(m))

		sum.Add(sum, prod)
	}

	return sum, ok
}

func (g *Graph) ReorderingCount(feature, key string) (*big.Float, bool) {
	sum := big.NewFloat(0)

	var ms []*Node
	var ok bool

	if ms, ok = g.reorderings[feature][key]; !ok {
		return sum, ok
	}

	for _, m := range ms {
		prod := big.NewFloat(1)

		prod.Mul(prod, g.pAlpha[m])
		prod.Mul(prod, g.InsideWeightsInterior(m, "", key))

		prod.Quo(prod, g.Beta(m))

		sum.Add(sum, prod)
	}

	return sum, ok
}

func (g *Graph) TranslationCount(feature, key string) (*big.Float, bool) {
	sum := big.NewFloat(0)

	var ms []*Node
	var ok bool

	if ms, ok = g.translations[feature][key]; !ok {
		return sum, ok
	}

	for _, m := range ms {
		prod := big.NewFloat(1)

		prod.Mul(prod, g.pAlpha[m])
		prod.Mul(prod, g.InsideWeightsTerminal(m, "", key))

		prod.Quo(prod, g.Beta(m))

		sum.Add(sum, prod)
	}

	return sum, ok
}
