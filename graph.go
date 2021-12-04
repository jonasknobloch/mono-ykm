package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"gonum.org/v1/gonum/stat/combin"
	"strconv"
	"strings"
)

type Graph struct {
	nodes []*Node
	edges map[[2]*Node]float64

	// TODO pAlpha map[*Node]float64
	// TODO pBeta  map[*Node]float64
}

func NewGraph(t *tree.Tree, f []string) {
	g := &Graph{
		nodes: make([]*Node, 0),
		edges: make(map[[2]*Node]float64),
	}

	n := &Node{tr: t, f: f, k: 0, l: len(f), nType: MajorNode}

	g.AddNode(n)
	g.Expand(n)

	fmt.Println(len(g.nodes))
}

func (g *Graph) AddNode(n *Node) {
	g.nodes = append(g.nodes, n)
}

func (g *Graph) AddEdge(n1, n2 *Node) {
	g.edges[[2]*Node{n1, n2}] = 1
}

func iValues(k, l int, f []string) []string {
	ns := make([]string, 0)
	ns = append(ns, "n")

	if l == 0 {
		return ns
	}

	ns = append(ns, "l"+f[k])
	ns = append(ns, "r"+f[k+l-1])

	return ns
}

func rValues(t *tree.Tree) []string {
	rs := make([]string, 0)

	if len(t.Children) == 0 {
		return rs
	}

	join := func(p []int) string {
		sb := strings.Builder{}

		for _, d := range p {
			sb.WriteString(" ")
			sb.WriteString(strconv.Itoa(d))
		}

		return sb.String()[1:]
	}

	g := combin.NewPermutationGenerator(len(t.Children), len(t.Children))

	for g.Next() {
		rs = append(rs, join(g.Permutation(nil)))
	}

	return rs
}

func partitionings(n, k int) [][]int {
	var p func(n, k int, r [][]int) [][]int
	p = func(n, k int, r [][]int) [][]int {
		if k == 1 {
			r = append(r, []int{n})
			return r
		}

		for i := 0; i < n+1; i++ {
			for _, sp := range p(n-i, k-1, make([][]int, 0)) {
				r = append(r, append([]int{i}, sp...))
			}
		}

		return r
	}

	return p(n, k, make([][]int, 0))
}

func (g *Graph) Expand(n *Node) {
	major := make(map[*tree.Tree]map[string]*Node)

	for _, iVal := range iValues(n.k, n.l, n.f) {
		k := n.k
		l := n.l

		if iVal[0:1] == "l" {
			k++
			l--
		}

		if iVal[0:1] == "r" {
			l--
		}

		i := &Node{n: iVal, tr: n.tr, f: n.f, k: k, l: l, nType: SubNode}

		g.AddNode(i)
		g.AddEdge(n, i)

		if len(n.tr.Children) == 0 {
			f := &Node{
				t:     i.Substring(),
				tr:    i.tr,
				nType: FinalNode,
			}

			g.AddNode(f)
			g.AddEdge(i, f)

			continue
		}

		for _, rVal := range rValues(n.tr) {
			r := &Node{n: i.n, r: rVal, tr: i.tr, f: i.f, k: i.k, l: i.l, nType: SubNode}

			g.AddNode(r)
			g.AddEdge(i, r)

			for _, partitioning := range partitionings(r.l, len(r.tr.Children)) {
				if len(partitioning) != len(r.tr.Children) {
					panic("fak")
				}

				p := &Node{n: r.n, r: r.r, p: partitioning, tr: r.tr, f: r.f, k: r.k, l: r.l, nType: SubNode}

				g.AddNode(p)
				g.AddEdge(r, p)

				k := r.k

				for i, c := range r.tr.Children {
					if _, ok := major[c]; !ok {
						major[c] = make(map[string]*Node)
					}

					m := &Node{
						tr:    c,
						f:     p.f,
						k:     k,
						l:     partitioning[i],
						nType: MajorNode,
					}

					k = partitioning[i]

					sub := m.Substring()
					if _, ok := major[c][sub]; !ok {
						major[c][sub] = m
						g.AddNode(m)
					}

					g.AddEdge(p, major[c][sub])
				}
			}
		}
	}

	for _, v := range major {
		for _, node := range v {
			g.Expand(node)
		}
	}
}
