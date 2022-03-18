package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"github.com/jonasknobloch/jinn/pkg/utility"
	"strings"
)

var phrasalFrequencies map[string]map[string]int

func initPhrasalFrequencies() {
	phrasalFrequencies = countPhrasalPairs()
}

func countPhrasalPairs() map[string]map[string]int {
	fmt.Println("Counting phrasal pairs...")

	initCorpus()

	pairs := make(map[string]map[string]int)

	valid := func(es, et []string) bool {
		if len(es) < 2 && len(et) < 2 {
			return false
		}

		if len(es) < 2 && !Config.EnableInterior1ToNTranslations {
			return false
		}

		if len(et) < 2 && !Config.EnableInteriorNTo1Translations {
			return false
		}

		if len(es) > Config.PhraseLengthLimit || len(et) > Config.PhraseLengthLimit {
			return false
		}

		return true
	}

	add := func(feature, key string) {
		if _, ok := pairs[feature]; !ok {
			pairs[feature] = make(map[string]int)
		}

		pairs[feature][key]++
	}

	counted, limit := 0, Config.TrainingSampleLimit

	for corpus.Next() && (limit == -1 || counted < limit) {
		sample := corpus.Sample()

		mt, e, err := initSample(sample)

		if err != nil {
			continue
		}

		mt.Tree.Walk(func(st *tree.Tree) {
			if len(st.Children) == 0 {
				return
			}

			source := st.Sentence()
			sourceTokens := strings.Split(source, " ")

			if valid(sourceTokens, []string{}) {
				add(source, "")
			}

			for i := 1; i <= Config.PhraseLengthLimit; i++ {
				for _, ngram := range utility.NGrams(e, i, nil) {
					if !valid(sourceTokens, ngram) {
						continue
					}

					add(source, strings.Join(ngram, " "))
				}
			}

		})

		counted++
	}

	return pairs
}
