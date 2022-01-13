package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

func Import(name string) (map[string]map[string]float64, error) {
	f, err := os.Open(name)

	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	defer f.Close()

	r := csv.NewReader(f)

	r.Comma = '\t'

	var line int
	var keys []string

	p := make(map[string]map[string]float64)

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		line++

		if line == 1 {
			keys = record
			continue
		}

		var feature string

		for i, v := range record {
			if i == 0 {
				feature = v
				continue
			}

			if v == "" {
				continue
			}

			if _, ok := p[feature]; !ok {
				p[feature] = make(map[string]float64, len(keys))
			}

			if value, err := strconv.ParseFloat(v, 64); err == nil {
				p[feature][keys[i]] = value
			} else {
				return nil, fmt.Errorf("error converting value: %w", err)
			}
		}
	}

	return p, nil
}
