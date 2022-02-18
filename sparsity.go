package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"strings"
)

const UnknownToken = "$X$"

var tokenOccurrences map[string]int

func initTokenOccurrences() {
	tokenOccurrences = countTokenOccurrences()
}

func countTokenOccurrences() map[string]int {
	fmt.Println("Counting token occurrences...")

	initCorpus()

	occurrences := make(map[string]int)

	count := func(text string) {
		tokens := strings.Split(text, " ")

		for _, t := range tokens {
			occurrences[t]++
		}
	}

	counted, limit := 0, Config.TrainingSampleLimit

	for corpus.Next() && (limit == -1 || counted < limit) {
		sample := corpus.Sample()

		count(sample.Sentence)

		counted++
	}

	return occurrences
}

func replaceSparseLabels(leaves []*tree.Tree, occurrences map[string]int) {
	for _, leaf := range leaves {
		if occurrences[leaf.Label] > 1 {
			continue
		}

		leaf.Label = UnknownToken
	}
}

func replaceSparseTokens(tokens []string, occurrences map[string]int) {
	for i, token := range tokens {
		if occurrences[token] > 1 {
			continue
		}

		tokens[i] = UnknownToken
	}
}
