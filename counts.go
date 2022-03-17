package main

import (
	"math/big"
)

func (g *Graph) InsideWeight(n *Node, filter [3]string, lambda, kappa *big.Float) *big.Float {
	if lambda == nil {
		lambda = n.lambda
	}

	if kappa == nil {
		kappa = n.kappa
	}

	sumI := new(big.Float)

	for _, i := range g.succ[n] {
		if filter[0] != "" && i.n.Key() != filter[0] {
			continue
		}

		sumT := new(big.Float)
		sumR := new(big.Float)

		for _, rt := range g.succ[i] {
			if rt.nType == SubNode {
				if filter[1] != "" && rt.r.Key() != filter[1] {
					continue
				}

				sumP := new(big.Float)

				for _, p := range g.succ[rt] {
					prod := big.NewFloat(1)

					for _, m := range g.succ[p] {
						prod.Mul(prod, g.Beta(m))
					}

					sumP.Add(sumP, prod)
				}

				sumR.Add(sumR, sumP.Mul(sumP, g.edges[[2]*Node{i, rt}]))
			}

			if rt.nType == FinalNode {
				if filter[2] != "" && rt.t.Key() != filter[2] {
					continue
				}

				p := new(big.Float).Copy(g.edges[[2]*Node{i, rt}])

				if rt.t.IsPhrasal() {
					p.Mul(p, lambda)
				} else {
					p.Mul(p, kappa)
				}

				sumT.Add(sumT, p)
			}
		}

		if len(n.tree.Children) != 0 {
			sumR.Mul(sumR, kappa)
		}

		sumI.Add(sumI, sumT.Mul(sumT, g.edges[[2]*Node{n, i}]))
		sumI.Add(sumI, sumR.Mul(sumR, g.edges[[2]*Node{n, i}]))
	}

	return sumI
}

func (g *Graph) InsertionCount(feature, key string) (*big.Float, bool) {
	sum := new(big.Float)

	var ms []*Node
	var ok bool

	if ms, ok = g.insertions[feature][key]; !ok {
		return sum, ok
	}

	for _, m := range ms {
		prod := big.NewFloat(1)

		prod.Mul(prod, g.pAlpha[m])
		prod.Mul(prod, g.InsideWeight(m, [3]string{key}, nil, nil))

		prod.Quo(prod, g.Beta(g.nodes[0]))

		sum.Add(sum, prod)
	}

	return sum, ok
}

func (g *Graph) ReorderingCount(feature, key string) (*big.Float, bool) {
	sum := new(big.Float)

	var ms []*Node
	var ok bool

	if ms, ok = g.reorderings[feature][key]; !ok {
		return sum, ok
	}

	for _, m := range ms {
		prod := big.NewFloat(1)

		prod.Mul(prod, g.pAlpha[m])
		prod.Mul(prod, g.InsideWeight(m, [3]string{"", key}, nil, nil))

		prod.Quo(prod, g.Beta(g.nodes[0]))

		sum.Add(sum, prod)
	}

	return sum, ok
}

func (g *Graph) TranslationCount(feature, key string) (*big.Float, bool) {
	sum := new(big.Float)

	var ms []*Node
	var ok bool

	if ms, ok = g.translations[feature][key]; !ok {
		return sum, ok
	}

	for _, m := range ms {
		prod := big.NewFloat(1)

		prod.Mul(prod, g.pAlpha[m])
		prod.Mul(prod, g.InsideWeight(m, [3]string{"", "", key}, nil, nil))

		prod.Quo(prod, g.Beta(g.nodes[0]))

		sum.Add(sum, prod)
	}

	return sum, ok
}

func (g *Graph) LambdaCount(feature, key string) (*big.Float, bool) {
	sum := new(big.Float)

	var ms []*Node
	var ok bool

	if ms, ok = g.lambda[feature][key]; !ok {
		return sum, ok
	}

	for _, m := range ms {
		prod := big.NewFloat(1)

		prod.Mul(prod, g.pAlpha[m])

		switch key {
		case LambdaKey:
			prod.Mul(prod, g.InsideWeight(m, [3]string{}, nil, new(big.Float)))
		case KappaKey:
			prod.Mul(prod, g.InsideWeight(m, [3]string{}, new(big.Float), nil))
		default:
			panic("unknown key")
		}

		prod.Quo(prod, g.Beta(g.nodes[0]))

		sum.Add(sum, prod)
	}

	return sum, ok
}
