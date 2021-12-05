package main

import (
	"fmt"
	"testing"
)

func TestPartitionings(t *testing.T) {
	e := []string{"I", "do", "like", "dogs", "."}

	for _, p := range partitionings(5, 3) {

		k := 0

		for _, l := range p {
			fmt.Printf(" %v ", e[k:k+l])
			k = l
		}

		fmt.Printf("\n")
	}

	// TODO implement
}
