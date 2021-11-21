package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"gonum.org/v1/gonum/stat/combin"
	"strconv"
	"strings"
)

type NP map[string]map[string]map[int]map[string]float64
type RP map[string]map[string]float64
type TP map[string]map[string]float64

var labels []string

func init() {
	labels = []string{
		"S",
		"SBAR",
		"SBARQ",
		"SINV",
		"SQ",
		"ADJP",
		"ADVP",
		"CONJP",
		"FRAG",
		"INTJ",
		"LST",
		"NAC",
		"NP",
		"NX",
		"PP",
		"PRN",
		"QP",
		"RRC",
		"UCP",
		"VP",
		"WHADJP",
		"WHADVP",
		"WHNP",
		"WHPP",
		"X",
	}

	// TODO add POS tags
}

func InitRP(zeroInit bool) RP {
	join := func(p []int, l bool) string {
		if len(p) == 0 {
			panic("invalid input")
		}

		sb := strings.Builder{}

		for _, d := range p {
			sb.WriteString(" ")

			if l {
				sb.WriteString(labels[d])
			} else {
				sb.WriteString(strconv.Itoa(d))
			}
		}

		return sb.String()[1:]
	}

	rp := make(map[string]map[string]float64)

	for k := 2; k < 5; k++ {
		pg := combin.NewPermutationGenerator(len(labels), 3)

		for pg.Next() {
			p := pg.Permutation(nil)

			f := join(p, true)

			nP := combin.NumPermutations(k, k)
			rp[f] = make(map[string]float64, nP)

			for _, q := range combin.Permutations(k, k) {
				if zeroInit {
					rp[f][join(q, false)] = 0
					continue
				}

				rp[f][join(q, false)] = 1 / float64(nP)
			}
		}
	}

	return rp
}

func InitNP(dict []string, zeroInit bool) NP {
	labels := append([]string{"ROOT"}, labels...)
	np := make(map[string]map[string]map[int]map[string]float64, len(labels))

	for _, l1 := range labels {
		np[l1] = make(map[string]map[int]map[string]float64, len(labels))

		for _, l2 := range labels {
			np[l1][l2] = make(map[int]map[string]float64, 3)

			if zeroInit {
				np[l1][l2][0] = map[string]float64{"": 0}
			} else {
				np[l1][l2][0] = map[string]float64{"": 1 / 3.0}
			}

			for i := 1; i < 3; i++ {
				np[l1][l2][i] = make(map[string]float64, len(dict))

				for _, w := range dict {
					if zeroInit {
						np[l1][l2][i][w] = 0
						continue
					}

					np[l1][l2][i][w] = 1 / 3.0 / float64(len(dict))
				}
			}

		}
	}

	return np
}

func InitTP(dict []string, zeroInit bool) TP {
	tp := make(map[string]map[string]float64, len(dict))

	for _, w1 := range dict {
		if w1 == "" {
			continue
		}

		tp[w1] = make(map[string]float64, len(dict)-1)

		for _, w2 := range dict {
			if zeroInit {
				tp[w1][w2] = 0
				continue
			}

			tp[w1][w2] = 1 / float64(len(dict))
		}
	}

	return tp
}

func Train(d Corpus, nd, td []string) {
	np := InitNP(nd, false)
	rp := InitRP(false)
	tp := InitTP(td, false)

	var cn NP
	var cr RP
	var ct TP

	join := func(p []int) string {
		if len(p) == 0 {
			panic("invalid input")
		}

		sb := strings.Builder{}

		for _, d := range p {
			sb.WriteString(" ")
			sb.WriteString(strconv.Itoa(d))
		}

		return sb.String()[1:]
	}

	pn := func(tr *tree.Tree, seq sequence, np NP, cb func(p float64) float64) {
		f := seq.f[tr].n
		o := seq.s[tr].n

		// TODO to error cases
		if _, ok := np[f[0]][f[1]]; !ok {
			return
		}

		if o[:1] == "n" {
			np[f[0]][f[1]][0][""] = cb(np[f[0]][f[1]][0][""])
		}

		if seq.s[tr].n[:1] == "l" {
			np[f[0]][f[1]][1][o[1:]] = cb(np[f[0]][f[1]][1][o[1:]])
		}

		if seq.s[tr].n[:1] == "r" {
			np[f[0]][f[1]][2][o[1:]] = cb(np[f[0]][f[1]][2][o[1:]])
		}
	}

	pr := func(tr *tree.Tree, seq sequence, rp RP, cb func(p float64) float64) {
		f := seq.f[tr].r
		o := seq.s[tr].r

		if f == "" {
			return
		}

		// TODO to error cases
		if pp, ok := rp[f][join(o)]; ok {
			rp[f][join(o)] = cb(pp)
		}
	}

	pt := func(tr *tree.Tree, seq sequence, tp TP, cb func(p float64) float64) {
		f := seq.f[tr].t
		o := seq.s[tr].t

		if f == "" {
			return
		}

		// TODO to error cases
		if pp, ok := tp[f][o]; ok {
			tp[f][o] = cb(pp)
		}
	}

	P := func(tr *tree.Tree, seq sequence) float64 {
		p := float64(1)

		tr.Walk(func(st *tree.Tree) {
			mul := func(pp float64) float64 {
				p *= pp
				return pp
			}

			pn(st, seq, np, mul)
			pr(st, seq, rp, mul)
			pt(st, seq, tp, mul)
		})

		return p
	}

	for i := 0; i < 3; i++ {
		cn = InitNP(nd, true)
		cr = InitRP(true)
		ct = InitTP(td, true)

		fmt.Printf("count N: %d", len(cn))
		fmt.Printf(" count R: %d", len(cr))
		fmt.Printf(" count T: %d\n", len(ct))

		for _, s := range d {
			Z := float64(0)

			// TODO streamline insertion feature

			nd2 := []string{"n"}

			for _, w := range nd {
				nd2 = append(nd2, "l"+w)
				nd2 = append(nd2, "r"+w)
			}

			seq := sequences(s.t, s.e, td, nd2)

			for _, ss := range seq {
				Z += P(s.t, ss)
			}

			fmt.Printf("Z: %f\n", Z)

			for _, ss := range seq {
				c := P(s.t, ss) / Z

				s.t.Walk(func(st *tree.Tree) {
					add := func(p float64) float64 {
						return p + c
					}

					pn(st, ss, cn, add)
					pr(st, ss, cr, add)
					pt(st, ss, ct, add)
				})
			}
		}

		for k1, v1 := range cn {
			for k2, v2 := range v1 {
				var sum float64

				for _, v3 := range v2 {
					for _, v4 := range v3 {
						sum += v4
					}
				}

				for k3, v3 := range v2 {
					for k4, v4 := range v3 {
						np[k1][k2][k3][k4] = v4 / sum
					}
				}
			}
		}

		for k1, v1 := range cr {
			var sum float64

			for _, v2 := range v1 {
				sum += v2
			}

			for k2, v2 := range v1 {
				rp[k1][k2] = v2 / sum
			}
		}

		for k1, v1 := range ct {
			var sum float64

			for _, v2 := range v1 {
				sum += v2
			}

			for k2, v2 := range v1 {
				tp[k1][k2] = v2 / sum
			}
		}
	}
}
