package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/msrpc"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"net/url"
	"strconv"
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

func TrainEM(iterations, samples int) {
	m := NewModel()

	fmt.Println("Building dictionaries...")

	nDict, tDict := buildDictionaries()

	fmt.Println("Initializing weights...")

	m.InitInsertionWeights(nDict)
	m.InitTranslationWeights(tDict)

	nC := NewCount()
	nR := NewCount()
	nT := NewCount()

	for i := 0; i < iterations; i++ {
		fmt.Printf("\nStarting training iteration #%d\n", i)

		initCorpus() // TODO just reset iterator

		fmt.Println("Resetting counts...")

		nC.Reset()
		nR.Reset()
		nT.Reset()

		limit := samples
		counter := 0

		for corpus.Next() && (limit == -1 || counter < limit) {
			sample := corpus.Sample()

			fmt.Printf("Analyzing sample #%d\n", counter)

			mt, _ := parse(sample.String1)
			f, _ := tokenize(sample.String2)

			g := NewGraph(mt, f, m)

			fmt.Printf("Nodes: %d (%d) Edges: %d\n", len(g.nodes)-len(g.pruned), len(g.nodes), len(g.edges))
			fmt.Printf("Alpha: %e Beta: %e\n", g.Alpha(g.nodes[0]), g.Beta(g.nodes[0]))

			fmt.Println("Updating counts...")

			nC.ForEach(m.n, g.InsertionCount)
			nR.ForEach(m.r, g.ReorderingCount)
			nT.ForEach(m.t, g.TranslationCount)

			g.Draw(strconv.Itoa(i), strconv.Itoa(counter))

			counter++
		}

		fmt.Println("Adjusting model weights...")

		m.UpdateWeights(nC, nR, nT)
	}
}

func buildDictionaries() (map[string]map[string]int, map[string]map[string]int) {
	insertions := make(map[string]map[string]int)
	translations := make(map[string]map[string]int)

	addParameter := func(m map[string]map[string]int, f, k string) {
		if _, ok := m[f]; !ok {
			m[f] = make(map[string]int)
		}

		if _, ok := m[f][k]; !ok {
			m[f][k] = 0
		}

		m[f][k]++
	}

	for corpus.Next() {
		sample := corpus.Sample()

		if !sample.Quality {
			continue
		}

		// TODO consider both directions

		mt, _ := parse(sample.String1)
		e, _ := tokenize(sample.String2)

		mt.Tree.Walk(func(st *tree.Tree) {
			for _, i := range Insertions(st, e, mt.meta[st][0]) {
				addParameter(insertions, i.Feature(), i.Key())
			}

			for _, t := range Translations(st, e, mt.meta[st][2]) {
				addParameter(translations, t.Feature(), t.Key())
			}
		})
	}

	return insertions, translations
}
