package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/msrpc"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"io"
	"log"
	"net/url"
	"os"
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
	c, err := msrpc.NewIterator(Config.TrainingDataPath)

	if err != nil {
		panic(err)
	}

	corpus = c
}

func initParser() {
	u, _ := url.Parse(Config.CoreNLPUrl)

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
	u, _ := url.Parse(Config.CoreNLPUrl)

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

		if len(tr.Children) != 1 && tr.Label != "ROOT" {
			return nil, errors.New("unexpected tree structure")
		}

		pCache[str] = NewMetaTree(tr.Children[0])
	}

	return pCache[str], nil
}

func importTrees(name string) error {
	var r *csv.Reader

	if f, err := os.Open(name); err != nil {
		return fmt.Errorf("error opening file: %w", err)
	} else {
		r = csv.NewReader(f)
		defer f.Close()
	}

	r.Comma = '\t'

	_, err := r.Read()

	if err != nil {
		return err
	}

	dec := tree.NewDecoder()

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		tr, err := dec.Decode(record[1])

		if err != nil {
			return err
		}

		pCache[record[0]] = NewMetaTree(tr)
	}

	return nil
}

func TrainEM(iterations, samples int) {
	if Config.TreeMockDataPath != "" {
		err := importTrees(Config.TreeMockDataPath)

		if err != nil {
			log.Fatal(err)
		}
	}

	m := NewModel()

	fmt.Println("Building dictionaries...")

	nDict, tDict := buildDictionaries(samples)

	fmt.Println("Initializing weights...")

	m.InitInsertionWeights(nDict)
	m.InitTranslationWeights(tDict)

	if Config.ExportModel {
		_ = Export(m.n, strconv.Itoa(0), "n")
		_ = Export(m.t, strconv.Itoa(0), "t")
	}

	nC := NewCount()
	nR := NewCount()
	nT := NewCount()

	watch := NewStopWatch()

	for i := 1; i < iterations+1; i++ {
		watch.Start()

		fmt.Printf("\nStarting training iteration #%d\n", i)

		initCorpus() // TODO just reset iterator

		fmt.Println("Resetting counts...")

		nC.Reset()
		nR.Reset()
		nT.Reset()

		watch.Lap("init")

		counter := 0

		for corpus.Next() && (samples == -1 || counter < samples) {
			if !corpus.Sample().Quality {
				continue
			}

			sample := corpus.Sample()

			fmt.Printf("Analyzing sample %d_%d (#%d)\n", sample.ID1, sample.ID2, counter)

			mt, err := parse(sample.String1)

			if err != nil {
				panic(err)
			}

			f, err := tokenize(sample.String2)

			if err != nil {
				panic(err)
			}

			c, ok := O(mt.Tree, len(f))

			fmt.Printf("Complexity estimation: %d (%t)\n", c, ok)

			if !ok || (Config.TrainingComplexityLimit != -1 && Config.TrainingComplexityLimit < c) {
				fmt.Println("Skipping sample...")
				continue
			}

			watch.Lap(fmt.Sprintf("#%d init", counter))

			g := NewGraph(mt, f, m)

			watch.Lap(fmt.Sprintf("#%d graph", counter))

			fmt.Printf("Nodes: %d Edges: %d\n", len(g.nodes), len(g.edges))
			fmt.Printf("Alpha: %e Beta: %e\n", g.Alpha(g.nodes[0]), g.Beta(g.nodes[0]))

			watch.Lap(fmt.Sprintf("#%d validation", counter))

			fmt.Println("Updating counts...")

			nC.ForEach(m.n, g.InsertionCount)
			nR.ForEach(m.r, g.ReorderingCount)
			nT.ForEach(m.t, g.TranslationCount)

			watch.Lap(fmt.Sprintf("#%d count updates", counter))

			if Config.ExportGraphs {
				g.Draw(strconv.Itoa(i), strconv.Itoa(counter))
			}

			watch.Lap(fmt.Sprintf("#%d graph export", counter))

			counter++
		}

		fmt.Println("Adjusting model weights...")

		m.UpdateWeights(nC, nR, nT)

		watch.Lap("weights")

		if Config.ExportModel {
			_ = Export(m.n, strconv.Itoa(i), "n")
			_ = Export(m.r, strconv.Itoa(i), "r")
			_ = Export(m.t, strconv.Itoa(i), "t")

			watch.Lap("model export")
		}

		watch.Stop()

		fmt.Printf("%s", watch)

		watch.Reset()
	}
}

func buildDictionaries(samples int) (map[string]map[string]int, map[string]map[string]int) {
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

	counter := 0

	for corpus.Next() && (samples == -1 || counter < samples) {
		sample := corpus.Sample()

		if !sample.Quality {
			continue
		}

		// TODO consider both directions

		mt, err := parse(sample.String1)

		if err != nil {
			panic(err)
		}

		e, err := tokenize(sample.String2)

		if err != nil {
			panic(err)
		}

		c, ok := O(mt.Tree, len(e))

		if !ok || (Config.TrainingComplexityLimit != -1 && Config.TrainingComplexityLimit < c) {
			continue
		}

		mt.Tree.Walk(func(st *tree.Tree) {
			for _, i := range Insertions(st, e, mt.meta[st][0], true) {
				addParameter(insertions, i.Feature(), i.Key())
			}

			for _, t := range Translations(st, e, mt.meta[st][2]) {
				addParameter(translations, t.Feature(), t.Key())
			}
		})

		counter++
	}

	return insertions, translations
}
