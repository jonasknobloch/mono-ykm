package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Export(t map[string]map[string]float64, stubs ...string) error {
	if len(t) == 0 {
		return nil
	}

	name := fmt.Sprintf("model_%s.gob", strings.Join(stubs, "-"))
	file, err := os.Create(filepath.Join(Config.ModelExportDirectory, name))

	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	enc := gob.NewEncoder(file)

	if err := enc.Encode(t); err != nil {
		return fmt.Errorf("error encoding table: %w", err)
	}

	return nil
}
