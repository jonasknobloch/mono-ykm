package main

func (g *Graph) InsideWeightsInterior(n *Node, filter ...string) float64 {
	if n.nType != MajorNode {
		panic("not a major node")
	}

	if len(n.tree.Children) == 0 {
		panic("not an interior node node")
	}

	sumI := float64(0)

	for _, i := range g.Successor(n) {
		if len(filter) > 0 && filter[0] != "" && i.n.key != filter[0] {
			continue
		}

		sumR := float64(0)

		for _, r := range g.Successor(i) {
			if len(filter) > 1 && filter[1] != "" && r.r.key != filter[1] {
				continue
			}

			sumP := float64(0)

			for _, p := range g.Successor(r) {
				prod := float64(1)

				for _, m := range g.Successor(p) {
					prod *= g.Beta(m)
				}

				sumP += prod
			}

			sumR += g.edges[[2]*Node{i, r}] * sumP
		}

		sumI += g.edges[[2]*Node{n, i}] * sumR
	}

	return sumI
}

func (g *Graph) InsideWeightsTerminal(n *Node, filter ...string) float64 {
	if n.nType != MajorNode {
		panic("not a major node")
	}

	if len(n.tree.Children) > 0 {
		panic("not a terminal node")
	}

	sumI := float64(0)

	for _, i := range g.Successor(n) {
		if len(filter) > 0 && filter[0] != "" && i.n.key != filter[0] {
			continue
		}

		sumT := float64(0)

		for _, t := range g.Successor(i) {
			if len(filter) > 1 && filter[1] != "" && t.t.key != filter[1] {
				continue
			}

			sumT += g.edges[[2]*Node{i, t}]
		}

		sumI += g.edges[[2]*Node{n, i}] * sumT
	}

	return sumI
}

func (g *Graph) InsertionCount(feature, key string) float64 {
	sum := float64(0)

	for _, m := range g.InsertionCandidateNodes(feature) {
		prod := float64(1)

		prod *= g.pAlpha[m]

		if len(m.tree.Children) == 0 {
			prod *= g.InsideWeightsTerminal(m, key)
		} else {
			prod *= g.InsideWeightsInterior(m, key)
		}

		prod /= g.Beta(m)

		sum += prod
	}

	return sum
}

func (g *Graph) ReorderingCount(feature, key string) float64 {
	sum := float64(0)

	for _, m := range g.ReorderingCandidateNodes(feature) {
		prod := float64(1)

		prod *= g.pAlpha[m]
		prod *= g.InsideWeightsInterior(m, "", key)

		prod /= g.Beta(m)

		sum += prod
	}

	return sum
}

func (g *Graph) TranslationCount(feature, key string) float64 {
	sum := float64(0)

	for _, m := range g.TranslationCandidateNode(feature) {
		prod := float64(1)

		prod *= g.pAlpha[m]
		prod *= g.InsideWeightsTerminal(m, "", key)

		prod /= g.Beta(m)

		sum += prod
	}

	return sum
}
