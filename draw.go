package main

import (
	"fmt"
	"os"
)

func (g *Graph) Draw() {
	f, _ := os.Create("graph.dot")
	defer f.Close()

	_, _ = f.WriteString("digraph D {\n")
	_, _ = f.WriteString("  node [shape=record]\n")

	for _, n := range g.nodes {

		if _, ok := g.pruned[n]; ok {
			continue
		}

		_, _ = f.WriteString(fmt.Sprintf("  PTR%p", n))
		_, _ = f.WriteString(" [")

		if n.nType == MajorNode {
			_, _ = f.WriteString("color=red ")
		}

		if n.nType == FinalNode {
			_, _ = f.WriteString("color=blue ")
		}

		_, _ = f.WriteString("label=\"")
		_, _ = f.WriteString(fmt.Sprintf("%s ", n.tree.Label))

		if n.n.key != "" {
			_, _ = f.WriteString(fmt.Sprintf("| %s ", n.n.key))
		}

		if n.r.key != "" {
			_, _ = f.WriteString(fmt.Sprintf("| %s ", n.r.key))
		}

		if len(n.p) > 0 {
			_, _ = f.WriteString(fmt.Sprintf("| %v ", n.p))
		}

		if n.t.key != "" {
			_, _ = f.WriteString(fmt.Sprintf("| %s ", n.t.key))
		}

		_, _ = f.WriteString(fmt.Sprintf("| %s", n.Substring()))

		if n.nType == MajorNode {
			_, _ = f.WriteString(fmt.Sprintf(" | %e | %e", g.pAlpha[n], g.pBeta[n]))
		}

		if _, ok := g.pruned[n]; ok {
			_, _ = f.WriteString(" | PRUNED")
		}

		_, _ = f.WriteString("\"")
		_, _ = f.WriteString("]")
		_, _ = f.WriteString("\n")
	}

	for k, v := range g.edges {
		if _, ok := g.pruned[k[0]]; ok {
			continue
		}

		if _, ok := g.pruned[k[1]]; ok {
			continue
		}

		_, _ = f.WriteString(fmt.Sprintf("  PTR%p -> PTR%p [label=%f]\n", k[0], k[1], v))
	}

	_, _ = f.WriteString("}\n")
}
