package main

import (
	"encoding/csv"
	"errors"
	"os"
)

type Iterator struct {
	reader *csv.Reader
	error  error
	sample *Sample
}

func NewIterator(name string) (*Iterator, error) {
	f, err := os.Open(name)

	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)

	r.Comma = '\t'

	var header [4]string

	if record, err := r.Read(); err != nil {
		return nil, err
	} else {
		copy(header[:], record[0:4])
	}

	if header != [4]string{"ID", "Tree", "Sentence", "Label"} {
		return nil, errors.New("unexpected header row")
	}

	return &Iterator{
		reader: r,
		sample: nil,
	}, nil
}

func (i *Iterator) Next() bool {
	record, err := i.reader.Read()

	if err != nil {
		i.error = err

		return false
	}

	i.sample, i.error = NewSample(record)

	if i.error != nil {
		return false
	}

	return true
}

func (i *Iterator) Error() error {
	return i.error
}

func (i *Iterator) Sample() *Sample {
	if i.sample == nil {
		panic("Sample called before Next")
	}

	return i.sample
}
