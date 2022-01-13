package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Export(p map[string]map[string]float64, stubs ...string) error {
	if len(p) == 0 {
		return nil
	}

	name := fmt.Sprintf("model_%s.tsv", strings.Join(stubs, "-"))

	f, err := os.Create(filepath.Join(Config.ModelExportDirectory, name))

	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer f.Close()

	w := csv.NewWriter(f)

	w.Comma = '\t'

	features := make([]string, 0)
	keys := make([]string, 0)

	seen := make(map[string]struct{})

	for feature := range p {
		features = append(features, feature)

		for key := range p[feature] {
			if _, ok := seen[key]; ok {
				continue
			}

			keys = append(keys, key)
			seen[key] = struct{}{}
		}
	}

	sort.Strings(features)
	sort.Strings(keys)

	records := make([][]string, 0)

	records = append(records, append([]string{"X"}, keys...))

	for _, feature := range features {
		record := []string{feature}

		for _, key := range keys {
			if value, ok := p[feature][key]; ok {
				record = append(record, strconv.FormatFloat(value, 'E', -1, 64))
			} else {
				record = append(record, "")
			}
		}

		records = append(records, record)
	}

	if err := w.WriteAll(records); err != nil {
		return fmt.Errorf("error writing csv: %w", err)
	}

	return nil
}
