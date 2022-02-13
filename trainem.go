package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/msrpc"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"golang.org/x/sync/semaphore"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"sync"
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

		if Config.ReplaceSparseTokens {
			if tokenOccurrences == nil {
				return e, nil
			}

			replaceSparseTokens(e, tokenOccurrences)
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

		if Config.ReplaceSparseTokens {
			replaceSparseLabels(tr.Leaves(), tokenOccurrences)
		}

		pCache[str] = NewMetaTree(tr.Children[0])
	}

	return pCache[str], nil
}

func initModel(samples int) *Model {
	m := NewModel()

	fmt.Println("Building dictionaries...")

	nDict, rDict, tDict := buildDictionaries(samples)

	fmt.Println("Initializing weights...")

	m.InitTable(m.n, nDict)
	m.InitTable(m.r, rDict)
	m.InitTable(m.t, tDict)

	if Config.ExportModel {
		_ = Export(m.n, strconv.Itoa(0), "n")
		_ = Export(m.r, strconv.Itoa(0), "r")
		_ = Export(m.t, strconv.Itoa(0), "t")
	}

	return m
}

func initSample(sentence1, sentence2 string) (*MetaTree, []string, error) {
	mt, err := parse(sentence1)

	if err != nil {
		return nil, nil, err
	}

	e, err := tokenize(sentence2)

	if err != nil {
		return nil, nil, err
	}

	unreachable := !Config.AllowTerminalInsertions && mt.Tree.Size() < len(e)
	unreachable = unreachable || Config.AllowTerminalInsertions && mt.Tree.Size()+len(mt.Tree.Leaves()) < len(e)

	if unreachable {
		return nil, nil, errors.New("target sentence unreachable")
	}

	c, ok := O(mt.Tree, len(e))

	if !ok || (Config.TrainingComplexityLimit != -1 && Config.TrainingComplexityLimit < c) {
		return nil, nil, errors.New("sample exceeds complexity limit")
	}

	return mt, e, nil
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

func importModel(name string) (*Model, error) {
	fmt.Println("Importing model...")

	n, err := Import(name + "-n.tsv")

	if err != nil {
		return nil, err
	}

	r, err := Import(name + "-r.tsv")

	if err != nil {
		return nil, err
	}

	t, err := Import(name + "-t.tsv")

	if err != nil {
		return nil, err
	}

	return &Model{n, r, t}, nil
}

func TrainEM(iterations, samples int) {
	if Config.TreeMockDataPath != "" {
		err := importTrees(Config.TreeMockDataPath)

		if err != nil {
			log.Fatal(err)
		}
	}

	if Config.ReplaceSparseTokens {
		initTokenOccurrences()
	}

	var m *Model
	var o int

	if Config.InitModelPath != "" {
		if model, err := importModel(Config.InitModelPath); err != nil {
			log.Fatal(err)
		} else {
			m = model
			o = Config.InitModelIteration
		}
	} else {
		m = initModel(samples)
	}

	nC := NewCount()
	nR := NewCount()
	nT := NewCount()

	ctx := context.TODO()
	sem := semaphore.NewWeighted(int64(Config.ConcurrentSampleEvaluations))

	var wg sync.WaitGroup

	watch := NewStopWatch()

	for i := 1 + o; i < iterations+o+1; i++ {
		watch.Start()

		fmt.Printf("\nStarting training iteration #%d\n\n", i)

		initCorpus() // TODO just reset iterator

		nC.Reset()
		nR.Reset()
		nT.Reset()

		watch.Lap("init")

		eval := 0
		skip := 0

		for corpus.Next() && (samples == -1 || eval < samples) {
			if !corpus.Sample().Quality {
				continue
			}

			sample := corpus.Sample()

			mt, f, err := initSample(sample.String1, sample.String2)

			if err != nil {
				skip++

				fmt.Printf("Skipped sample %d_%d (%s)\n", sample.ID1, sample.ID2, err)

				continue
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				log.Fatalf("Failed to acquire semaphore: %v", err)
			}

			wg.Add(1)

			go func() {
				defer sem.Release(1)
				defer wg.Done()

				w := NewStopWatch()

				w.Start()

				g := NewGraph(mt, f, m)

				if Config.ExportGraphs {
					g.Draw()
				}

				nC.ForEach(m.n, g.InsertionCount)
				nR.ForEach(m.r, g.ReorderingCount)
				nT.ForEach(m.t, g.TranslationCount)

				w.Stop()

				fmt.Printf("Evaluated sample %d_%d (eval: %d skip: %d) [%s]\n", sample.ID1, sample.ID2, eval, skip, w.Result())
			}()

			eval++
		}

		wg.Wait()

		watch.Lap("samples")

		fmt.Printf("\nAdjusting model weights...\n\n")

		m.UpdateWeights(nC, nR, nT)

		watch.Lap("weights")

		if Config.ExportModel {
			_ = Export(m.n, strconv.Itoa(i), "n")
			_ = Export(m.r, strconv.Itoa(i), "r")
			_ = Export(m.t, strconv.Itoa(i), "t")

			watch.Lap("export")
		}

		watch.Stop()

		fmt.Printf("%s", watch)

		watch.Reset()
	}
}

func buildDictionaries(samples int) (map[string]map[string]int, map[string]map[string]int, map[string]map[string]int) {
	insertions := make(map[string]map[string]int)
	reorderings := make(map[string]map[string]int)
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

	initCorpus()

	for corpus.Next() && (samples == -1 || counter < samples) {
		sample := corpus.Sample()

		if !sample.Quality {
			continue
		}

		// TODO consider both directions

		mt, e, err := initSample(sample.String1, sample.String2)

		if err != nil {
			continue
		}

		mt.Tree.Walk(func(st *tree.Tree) {
			for _, i := range Insertions(st, e, mt.meta[st][0], true) {
				addParameter(insertions, i.Feature(), i.Key())
			}

			for _, r := range Reorderings(st, mt.meta[st][1]) {
				addParameter(reorderings, r.Feature(), r.Key())
			}

			for _, t := range Translations(st, e, mt.meta[st][2]) {
				addParameter(translations, t.Feature(), t.Key())
			}
		})

		counter++
	}

	return insertions, reorderings, translations
}
