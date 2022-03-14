package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"gonum.org/v1/gonum/stat/combin"
	"math"
)

const intSize = 32 << (^int(0) >> 32 & 1)

func O(t *tree.Tree, l int) (int, bool) {
	numSubstrings := func(l, k int) int {
		if k == 0 {
			return 1
		}

		return l - k + 1
	}

	result := 0
	overflow := false

	k := 0
	n := numSubstrings(l, 0)

	for n > 0 {
		t.Walk(func(st *tree.Tree) {
			sum := 1 // major node

			if len(st.Children) == 0 {
				numInsertions := 3
				numTranslations := numInsertions * 1

				sum += numInsertions
				sum += numTranslations
			}

			if len(st.Children) > 0 {
				numInsertions := 3
				numTranslations := 0

				if Config.EnablePhrasalTranslations {
					numTranslations = numInsertions * 1
				}

				numReorderings := numInsertions * combin.NumPermutations(len(st.Children), len(st.Children))

				numPartitionings := 0

				if k == 1 {
					numPartitionings = 1
				}

				if k > 1 {
					for i := k; i <= k+len(st.Children); i++ {
						numPartitionings += combin.Binomial(i-2, k-2)
					}
				}

				numPartitionings *= numReorderings

				sum += numInsertions
				sum += numTranslations
				sum += numReorderings
				sum += numPartitionings
			}

			gain := sum * n

			if intSize == 32 && (result > math.MaxInt32-gain || gain > math.MaxInt32-result) {
				overflow = true
			}

			if intSize == 64 && (result > math.MaxInt64-gain || gain > math.MaxInt64-result) {
				overflow = true
			}

			result += gain
		})

		k++
		n = numSubstrings(l, k)
	}

	return result, !overflow
}
