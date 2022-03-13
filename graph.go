package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"math/big"
	"strings"
)

type Graph struct {
	nodes []*Node
	edges map[[2]*Node]*big.Float
	pred  map[*Node][]*Node
	succ  map[*Node][]*Node

	pAlpha map[*Node]*big.Float
	pBeta  map[*Node]*big.Float

	insertions   map[string]map[string][]*Node
	reorderings  map[string]map[string][]*Node
	translations map[string]map[string][]*Node

	lambda map[string]map[string][]*Node

	major map[*tree.Tree]map[string]*Node
}

const LambdaKey = "l"
const KappaKey = "k"

func NewGraph(mt *MetaTree, f []string, m *Model) *Graph {
	n := &Node{
		tree:  mt.Tree,
		f:     f,
		k:     0,
		l:     len(f),
		nType: MajorNode,
	}

	g := &Graph{
		nodes: make([]*Node, 0),
		edges: make(map[[2]*Node]*big.Float),
		pred:  make(map[*Node][]*Node),
		succ:  make(map[*Node][]*Node),

		pAlpha: make(map[*Node]*big.Float),
		pBeta:  make(map[*Node]*big.Float),

		insertions:   make(map[string]map[string][]*Node),
		reorderings:  make(map[string]map[string][]*Node),
		translations: make(map[string]map[string][]*Node),

		lambda: make(map[string]map[string][]*Node),

		major: make(map[*tree.Tree]map[string]*Node),
	}

	g.AddNode(n)
	g.Expand(n, m, mt)
	g.Beta(n)

	for _, node := range g.nodes {
		if node.nType != MajorNode {
			continue
		}

		g.Alpha(node)
	}

	return g
}

func (g *Graph) AddNode(n *Node) {
	g.nodes = append(g.nodes, n)
}

func (g *Graph) AddEdge(n1, n2 *Node, w *big.Float) {
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

func (g *Graph) TrackNode(m map[string]map[string][]*Node, feature, key string, n *Node) {
	if _, ok := m[feature]; !ok {
		m[feature] = make(map[string][]*Node)
	}

	if _, ok := m[feature][key]; !ok {
		m[feature][key] = make([]*Node, 0)
	}

	m[feature][key] = append(m[feature][key], n)
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

	g.TrackNode(m, feature, key, n)
}

func partitionings(reordering *Node) [][]int {
	validate := func(p, i int) bool {
		st := reordering.tree.Children[reordering.r.Reordering[i]]

		return !(p > st.Size()+(len(st.Leaves())*Config.PhraseLengthLimit))
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

func (g *Graph) Expand(n *Node, m *Model, mt *MetaTree) {
	e := func() []string {
		leaves := n.tree.Leaves()
		labels := make([]string, len(leaves))

		for i, leaf := range leaves {
			labels[i] = leaf.Label
		}

		return labels
	}()

	eStr := strings.Join(e, " ")

	for _, op := range Insertions(n.tree, n.f[n.k:n.k+n.l], mt.Feature(n.tree, InsertionFeature)) {
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
		g.AddEdge(n, i, m.Probability(insertion))
		g.AddOperation(insertion, n)

		phrasal := true

		phrasal = phrasal && len(e) > 1
		phrasal = phrasal && l > 1

		phrasal = phrasal && len(e) <= Config.PhraseLengthLimit
		phrasal = phrasal && l <= Config.PhraseLengthLimit

		// TODO phrase frequency cutoff

		if len(n.tree.Children) == 0 || phrasal {
			translation := NewTranslation(i.Substring(), mt.Feature(n.tree, TranslationFeature))

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
			g.AddEdge(i, f, m.Probability(translation))
			g.AddOperation(translation, n)

			if len(n.tree.Children) == 0 {
				continue
			}

			if ms, ok := g.lambda[eStr][LambdaKey]; ok {
				if ms[len(ms)-1] == n {
					continue
				}
			}

			if len(n.tree.Children) != 0 {
				n.lambda, n.kappa = m.Lambda(eStr)

				g.TrackNode(g.lambda, eStr, LambdaKey, n)
				g.TrackNode(g.lambda, eStr, KappaKey, n)
			}
		}

		for _, op := range Reorderings(n.tree, mt.Feature(n.tree, ReorderingFeature)) {
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
			g.AddEdge(i, r, m.Probability(reordering))
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
				g.AddEdge(r, p, big.NewFloat(1))

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
						g.Expand(major, m, mt)
					}

					g.AddEdge(p, g.major[c][sub], big.NewFloat(1))
				}
			}
		}
	}
}

func (g *Graph) Alpha(n *Node) *big.Float {
	if a, ok := g.pAlpha[n]; ok {
		return a
	}

	if n == g.nodes[0] {
		g.pAlpha[n] = big.NewFloat(1)

		return g.pAlpha[n]
	}

	sum := new(big.Float)

	for _, partitioning := range g.pred[n] {
		prod := big.NewFloat(1)

		reordering := g.pred[partitioning][0]
		insertion := g.pred[reordering][0]
		major := g.pred[insertion][0]

		nProb := new(big.Float).Copy(g.edges[[2]*Node{major, insertion}])
		rProb := new(big.Float).Copy(g.edges[[2]*Node{insertion, reordering}])

		var translation *Node
		var tProb *big.Float

		for _, t := range g.succ[insertion] {
			if t.nType == FinalNode {
				translation = t
				break
			}
		}

		tProb = new(big.Float)

		if translation != nil {
			tProb.Copy(g.edges[[2]*Node{insertion, translation}])
		}

		for _, sibling := range g.succ[partitioning] {
			if sibling == n {
				continue
			}

			rProb.Mul(prod, g.Beta(sibling))
		}

		prod.Mul(prod, g.Alpha(major))

		if major.lambda != nil && major.kappa != nil {
			tProb.Mul(tProb, major.lambda).Mul(tProb, nProb)
			rProb.Mul(rProb, major.kappa).Mul(rProb, nProb)
		}

		sum.Add(sum, tProb)
		sum.Add(sum, rProb)
	}

	g.pAlpha[n] = sum

	return g.pAlpha[n]
}

func (g *Graph) Beta(n *Node) *big.Float {
	if b, ok := g.pBeta[n]; ok {
		return b
	}

	g.pBeta[n] = g.InsideWeight(n, [3]string{}, nil, nil)

	return g.pBeta[n]
}
