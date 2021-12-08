package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
)

type Graph struct {
	root   *Node
	nodes  []*Node
	edges  map[[2]*Node]float64
	pred   map[*Node][]*Node
	succ   map[*Node][]*Node
	major  map[*Node]*Node
	pAlpha map[*Node]float64
	pBeta  map[*Node]float64
}

func NewGraph(t *tree.Tree, f []string, m *Model) *Graph {
	n := &Node{tree: t, f: f, k: 0, l: len(f), nType: MajorNode}

	g := &Graph{
		root:   n,
		nodes:  make([]*Node, 0),
		edges:  make(map[[2]*Node]float64),
		pred:   make(map[*Node][]*Node),
		succ:   make(map[*Node][]*Node),
		major:  make(map[*Node]*Node), // partitioning -> parent major
		pAlpha: make(map[*Node]float64),
		pBeta:  make(map[*Node]float64),
	}

	g.AddNode(n)
	g.Expand(n, m)

	g.Beta(n)

	for _, node := range g.nodes {
		if node.nType != FinalNode {
			continue
		}

		g.Alpha(g.pred[g.pred[node][0]][0])
	}

	return g
}

func (g *Graph) AddNode(n *Node) {
	g.nodes = append(g.nodes, n)
}

func (g *Graph) AddEdge(n1, n2 *Node, w float64) {
	g.edges[[2]*Node{n1, n2}] = w

	if g.pred[n2] == nil {
		g.pred[n2] = make([]*Node, 0)
	}

	if g.succ[n1] == nil {
		g.succ[n2] = make([]*Node, 0)
	}

	g.pred[n2] = append(g.pred[n2], n1)
	g.succ[n1] = append(g.succ[n1], n2)
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

func (g *Graph) Expand(n *Node, m *Model) {
	major := make(map[*tree.Tree]map[string]*Node)

	feats, ok := m.trees2[g.root.tree].Annotation(n.tree)

	if !ok {
		panic("unknown feature")
	}

	d := make([]string, 0)

	if n.l > 0 {
		d = append(d, n.f[n.k])
	}

	if n.l > 1 {
		d = append(d, n.f[n.l-1])
	}

	for _, op := range Insertions(n.tree, d, feats[InsertionFeature]) {
		insertion := op.(Insertion)

		k := n.k
		l := n.l

		if insertion.Position == Left {
			k++
			l--
		}

		if insertion.Position == Right {
			l--
		}

		i := &Node{
			n:     insertion,
			tree:  n.tree,
			f:     n.f,
			k:     k,
			l:     l,
			nType: SubNode,
		}

		w := float64(1)

		switch insertion.Position {
		case Left:
			w *= m.n1[feats[InsertionFeature]][1]
			w *= m.n2[insertion.Word] // TODO n2 empty
		case Right:
			w *= m.n1[feats[InsertionFeature]][2]
			w *= m.n2[insertion.Word] // TODO n2 empty
		default:
			w *= m.n1[feats[InsertionFeature]][0]
		}

		g.AddNode(i)
		g.AddEdge(n, i, w)

		if len(n.tree.Children) == 0 {
			translation := NewTranslation(i.Substring(), feats[TranslationFeature])

			f := &Node{
				t:     translation,
				tree:  i.tree,
				nType: FinalNode,
			}

			w := float64(1)

			if v, ok := m.t[feats[TranslationFeature]][translation.Key()]; ok {
				w = v // TODO t empty
			}

			g.AddNode(f)
			g.AddEdge(i, f, w)

			continue
		}

		for _, op := range Reorderings(n.tree, feats[ReorderingFeature]) {
			reordering := op.(Reordering)

			r := &Node{
				n:     i.n,
				r:     reordering,
				tree:  i.tree,
				f:     i.f,
				k:     i.k,
				l:     i.l,
				nType: SubNode,
			}

			g.AddNode(r)
			g.AddEdge(i, r, m.r[feats[ReorderingFeature]][reordering.Key()]) // TODO possible map access error?

			for _, partitioning := range partitionings(r.l, len(r.tree.Children)) {
				p := &Node{
					n:     r.n,
					r:     r.r,
					p:     partitioning,
					tree:  r.tree,
					f:     r.f,
					k:     r.k,
					l:     r.l,
					nType: SubNode,
				}

				g.AddNode(p)
				g.AddEdge(r, p, 1)

				g.major[p] = n

				k := r.k

				for i := 0; i < len(r.tree.Children); i++ {
					c := r.tree.Children[p.r.Reordering[i]]

					if _, ok := major[c]; !ok {
						major[c] = make(map[string]*Node)
					}

					m := &Node{
						tree:  c,
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

					g.AddEdge(p, major[c][sub], 1)
				}
			}
		}
	}

	for _, v := range major {
		for _, node := range v {
			g.Expand(node, m)
		}
	}
}

func (g *Graph) Alpha(n *Node) float64 {
	if a, ok := g.pAlpha[n]; ok {
		return a
	}

	if n == g.root {
		g.pAlpha[n] = float64(1)

		return float64(1)
	}

	sum := float64(0)

	for _, parent := range g.pred[n] {
		mp := g.major[parent]
		prod := g.Alpha(mp)

		for _, i := range g.succ[mp] {
			prod *= g.edges[[2]*Node{mp, i}]

			for _, r := range g.succ[i] {
				prod *= g.edges[[2]*Node{i, r}]

				for _, p := range g.succ[r] {
					for _, m := range g.succ[p] {
						if m == n {
							continue
						}

						beta := g.Beta(m)
						prod *= beta
					}
				}
			}
		}

		sum += prod
	}

	g.pAlpha[n] = sum

	return sum
}

func (g *Graph) Beta(n *Node) float64 {
	if b, ok := g.pBeta[n]; ok {
		return b
	}

	if len(n.tree.Children) == 0 {
		return 0.5 // TODO real weights
	}

	sumI := float64(0)
	sumR := float64(0)
	sumP := float64(0)

	for _, i := range g.succ[n] {
		w := g.edges[[2]*Node{n, i}]
		sumI += w

		for _, r := range g.succ[g.succ[n][0]] {
			sumR += w * g.edges[[2]*Node{g.succ[n][0], r}] // TODO verify

			for _, p := range g.succ[r] {
				prod := float64(1)

				for _, m := range g.succ[p] {
					prod *= g.Beta(m)
				}

				sumP += prod
			}
		}
	}

	b := sumI * sumR * sumP

	if b > 1 {
		fmt.Println("pBeta > 1")

		b = 1 // TODO fix
	}

	g.pBeta[n] = b

	return b
}
