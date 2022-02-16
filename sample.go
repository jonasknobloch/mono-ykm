package main

import (
	"errors"
)

type Sample struct {
	ID       string
	Tree     string
	Sentence string
	Label    bool
}

func NewSample(record []string) (*Sample, error) {
	out := &Sample{}

	out.ID = record[0]
	out.Tree = record[1]
	out.Sentence = record[2]

	if record[3] != "0" && record[3] != "1" {
		return nil, errors.New("unknown quality")
	}

	out.Label = record[3] == "1"

	return out, nil
}
