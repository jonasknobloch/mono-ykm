package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"golang.org/x/sync/semaphore"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"
)

var corpus *Iterator
var model *Model

func init() {
	initCorpus()
}

func initCorpus() {
	c, err := NewIterator(Config.TrainingDataPath)

	if err != nil {
		log.Fatalln(err)
	}

	corpus = c
}

func initSample(sample *Sample) (*MetaTree, []string, error) {
	dec := tree.NewDecoder()

	t, err := dec.Decode(sample.Tree)

	if err != nil {
		return nil, nil, err
	}

	if Config.ReplaceSparseTokens && tokenOccurrences != nil {
		replaceSparseLabels(t.Leaves(), tokenOccurrences)
	}

	mt := NewMetaTree(t)

	mt.CollectFeatures()

	e := strings.Split(sample.Sentence, " ")

	if Config.ReplaceSparseTokens && tokenOccurrences != nil {
		replaceSparseTokens(e, tokenOccurrences)
	}

	if mt.Tree.Size()+len(mt.Tree.Leaves()) < len(e) {
		return nil, nil, errors.New("target sentence unreachable")
	}

	c, ok := O(mt.Tree, len(e))

	if !ok || (Config.TrainingComplexityLimit != -1 && Config.TrainingComplexityLimit < c) {
		return nil, nil, errors.New("sample exceeds complexity limit")
	}

	return mt, e, nil
}

func importModel(name string) (*Model, error) {
	fmt.Println("Importing model...")

	n, err := Import(name + "-n.gob")

	if err != nil {
		return nil, err
	}

	r, err := Import(name + "-r.gob")

	if err != nil {
		return nil, err
	}

	t, err := Import(name + "-t.gob")

	if err != nil {
		return nil, err
	}

	return &Model{n, r, t}, nil
}

func TrainEM(iterations, samples int) {
	if Config.ReplaceSparseTokens {
		initTokenOccurrences()
	}

	var o int

	if Config.InitModelPath != "" {
		if m, err := importModel(Config.InitModelPath); err != nil {
			log.Fatal(err)
		} else {
			model = m
			o = Config.InitModelIteration
		}
	} else {
		model = NewModel()
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

		lh := big.NewFloat(1)

		for corpus.Next() && (samples == -1 || eval < samples) {
			if !corpus.Sample().Label {
				continue
			}

			sample := corpus.Sample()

			mt, e, err := initSample(sample)

			if err != nil {
				skip++

				fmt.Printf("Skipped sample %s (%s)\n", sample.ID, err)

				continue
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				log.Fatalf("Failed to acquire semaphore: %v", err)
			}

			wg.Add(1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Panic while evaluating sample %s", sample.ID)
						panic(r)
					}

					defer sem.Release(1)
					defer wg.Done()
				}()

				w := NewStopWatch()

				w.Start()

				g := NewGraph(mt, e, model)

				p := g.pBeta[g.nodes[0]]

				lh.Mul(lh, p)

				if Config.ExportGraphs {
					g.Draw()
				}

				nC.ForEach(g.insertions, g.InsertionCount)
				nR.ForEach(g.reorderings, g.ReorderingCount)
				nT.ForEach(g.translations, g.TranslationCount)

				w.Stop()

				fmt.Printf("Evaluated sample %s (eval: %d skip: %d) [%s] [%e]\n", sample.ID, eval, skip, w.Result(), p)
			}()

			eval++
		}

		wg.Wait()

		watch.Lap("samples")

		fmt.Printf("\nAdjusting model weights...\n")

		if err := model.UpdateWeights(nC, nR, nT); err != nil {
			log.Fatalf("Error updating model weights: %v", err)
		}

		watch.Lap("weights")

		if Config.ExportModel {
			_ = Export(model.n, strconv.Itoa(i), "n")
			_ = Export(model.r, strconv.Itoa(i), "r")
			_ = Export(model.t, strconv.Itoa(i), "t")

			watch.Lap("export")
		}

		if Config.PrintCorpusLikelihood {
			fmt.Printf("\nCorpus likelihood: %e", lh)
		}

		// https://github.com/golang/go/issues/11068

		lhExp := math.Log10(2) * float64(lh.MantExp(nil))

		fmt.Printf("\nLikelihood exponent: %d\n\n", int(lhExp))

		watch.Lap("likelihood")

		watch.Stop()

		fmt.Printf("%s", watch)

		watch.Reset()
	}
}
