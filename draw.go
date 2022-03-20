package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (g *Graph) Draw(stubs ...string) (int, error) {
	name := fmt.Sprintf("graph_%s.dot", strings.Join(stubs, "-"))

	f, err := os.Create(filepath.Join(Config.GraphExportDirectory, name))

	if err != nil {
		return 0, err
	}

	defer f.Close()

	sb := strings.Builder{}

	sb.WriteString("digraph D {\n")
	sb.WriteString("  node [shape=record]\n")

	for _, n := range g.nodes {
		sb.WriteString(fmt.Sprintf("  PTR%p", n))
		sb.WriteString(" [")

		if n.nType == MajorNode {
			sb.WriteString("color=red label=\"")

			sb.WriteString(fmt.Sprintf("%s ", n.tree.Label))
			sb.WriteString(fmt.Sprintf("| %s ", n.tree.Sentence()))
			sb.WriteString(fmt.Sprintf("| %s ", n.Substring()))
			sb.WriteString(fmt.Sprintf("| { α: %e | β: %e }", g.pAlpha[n], g.pBeta[n]))

			if n.lambda != nil && n.kappa != nil {
				sb.WriteString(fmt.Sprintf("| { λ: %e | κ: %e }", n.lambda, n.kappa))
			}

			if !n.valid {
				sb.WriteString("| pruned")
			}

			sb.WriteString("\"]\n")
		}

		if n.nType == FinalNode {
			sb.WriteString("color=blue label=\"")

			sb.WriteString(fmt.Sprintf("%s ", n.tree.Label))
			sb.WriteString(fmt.Sprintf("| %s ", n.tree.Sentence()))
			sb.WriteString(fmt.Sprintf("| %s ", n.t.Key()))

			if !n.valid {
				sb.WriteString("| pruned")
			}

			sb.WriteString("\"]\n")
		}

		if n.nType == SubNode {
			sb.WriteString("label=\"")

			if len(n.p) > 0 {
				sb.WriteString(fmt.Sprintf("%v ", n.p))
			} else if n.r.Key() != "" {
				sb.WriteString(fmt.Sprintf("%s ", n.r.Key()))
			} else if n.n.Key() != "" {
				sb.WriteString(fmt.Sprintf("%s ", n.n.Key()))
			}

			if !n.valid {
				sb.WriteString("| pruned")
			}

			sb.WriteString("\"]\n")
		}
	}

	for k, v := range g.edges {
		sb.WriteString(fmt.Sprintf("  PTR%p -> PTR%p [label=\"%e\"]\n", k[0], k[1], v))
	}

	sb.WriteString("}\n")

	return f.WriteString(sb.String())
}
