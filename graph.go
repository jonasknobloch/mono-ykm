package main

import (
	"errors"
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

func NewGraph(mt *MetaTree, f []string, m *Model) (*Graph, error) {
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

	var err error

	if !g.nodes[0].valid {
		err = errors.New("invalid root node")
	}

	return g, err
}

func (g *Graph) AddNode(n *Node) {
	if !n.valid && len(g.nodes) != 0 {
		panic("invalid node")
	}

	g.nodes = append(g.nodes, n)
}

func (g *Graph) AddEdge(n1, n2 *Node, w *big.Float) {
	if !n1.valid || !n2.valid {
		panic("invalid edge")
	}

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

func reachable(t *tree.Tree, l int) bool {
	size := t.Size()
	leaves := t.Leaves()

	max := 0

	if Config.EnableInteriorInsertions {
		max += size - len(leaves)
	}

	if Config.EnableTerminalInsertions {
		max += len(leaves)
	}

	if !Config.EnablePhrasalTranslations {
		max += len(leaves)
	} else if len(t.Children) == 0 {
		max += len(leaves)
	} else {
		max += len(leaves) * Config.PhraseLengthLimit // assumes each leaf has a preceding POS tag
	}

	return l <= max
}

func partitionings(reordering *Node) [][]int {
	validate := func(p, i int) bool {
		return reachable(reordering.tree.Children[reordering.r.Reordering[i]], p)
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

		phrasal := len(n.tree.Children) != 0 && Config.EnablePhrasalTranslations

		if phrasal && phrasalFrequencies != nil {
			frequency, ok := phrasalFrequencies[eStr][i.Substring()]
			phrasal = ok && frequency >= Config.PhraseFrequencyCutoff
		}

		if (len(n.tree.Children) == 0 && i.l < 2) || phrasal {
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

			f.valid = true
			i.valid = true
			n.valid = true

			g.AddNode(f)
			g.AddEdge(i, f, m.Probability(translation))
			g.AddOperation(translation, n)
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
					valid: true,
				}

				k := r.k

				nodes := make(map[*Node]struct{}) // new major nodes
				edges := make([][2]*Node, 0)      // new p -> major edges

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

						nodes[major] = struct{}{}

						g.Expand(major, m, mt)
					}

					p.valid = p.valid && g.major[c][sub].valid

					if !p.valid {
						break
					}

					edges = append(edges, [2]*Node{p, g.major[c][sub]})
				}

				if p.valid {
					for node := range nodes {
						g.AddNode(node)
					}

					for _, edge := range edges {
						g.AddEdge(edge[0], edge[1], big.NewFloat(1))
					}
				}

				r.valid = r.valid || p.valid
				i.valid = i.valid || r.valid
				n.valid = n.valid || i.valid

				if r.valid && p.valid {
					g.AddNode(p)
					g.AddEdge(r, p, big.NewFloat(1))
				}
			}

			if i.valid && r.valid {
				g.AddNode(r)
				g.AddEdge(i, r, m.Probability(reordering))
				g.AddOperation(reordering, n)
			}
		}

		if n.valid && i.valid {
			g.AddNode(i)
			g.AddEdge(n, i, m.Probability(insertion))
			g.AddOperation(insertion, n)
		}
	}

	if n.valid && len(n.tree.Children) != 0 {
		n.lambda, n.kappa = m.Lambda(eStr)

		g.TrackNode(g.lambda, eStr, LambdaKey, n)
		g.TrackNode(g.lambda, eStr, KappaKey, n)
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

		prod.Mul(prod, g.Alpha(major))

		prod.Mul(prod, g.edges[[2]*Node{major, insertion}])

		rProb := new(big.Float)
		tProb := new(big.Float)

		rProb.Copy(g.edges[[2]*Node{insertion, reordering}])

		for _, translation := range g.succ[insertion] {
			if translation.nType == FinalNode {
				tProb.Copy(g.edges[[2]*Node{insertion, translation}])
				break
			}
		}

		tProb.Mul(tProb, major.lambda)
		rProb.Mul(rProb, major.kappa)

		prod.Mul(prod, rProb.Add(rProb, tProb))

		for _, sibling := range g.succ[partitioning] {
			if sibling == n {
				continue
			}

			prod.Mul(prod, g.Beta(sibling))
		}

		sum.Add(sum, prod)
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
