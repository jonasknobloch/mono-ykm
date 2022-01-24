package main

import (
	"flag"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"log"
	"os"
	"runtime/pprof"
)

var cpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")

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

	tr := &tree.Tree{
		Label: "σ",
		Children: []*tree.Tree{
			{
				Label: "γ",
				Children: []*tree.Tree{
					{
						Label:    "α",
						Children: nil,
					},
				},
			},
			{
				Label: "γ",
				Children: []*tree.Tree{
					{
						Label:    "β",
						Children: nil,
					},
				},
			},
		},
	}

	pCache["α β"] = NewMetaTree(tr)

	TrainEM(Config.TrainingIterationLimit, Config.TrainingSampleLimit)
}
