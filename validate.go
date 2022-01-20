package main

import (
	"errors"
)

func (g *Graph) ValidateEdges() error {
	inSlice := func(s []*Node, n *Node) bool {
		found := false

		for _, m := range s {
			if m == n {
				found = true
			}
		}

		return found
	}

	for k := range g.edges {
		if !inSlice(g.succ[k[0]], k[1]) {
			return errors.New("edge not in successors")
		}

		if !inSlice(g.pred[k[1]], k[0]) {
			return errors.New("edge not in predecessors")
		}
	}

	return nil
}

func (g *Graph) Orphans() []*Node {
	orphans := make([]*Node, 0)

	for _, n := range g.nodes {
		if _, ok := g.pruned[n]; ok {
			continue
		}

		if len(g.Successor(n)) > 0 {
			continue
		}

		if len(g.Predecessor(n)) > 0 {
			continue
		}

		orphans = append(orphans, n)
	}

	return orphans
}
