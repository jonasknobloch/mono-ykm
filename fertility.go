package main

import (
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

			p := new(big.Float).Copy(val)

			if _, err := Root(p, len(target)); err != nil {
				// TODO estimate x^(1/n)
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
