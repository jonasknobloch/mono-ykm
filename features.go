package main

import "github.com/jonasknobloch/jinn/pkg/tree"

func nFeature(p, st *tree.Tree, replaceUnknownTokens bool) string {
	feature := func(p, st *tree.Tree) string {
		var f string

		if p == nil {
			f = "ROOT " + st.Label
		} else {
			f = p.Label + " " + st.Label
		}

		return f
	}

	f := feature(p, st)

	if f == "" || !replaceUnknownTokens {
		return f
	}

	if _, ok := model.n[f]; !ok {
		if len(st.Children) == 0 {
			st.Label = UnknownToken

			f = feature(p, st)
		}
	}

	return f
}

func rFeature(st *tree.Tree, replaceUnknownTokens bool) string {
	feature := func(st *tree.Tree) string {
		if len(st.Children) == 0 {
			return ""
		}

		var f string

		for _, c := range st.Children {
			f += " " + c.Label
		}

		return f[1:]
	}

	f := feature(st)

	if f == "" || !replaceUnknownTokens {
		return f
	}

	if _, ok := model.r[f]; !ok {
		for _, c := range st.Children {
			if len(st.Children) == 0 {
				c.Label = UnknownToken
			}
		}

		f = feature(st)
	}

	return f
}

func tFeature(st *tree.Tree, replaceUnknownTokens bool) string {
	feature := func(st *tree.Tree) string {
		if len(st.Children) != 0 {
			return ""
		}

		return st.Label
	}

	f := feature(st)

	if f == "" || !replaceUnknownTokens {
		return f
	}

	if _, ok := model.t[f]; !ok {
		st.Label = UnknownToken

		f = feature(st)
	}

	return f
}
