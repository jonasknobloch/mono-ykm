package main

import "github.com/jonasknobloch/jinn/pkg/tree"

func main() {
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
