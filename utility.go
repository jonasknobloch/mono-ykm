package main

import (
	"errors"
	"math"
	"math/big"
)

func Root(f *big.Float, n int) (big.Accuracy, error) {
	f64, a := f.Float64()

	if f64 == 0 {
		return a, errors.New("conversion underflow")
	}

	if math.IsInf(f64, 0) {
		return a, errors.New("conversion overflow")
	}

	f.SetFloat64(math.Pow(f64, 1/float64(n)))

	return a, nil
}
