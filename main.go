package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
)

var cpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")
var execMode = flag.String("m", "train", "choose execution mode")

const ModeTrain = "train"
const ModeEvaluate = "evaluate"

func main() {
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)

		if err != nil {
			log.Fatal(err)
		}

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}

		defer pprof.StopCPUProfile()
	}

	switch *execMode {
	case ModeTrain:
		TrainEM(Config.TrainingIterationLimit, Config.TrainingSampleLimit)
	case ModeEvaluate:
		Evaluate()
	}
}
