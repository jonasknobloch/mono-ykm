package main

import (
	"math"
	"math/big"
	"strings"
)

func DecomposeTranslationCount(count *Count) {
	for feature, keys := range count.val {
		for key, val := range keys {
			target := strings.Split(key, " ")

			if len(target) == 1 {
				continue
			}

			var p *big.Float

			if f, _ := val.Float64(); !math.IsInf(f, 0) {
				p = big.NewFloat(math.Pow(f, 1/float64(len(target))))
			} else {
				p = new(big.Float).Copy(val) // TODO estimate x^(1/n)
			}

			for _, token := range target {
				count.Add(feature, token, p)
			}

			count.rwm.Lock()

			delete(count.val[feature], key)

			if len(count.val[feature]) == 0 {
				delete(count.val, feature)
			}

			count.rwm.Unlock()
		}
	}
}
