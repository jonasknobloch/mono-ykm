package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
)

type Graph struct {
	root   *Node
	nodes  []*Node
	edges  map[[2]*Node]float64
	pred   map[*Node][]*Node
	succ   map[*Node][]*Node
	pAlpha map[*Node]float64
	pBeta  map[*Node]float64

	insertions   map[string]map[string][]*Node
	reorderings  map[string]map[string][]*Node
	translations map[string]map[string][]*Node

	major map[*tree.Tree]map[string]*Node
}

func NewGraph(mt *MetaTree, f []string, m *Model) *Graph {
	n := &Node{tree: mt.Tree, f: f, k: 0, l: len(f), nType: MajorNode}

	g := &Graph{
		root:   n,
		nodes:  make([]*Node, 0),
		edges:  make(map[[2]*Node]float64),
		pred:   make(map[*Node][]*Node),
		succ:   make(map[*Node][]*Node),
		pAlpha: make(map[*Node]float64),
		pBeta:  make(map[*Node]float64),

		insertions:   make(map[string]map[string][]*Node),
		reorderings:  make(map[string]map[string][]*Node),
		translations: make(map[string]map[string][]*Node),

		major: make(map[*tree.Tree]map[string]*Node),
	}

	g.AddNode(n)
	g.Expand(n, m, mt.meta)
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

	if _, ok := g.pred[n2]; !ok {
		if n2.nType == MajorNode {
			g.pred[n2] = make([]*Node, 0)
		} else {
			g.pred[n2] = make([]*Node, 0, 1)
		}
	}

	if _, ok := g.succ[n1]; !ok {
		g.succ[n1] = make([]*Node, 0)
	}

	g.pred[n2] = append(g.pred[n2], n1)
	g.succ[n1] = append(g.succ[n1], n2)
}

func (g *Graph) AddOperation(op Operation, n *Node) {
	var m map[string]map[string][]*Node

	switch op.(type) {
	case Insertion:
		m = g.insertions
	case Reordering:
		m = g.reorderings
	case Translation:
		m = g.translations
	default:
		panic("unexpected operation type")
	}

	feature, key := op.Feature(), op.Key()

	if _, ok := m[feature]; !ok {
		m[feature] = make(map[string][]*Node)
	}

	if _, ok := m[feature][key]; !ok {
		m[feature][key] = make([]*Node, 0)
	}

	m[feature][key] = append(m[feature][key], n)
}

func partitionings(reordering *Node) [][]int {
	validate := func(p, i int) bool {
		c := reordering.tree.Children[reordering.r.Reordering[i]]

		if Config.AllowTerminalInsertions && p > c.Size()+1 {
			return false
		}

		if !Config.AllowTerminalInsertions && p > c.Size() {
			return false
		}

		return true
	}

	var p func(n, k int, r [][]int) [][]int
	p = func(n, k int, r [][]int) [][]int {
		if k == 1 {
			if validate(n, 0) {
				r = append(r, []int{n})
			}

			return r
		}

		for i := n; i >= 0; i-- {
			for _, sp := range p(n-i, k-1, make([][]int, 0)) {
				if !validate(i, len(sp)) {
					continue
				}

				r = append(r, append(sp, i))
			}
		}

		return r
	}

	return p(reordering.l, len(reordering.tree.Children), make([][]int, 0))
}

func (g *Graph) Expand(n *Node, m *Model, fm map[*tree.Tree][3]string) {
	feats, ok := fm[n.tree]

	if !ok {
		panic("unknown feature")
	}

	for _, op := range Insertions(n.tree, n.f[n.k:n.k+n.l], feats[InsertionFeature], false) {
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

		g.AddNode(i)
		g.AddEdge(n, i, m.PInsertion(insertion))
		g.AddOperation(insertion, n)

		if len(n.tree.Children) == 0 {
			translation := NewTranslation(i.Substring(), feats[TranslationFeature])

			f := &Node{
				n:     i.n,
				t:     translation,
				tree:  i.tree,
				f:     i.f,
				k:     i.k,
				l:     i.l,
				nType: FinalNode,
			}

			g.AddNode(f)
			g.AddEdge(i, f, m.PTranslation(translation))
			g.AddOperation(translation, n)

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
			g.AddEdge(i, r, m.PReordering(reordering))
			g.AddOperation(reordering, n)

			for _, partitioning := range partitionings(r) {
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

				k := r.k

				for i := 0; i < len(r.tree.Children); i++ {
					c := r.tree.Children[p.r.Reordering[i]]

					if _, ok := g.major[c]; !ok {
						g.major[c] = make(map[string]*Node)
					}

					major := &Node{
						tree:  c,
						f:     p.f,
						k:     k,
						l:     partitioning[i],
						nType: MajorNode,
					}

					k += partitioning[i]

					sub := major.Substring()

					if _, ok := g.major[c][sub]; !ok {
						g.major[c][sub] = major

						g.AddNode(major)
						g.Expand(major, m, fm)
					}

					g.AddEdge(p, g.major[c][sub], 1)
				}
			}
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

	for _, partitioning := range g.pred[n] {
		prod := float64(1)

		reordering := g.pred[partitioning][0]
		insertion := g.pred[reordering][0]
		major := g.pred[insertion][0]

		prod *= g.Alpha(major)

		prod *= g.edges[[2]*Node{major, insertion}]
		prod *= g.edges[[2]*Node{insertion, reordering}]

		for _, sibling := range g.succ[partitioning] {
			if sibling == n {
				continue
			}

			prod *= g.Beta(sibling)
		}

		sum += prod
	}

	g.pAlpha[n] = sum

	return g.pAlpha[n]
}

func (g *Graph) Beta(n *Node) float64 {
	if b, ok := g.pBeta[n]; ok {
		return b
	}

	if len(n.tree.Children) == 0 {
		g.pBeta[n] = g.InsideWeightsTerminal(n)
	} else {
		g.pBeta[n] = g.InsideWeightsInterior(n)
	}

	return g.pBeta[n]
}
