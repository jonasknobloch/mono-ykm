package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/data"
	"github.com/jonasknobloch/jinn/pkg/data/msrpc"
	"github.com/jonasknobloch/jinn/pkg/data/paws"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var parser *corenlp.Client
var tokenizer *corenlp.Client

type qqpSample struct {
	judgement bool
	question1 string
	question2 string
	pairID    int
}

func init() {
	u, err := url.Parse("http://localhost:9000")

	if err != nil {
		log.Fatalln(err)
	}

	if err := initParser(u); err != nil {
		log.Fatalln(err)
	}

	if err := initTokenizer(u); err != nil {
		log.Fatalln(err)
	}
}

func initParser(u *url.URL) error {
	c, err := corenlp.NewClient(u, corenlp.Properties{
		Annotators:   corenlp.Annotators{corenlp.ParserAnnotator},
		OutputFormat: corenlp.FormatJSON,
	})

	if err != nil {
		return err
	}

	parser = c

	return nil
}

func initTokenizer(u *url.URL) error {
	c, err := corenlp.NewClient(u, corenlp.Properties{
		Annotators:   corenlp.Annotators{corenlp.TokenizerAnnotator},
		OutputFormat: corenlp.FormatJSON,
	})

	if err != nil {
		return err
	}

	tokenizer = c

	return nil
}

func main() {
	f, err := os.Create("mono-ykm_train.tsv")

	if err != nil {
		log.Fatalln(err)
	}

	w := csv.NewWriter(f)

	w.Comma = '\t'

	if err := w.Write([]string{"ID", "Tree", "Sentence", "Label"}); err != nil {
		log.Fatalln(err)
	}

	if err := MSRPC("data/msrpc/msr_paraphrase_train.txt", w); err != nil {
		log.Fatalln(err)
	}

	if err := PAWS("data/paws/train.tsv", w); err != nil {
		log.Fatalln(err)
	}

	if err := QQP("data/qqp_split/train.tsv", w); err != nil {
		log.Fatalln(err)
	}

	w.Flush()
}

func parse(s string) (string, error) {
	doc, err := parser.Annotate(s)

	if err != nil {
		return "", err
	}

	whitespaces := regexp.MustCompile(`[\s\n]+`)
	r := whitespaces.ReplaceAllString(doc.Sentences[0].Parse, " ")

	return r, nil
}

func tokenize(s string) (string, error) {
	rc, err := tokenizer.Run(s)

	if err != nil {
		return "", err
	}

	defer rc.Close()

	doc := &struct {
		Tokens []corenlp.Token `json:"tokens"`
	}{}

	if err := json.NewDecoder(rc).Decode(doc); err != nil {
		return "", err
	}

	sb := strings.Builder{}

	for i, t := range doc.Tokens {
		if i != 0 {
			sb.WriteString(" ")
		}

		sb.WriteString(t.Word)
	}

	return sb.String(), nil
}

func add(id, string1, string2 string, label bool, w *csv.Writer) error {
	t, err := parse(string1)

	if err != nil {
		return err
	}

	var l string

	if label {
		l = "1"
	} else {
		l = "0"
	}

	if err := w.Write([]string{id, t, string2, l}); err != nil {
		return err
	}

	return nil
}

func MSRPC(name string, w *csv.Writer) error {
	file, err := os.Open(name)

	if err != nil {
		return err
	}

	reader := csv.NewReader(file)

	reader.Comma = '\t'
	reader.LazyQuotes = true

	i := data.NewIterator(reader, msrpc.NewSample)

	i.Skip()

	for i.Next() {
		s := i.Sample().(*msrpc.Sample)

		if !s.Quality {
			continue
		}

		id1 := strings.Join([]string{"msrpc", strconv.Itoa(s.ID1), strconv.Itoa(s.ID2)}, "_")
		id2 := strings.Join([]string{"msrpc", strconv.Itoa(s.ID2), strconv.Itoa(s.ID1)}, "_")

		var string1 string
		var string2 string

		var err error

		if string1, err = tokenize(s.String1); err != nil {
			return err
		}

		if string2, err = tokenize(s.String2); err != nil {
			return err
		}

		if err := add(id1, string1, string2, s.Quality, w); err != nil {
			return err
		}

		if err := add(id2, string2, string1, s.Quality, w); err != nil {
			return err
		}
	}

	return nil
}

func PAWS(name string, w *csv.Writer) error {
	file, err := os.Open(name)

	if err != nil {
		return err
	}

	reader := csv.NewReader(file)

	reader.Comma = '\t'

	i := data.NewIterator(reader, paws.NewSample)

	i.Skip()

	for i.Next() {
		s := i.Sample().(*paws.Sample)

		if !s.Label {
			continue
		}

		id1 := strings.Join([]string{"paws", strconv.Itoa(s.ID), "1"}, "_")
		id2 := strings.Join([]string{"paws", strconv.Itoa(s.ID), "2"}, "_")

		if err := add(id1, s.Sentence1, s.Sentence2, s.Label, w); err != nil {
			return err
		}

		if err := add(id2, s.Sentence2, s.Sentence1, s.Label, w); err != nil {
			return err
		}
	}

	return nil
}

func QQP(name string, w *csv.Writer) error {
	file, err := os.Open(name)

	if err != nil {
		return err
	}

	reader := csv.NewReader(file)

	reader.Comma = '\t'

	factory := func(record []string) (interface{}, error) {
		if len(record) != 4 {
			return nil, errors.New("unexpected record length")
		}

		sample := &qqpSample{}

		if record[0] != "0" && record[0] != "1" {
			return nil, errors.New("unexpected judgement")
		}

		sample.judgement = record[0] == "1"

		sample.question1 = record[1]
		sample.question2 = record[2]

		if i, err := strconv.Atoi(record[3]); err != nil {
			return nil, err
		} else {
			sample.pairID = i
		}

		return sample, nil
	}

	i := data.NewIterator(reader, factory)

	for i.Next() {
		s := i.Sample().(*qqpSample)

		if !s.judgement {
			continue
		}

		id1 := strings.Join([]string{"qqp", strconv.Itoa(s.pairID), "1"}, "_")
		id2 := strings.Join([]string{"qqp", strconv.Itoa(s.pairID), "2"}, "_")

		if err := add(id1, s.question1, s.question2, s.judgement, w); err != nil {
			log.Printf("error adding %s: %v\n", id1, err)

			continue
		}

		if err := add(id2, s.question2, s.question1, s.judgement, w); err != nil {
			log.Printf("error adding %s: %v\n", id2, err)

			continue
		}
	}

	return nil
}
