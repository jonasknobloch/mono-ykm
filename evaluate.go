package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
	"math/big"
	"sync"
)

func Evaluate() {
	if m, err := importModel(Config.InitModelPath); err != nil {
		log.Fatal(err)
	} else {
		model = m
	}

	Verify(model, big.NewFloat(1e-5))

	tp := 0
	fp := 0
	tn := 0
	fn := 0

	pth := big.NewFloat(Config.ParaphraseThreshold)

	counter := 0

	numPos := 0
	numNeg := 0

	pos := new(big.Float)
	neg := new(big.Float)

	ctx := context.TODO()
	sem := semaphore.NewWeighted(int64(Config.ConcurrentSampleEvaluations))

	var wg sync.WaitGroup

	for corpus.Next() && (Config.TrainingSampleLimit == -1 || counter < Config.TrainingSampleLimit) {
		sample := corpus.Sample()

		mt, e, err := initSample(sample)

		if err != nil {
			continue
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			log.Fatalf("Failed to acquire semaphore: %v", err)
		}

		wg.Add(1)

		go func() {
			defer sem.Release(1)
			defer wg.Done()

			g, err := NewGraph(mt, e, model)

			if err != nil {
				return
			}

			p := g.pBeta[g.nodes[0]]

			if sample.Label && p.Cmp(pth) == 1 {
				tp++
			}

			if !sample.Label && p.Cmp(pth) == 1 {
				fp++
			}

			if !sample.Label && p.Cmp(pth) == -1 {
				tn++
			}

			if sample.Label && p.Cmp(pth) == -1 {
				fn++
			}

			fmt.Printf("TP: %d FP: %d TN: %d FN: %d (%e) [%s %t]\n", tp, fp, tn, fn, p, sample.ID, sample.Label)

			if p.Cmp(new(big.Float)) == 0 {
				return
			}

			if sample.Label {
				pos.Add(pos, p)
				numPos++
			} else {
				neg.Add(neg, p)
				numNeg++
			}
		}()

		counter++
	}

	wg.Wait()

	precision := float64(tp) / float64(tp+fp)
	recall := float64(tp) / float64(tp+fn)

	f := float64(2) * precision * recall / (precision + recall)

	avgPos := pos.Quo(pos, big.NewFloat(float64(numPos)))
	avgNeg := neg.Quo(neg, big.NewFloat(float64(numNeg)))

	mean := new(big.Float)

	mean.Add(mean, avgPos).Add(mean, avgNeg).Quo(mean, big.NewFloat(2))

	fmt.Printf("Precision: %e Recall: %e F1: %e\n", precision, recall, f)
	fmt.Printf("AvgPos: %e AvgNeg: %e Mean: %e\n", avgPos, avgNeg, mean)
}

func Verify(model *Model, threshold *big.Float) {
	verifyTable := func(table map[string]map[string]*big.Float) {
		for k, v := range table {
			sum := new(big.Float)

			for _, p := range v {
				sum.Add(sum, p)
			}

			upper := new(big.Float).Add(big.NewFloat(1), threshold)
			lower := new(big.Float).Sub(big.NewFloat(1), threshold)

			if sum.Cmp(upper) == 1 || sum.Cmp(lower) == -1 {
				fmt.Println(k, sum)
			}
		}
	}

	verifyTable(model.n)
	verifyTable(model.r)
	verifyTable(model.t)
}
