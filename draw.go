package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (g *Graph) Draw(stubs ...string) {
	name := fmt.Sprintf("graph_%s.dot", strings.Join(stubs, "-"))

	f, _ := os.Create(filepath.Join(Config.GraphExportDirectory, name))

	defer f.Close()

	_, _ = f.WriteString("digraph D {\n")
	_, _ = f.WriteString("  node [shape=record]\n")

	for _, n := range g.nodes {
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

		if n.n.Key() != "" {
			_, _ = f.WriteString(fmt.Sprintf("| %s ", n.n.key))
		}

		if n.r.Key() != "" {
			_, _ = f.WriteString(fmt.Sprintf("| %s ", n.r.key))
		}

		if len(n.p) > 0 {
			_, _ = f.WriteString(fmt.Sprintf("| %v ", n.p))
		}

		if n.t.Key() != "" {
			_, _ = f.WriteString(fmt.Sprintf("| %s ", n.t.key))
		}

		_, _ = f.WriteString(fmt.Sprintf("| %s", n.Substring()))

		if n.nType == MajorNode {
			_, _ = f.WriteString(fmt.Sprintf(" | %e | %e", g.pAlpha[n], g.pBeta[n]))
		}

		_, _ = f.WriteString("\"")
		_, _ = f.WriteString("]")
		_, _ = f.WriteString("\n")
	}

	for k, v := range g.edges {
		_, _ = f.WriteString(fmt.Sprintf("  PTR%p -> PTR%p [label=\"%e\"]\n", k[0], k[1], v))
	}

	_, _ = f.WriteString("}\n")
}
