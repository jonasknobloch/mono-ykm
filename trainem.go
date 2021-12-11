package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/msrpc"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"net/url"
)

var corpus *msrpc.Iterator
var parser *corenlp.Client
var tokenizer *corenlp.Client

var pCache map[string]*MetaTree
var tCache map[string][]string

func init() {
	initCorpus()
	initParser()
	initTokenizer()

	pCache = make(map[string]*MetaTree)
	tCache = make(map[string][]string)
}

func initCorpus() {
	c, err := msrpc.NewIterator("test/msr_paraphrase_mock.txt")

	if err != nil {
		panic(err)
	}

	corpus = c
}

func initParser() {
	u, _ := url.Parse("http://localhost:9000")

	c, err := corenlp.NewClient(u, corenlp.Properties{
		Annotators:   corenlp.Annotators{corenlp.ParserAnnotator},
		OutputFormat: corenlp.FormatJSON,
	})

	if err != nil {
		panic(err)
	}

	parser = c
}

func initTokenizer() {
	u, _ := url.Parse("http://localhost:9000")

	c, err := corenlp.NewClient(u, corenlp.Properties{
		Annotators:   corenlp.Annotators{corenlp.WordsToSentencesAnnotator},
		OutputFormat: corenlp.FormatJSON,
	})

	if err != nil {
		panic(err)
	}

	tokenizer = c
}

func tokenize(str string) ([]string, error) {
	if _, ok := tCache[str]; !ok {
		d, err := tokenizer.Annotate(str)

		if err != nil {
			return nil, err
		}

		e := make([]string, 0, len(d.Sentences[0].Tokens))

		for _, t := range d.Sentences[0].Tokens {
			e = append(e, t.Word)
		}

		tCache[str] = e
	}

	return tCache[str], nil
}

func parse(str string) (*MetaTree, error) {
	if _, ok := pCache[str]; !ok {
		doc, err := parser.Annotate(str)

		if err != nil {
			return nil, err
		}

		dec := tree.NewDecoder()
		p := doc.Sentences[0].Parse

		tr, err := dec.Decode(p)

		if err != nil {
			return nil, err
		}

		pCache[str] = NewMetaTree(tr)
	}

	return pCache[str], nil
}

func TrainEM() {
	m := NewModel()

	fmt.Println("Building dictionaries...")

	nList, tDict := buildDictionaries()

	fmt.Println("Initializing weights...")

	m.InitInsertionWeights(nList)
	m.InitTranslationWeights(tDict)

	initCorpus() // TODO method arg for iterator?

	limit := 1
	counter := 0

	for corpus.Next() && counter < limit {
		sample := corpus.Sample()

		fmt.Println("Analyzing sample...")
		fmt.Println("Parsing source sentence...")

		mt, _ := parse(sample.String1)

		fmt.Println("Tokenizing target sentence...")

		f, _ := tokenize(sample.String2)

		fmt.Println("Generating training graph...")

		g := NewGraph(mt, f, m)

		fmt.Printf("\nNodes: %d (%d) Edges: %d\n", len(g.nodes)-len(g.pruned), len(g.nodes), len(g.edges))
		fmt.Printf("Alpha: %f Beta: %f\n\n", g.Alpha(g.nodes[0]), g.Beta(g.nodes[0]))

		fmt.Println("Drawing graph...")

		g.Draw()

		counter++
	}
}

func buildDictionaries() (map[string]int, map[string]map[string]int) {
	insertions := make(map[string]int)
	translations := make(map[string]map[string]int)

	addInsertion := func(w string) {
		if _, ok := insertions[w]; !ok {
			insertions[w] = 0
		}

		insertions[w]++
	}

	addTranslation := func(w1, w2 string) {
		if _, ok := translations[w1]; !ok {
			translations[w1] = make(map[string]int)
		}

		if _, ok := translations[w1][w2]; !ok {
			translations[w1][w2] = 0
		}

		translations[w1][w2]++
	}

	for corpus.Next() {
		sample := corpus.Sample()

		if !sample.Quality {
			continue
		}

		e, _ := tokenize(sample.String1)
		f, _ := tokenize(sample.String2)

		for _, w := range f {
			addInsertion(w)
		}

		for _, s := range e {
			for _, t := range f {
				addTranslation(s, t)
			}
		}

		// TODO consider both directions
		//
		// e1, _ := tokenize(sample.String1)
		// e2, _ := tokenize(sample.String2)
		//
		// for _, w := range append(e1, e2...) {
		// 	addInsertion(w)
		// }
		//
		// for _, w1 := range e1 {
		// 	for _, w2 := range e2 {
		// 		addTranslation(w1, w2)
		// 		addTranslation(w2, w1)
		// 	}
		// }
	}

	return insertions, translations
}
