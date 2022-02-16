package main

import (
	"encoding/gob"
	"fmt"
	"os"
)

func Import(name string) (map[string]map[string]float64, error) {
	file, err := os.Open(name)

	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	t := make(map[string]map[string]float64)

	dec := gob.NewDecoder(file)

	if err := dec.Decode(&t); err != nil {
		return nil, fmt.Errorf("error decoding file: %w", err)
	}

	return t, nil
}
