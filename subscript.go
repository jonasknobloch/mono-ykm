package main

// TODO use gonum implementation where possible

type SubscriptGenerator struct {
	dim  []int
	nSub int
	idx  int
	sub  []int
}

func NumSub(dim []int) int {
	nSub := 1

	for _, d := range dim {
		nSub = nSub * d
	}

	return nSub
}

func NewSubscriptGenerator(b []int) *SubscriptGenerator {
	return &SubscriptGenerator{
		dim:  b,
		nSub: NumSub(b),
		idx:  -1,
		sub:  make([]int, len(b)),
	}
}

func (g *SubscriptGenerator) Next() bool {
	if g.idx < 0 {
		g.idx++
		return true
	}

	i := len(g.sub) - 1

	if g.sub[i] < g.dim[i]-1 {
		g.sub[i]++
		g.idx++
		return true
	}

	for j := i; j > -1; j-- {
		if g.sub[j] < g.dim[j]-1 {
			g.sub[j]++
			g.idx++
			return true
		}

		g.sub[j] = 0
	}

	return false
}

func (g *SubscriptGenerator) Subscript() []int {
	if g.idx < 0 {
		panic("Subscript called before Next")
	}

	return g.sub
}

func IdxFor(sub, dims []int) (idx int) {
	m := NumSub(dims)

	for i, d := range dims {
		m /= d

		idx += sub[i] * m
	}

	return
}

func SubFor(sub []int, idx int, dims []int) {
	m := NumSub(dims)

	for i, d := range dims {
		m /= d

		for idx >= m {
			idx -= m
			sub[i]++
		}
	}
}
