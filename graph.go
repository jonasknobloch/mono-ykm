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
	pAlpha map[*Node]float64
	pBeta  map[*Node]float64
	pruned map[*Node]struct{}

	insertions   map[string][]*Node // feature -> MajorNode
	reorderings  map[string][]*Node // feature -> MajorNode
	translations map[string][]*Node // feature -> MajorNode

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
		pruned: make(map[*Node]struct{}),

		insertions:   make(map[string][]*Node),
		reorderings:  make(map[string][]*Node),
		translations: make(map[string][]*Node),

		major: make(map[*tree.Tree]map[string]*Node),
	}

	g.AddNode(n)
	g.Expand(n, m, mt.meta)

	for _, node := range g.nodes {
		if node.nType != FinalNode {
			continue
		}

		if g.edges[[2]*Node{g.pred[node][0], node}] == 0 {
			g.Prune(node)
		}
	}

	g.Beta(n)

	for _, node := range g.nodes {
		if node.nType != FinalNode {
			continue
		}

		if _, ok := g.pruned[node]; ok {
			continue
		}

		g.Alpha(g.pred[g.pred[node][0]][0])
	}

	return g
}

func (g *Graph) AddNode(n *Node) {
	g.nodes = append(g.nodes, n)

	if len(g.nodes) > 0 && len(g.nodes)%1000000 == 0 {
		fmt.Printf("%d\n", len(g.nodes))
	}
}

func (g *Graph) AddEdge(n1, n2 *Node, w float64) {
	g.edges[[2]*Node{n1, n2}] = w

	if _, ok := g.pred[n2]; !ok {
		g.pred[n2] = make([]*Node, 0)
	}

	if _, ok := g.succ[n1]; !ok {
		g.succ[n1] = make([]*Node, 0)
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

func (g *Graph) Expand(n *Node, m *Model, fm map[*tree.Tree][3]string) {
	feats, ok := fm[n.tree]

	if !ok {
		panic("unknown feature")
	}

	if feats[0] != "" {
		if _, ok := g.insertions[feats[0]]; !ok {
			g.insertions[feats[0]] = make([]*Node, 0)
		}

		g.insertions[feats[0]] = append(g.insertions[feats[0]], n)
	}

	if feats[1] != "" && len(n.tree.Children) > 0 {
		if _, ok := g.reorderings[feats[1]]; !ok {
			g.reorderings[feats[1]] = make([]*Node, 0)
		}

		g.reorderings[feats[1]] = append(g.reorderings[feats[1]], n)
	}

	if feats[2] != "" && len(n.tree.Children) == 0 {
		if _, ok := g.translations[feats[2]]; !ok {
			g.translations[feats[2]] = make([]*Node, 0)
		}

		g.translations[feats[2]] = append(g.translations[feats[2]], n)
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

	singleton := func(set []*Node) *Node {
		if len(set) != 1 {
			panic("unexpected set length")
		}

		return set[0]
	}

	sum := float64(0)

	for _, partitioning := range g.Predecessor(n) {
		prod := float64(1)

		reordering := singleton(g.Predecessor(partitioning))
		insertion := singleton(g.Predecessor(reordering))
		major := singleton(g.Predecessor(insertion))

		prod *= g.Alpha(major)

		prod *= g.edges[[2]*Node{major, insertion}]
		prod *= g.edges[[2]*Node{insertion, reordering}]

		for _, sibling := range g.Successor(partitioning) {
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

func (g *Graph) Successor(n *Node) []*Node {
	succ := make([]*Node, 0)

	for _, s := range g.succ[n] {
		if _, ok := g.pruned[s]; ok {
			continue
		}

		succ = append(succ, s)
	}

	return succ
}

func (g *Graph) Predecessor(n *Node) []*Node {
	pred := make([]*Node, 0)

	for _, p := range g.pred[n] {
		if _, ok := g.pruned[p]; ok {
			continue
		}

		pred = append(pred, p)
	}

	return pred
}

func (g *Graph) Prune(n *Node) {
	g.pruned[n] = struct{}{}

	if n.nType == MajorNode {
		for _, p := range g.Predecessor(n) {
			g.Prune(p)
		}

		return
	}

	for _, p := range g.Predecessor(n) {
		if len(g.Successor(p)) > 0 {
			continue
		}

		g.Prune(p)
	}
}

func (g *Graph) removePruned(nodes []*Node) []*Node {
	result := make([]*Node, 0)

	// TODO modify input slice

	for _, n := range nodes {
		if _, ok := g.pruned[n]; ok {
			continue
		}

		result = append(result, n)
	}

	return result
}

func (g *Graph) InsertionCandidateNodes(feature string) []*Node {
	return g.removePruned(g.insertions[feature])
}

func (g *Graph) ReorderingCandidateNodes(feature string) []*Node {
	return g.removePruned(g.reorderings[feature])
}

func (g *Graph) TranslationCandidateNode(feature string) []*Node {
	return g.removePruned(g.translations[feature])
}
