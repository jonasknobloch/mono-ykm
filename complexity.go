package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"gonum.org/v1/gonum/stat/combin"
)

func O(t *tree.Tree, l int) int {
	numSubstrings := func(l, k int) int {
		if k == 0 {
			return 1
		}

		return l - k + 1
	}

	result := 0

	k := 0
	n := numSubstrings(l, 0)

	for n > 0 {
		t.Walk(func(st *tree.Tree) {
			sum := 1 // major node

			if len(st.Children) == 0 {
				var numInsertions int

				if Config.AllowTerminalInsertions {
					numInsertions = 3
				} else {
					numInsertions = 1
				}

				sum += numInsertions
				sum += numInsertions * 1 // translations
			}

			if len(st.Children) > 0 {
				numInsertions := 3
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
				sum += numReorderings
				sum += numPartitionings
			}

			result += sum * n
		})

		k++
		n = numSubstrings(l, k)
	}

	return result
}
