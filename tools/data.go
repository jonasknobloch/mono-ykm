package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/msrpc"
	"github.com/jonasknobloch/jinn/pkg/paws"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var parser *corenlp.Client
var tokenizer *corenlp.Client

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

	if err := MSRPC("msrpc/msr_paraphrase_train.txt", w); err != nil {
		log.Fatalln(err)
	}

	if err := PAWS("paws/train.tsv", w); err != nil {
		log.Fatalln(err)
	}

	if err := QQP("qqp/quora_duplicate_questions.tsv", w); err != nil {
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

func MSRPC(name string, w *csv.Writer) error {
	i, err := msrpc.NewIterator(name)

	if err != nil {
		return err
	}

	add := func(id1, id2 int, string1, string2 string, quality bool) error {
		t, err := parse(string1)

		if err != nil {
			return err
		}

		s, err := tokenize(string2)

		if err != nil {
			return err
		}

		var l string

		if quality {
			l = "1"
		} else {
			l = "0"
		}

		id := strings.Join([]string{"msrpc", strconv.Itoa(id1), strconv.Itoa(id2)}, "_")

		if err := w.Write([]string{id, t, s, l}); err != nil {
			return err
		}

		return nil
	}

	for i.Next() {
		s := i.Sample()

		if !s.Quality {
			continue
		}

		if err := add(s.ID1, s.ID2, s.String1, s.String2, s.Quality); err != nil {
			return err
		}

		if err := add(s.ID2, s.ID1, s.String2, s.String1, s.Quality); err != nil {
			return err
		}
	}

	return nil
}

func PAWS(name string, w *csv.Writer) error {
	i, err := paws.NewIterator(name)

	if err != nil {
		return err
	}

	add := func(id, sentence1, sentence2 string, quality bool) error {
		t, err := parse(sentence1)

		if err != nil {
			return err
		}

		var l string

		if quality {
			l = "1"
		} else {
			l = "0"
		}

		if err := w.Write([]string{id, t, sentence2, l}); err != nil {
			return err
		}

		return nil
	}

	for i.Next() {
		s := i.Sample()

		if !s.Label {
			continue
		}

		id1 := strings.Join([]string{"paws", strconv.Itoa(s.ID), "1"}, "_")
		id2 := strings.Join([]string{"paws", strconv.Itoa(s.ID), "2"}, "_")

		if err := add(id1, s.Sentence1, s.Sentence2, s.Label); err != nil {
			return err
		}

		if err := add(id2, s.Sentence2, s.Sentence1, s.Label); err != nil {
			return err
		}
	}

	return nil
}

func QQP(name string, w *csv.Writer) error {
	f, err := os.Open(name)

	if err != nil {
		return err
	}

	r := csv.NewReader(f)

	r.Comma = '\t'

	record, err := r.Read()

	if err != nil {
		return err
	}

	var header [6]string
	copy(header[:], record[0:6])

	if header != [6]string{"id", "qid1", "qid2", "question1", "question2", "is_duplicate"} {
		return errors.New("unexpected header row")
	}

	add := func(id, question1, question2, duplicate string) error {
		t, err := parse(question1)

		if err != nil {
			return err
		}

		s, err := tokenize(question2)

		if err != nil {
			return err
		}

		if err := w.Write([]string{id, t, s, duplicate}); err != nil {
			return err
		}

		return nil
	}

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if record[5] != "1" {
			continue
		}

		id1 := strings.Join([]string{"qqp", record[0], record[1], record[2]}, "_")
		id2 := strings.Join([]string{"qqp", record[0], record[2], record[1]}, "_")

		if err := add(id1, record[3], record[4], record[5]); err != nil {
			return err
		}

		if err := add(id2, record[4], record[3], record[5]); err != nil {
			return err
		}
	}

	return nil
}
