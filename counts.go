package main

func (g *Graph) CumulativeWeights(n *Node) (float64, float64, float64) {
	if n.nType == MajorNode {
		panic("not a major node")
	}

	sumI := float64(0)
	sumR := float64(0)
	sumP := float64(0)

	for _, i := range g.Successor(n) {
		sumI += g.edges[[2]*Node{n, i}]

		for _, r := range g.Successor(i) {
			sumR += g.edges[[2]*Node{i, r}]

			for _, p := range g.Successor(r) {
				prod := float64(0)

				for _, m := range g.Successor(p) {
					prod *= g.Beta(m)
				}

				sumP += prod
			}
		}
	}

	return sumI, sumR, sumP
}

func (g *Graph) InsertionCount(key string, feature string) float64 {
	sum := float64(0)

	for _, m := range g.insertions[feature] {
		prod := float64(1)

		prod *= g.pAlpha[m]

		for _, i := range g.Successor(m) {
			if i.n.key == key {
				prod *= g.edges[[2]*Node{m, i}]
				break
			}
		}

		_, sumR, sumP := g.CumulativeWeights(m)

		prod *= sumR
		prod *= sumP

		prod /= g.Beta(m)
		sum += prod
	}

	return sum
}

func (g *Graph) ReorderingCount(key string, feature string) float64 {
	sum := float64(0)

	for _, m := range g.reorderings[feature] {
		prod := float64(1)

		prod *= g.pAlpha[m]

		// TODO multiple reorderings with identical keys

		for _, i := range g.Successor(m) {
			for _, r := range g.Successor(i) {
				if r.r.key == key {
					prod *= g.edges[[2]*Node{i, r}]
					break
				}
			}
		}

		sumI, _, sumP := g.CumulativeWeights(m)

		prod *= sumI
		prod *= sumP

		prod /= g.Beta(m)
		sum += prod
	}

	return sum
}

func (g *Graph) TranslationCount(key string, feature string) float64 {
	sum := float64(0)

	for _, m := range g.translations[feature] {
		prod := float64(1)

		prod *= g.pAlpha[m]

		// TODO multiple translations with identical keys

		for _, i := range g.Successor(m) {
			for _, f := range g.Successor(i) {
				if f.t.key == key {
					prod *= g.edges[[2]*Node{i, f}]
				}
			}
		}

		prod /= g.Beta(m)
	}

	return sum
}
