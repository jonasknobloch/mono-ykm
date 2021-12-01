package main

import (
	"fmt"
	"github.com/jonasknobloch/jinn/pkg/corenlp"
	"github.com/jonasknobloch/jinn/pkg/msrpc"
	"github.com/jonasknobloch/jinn/pkg/tree"
	"net/url"
)

func main() {
	u, _ := url.Parse("http://localhost:9000")

	c, _ := corenlp.NewClient(u, corenlp.Properties{
		Annotators:   corenlp.Annotators{corenlp.ParserAnnotator},
		OutputFormat: corenlp.FormatJSON,
	})

	dec := tree.NewDecoder()

	it, err := msrpc.NewIterator("msr_paraphrase_train.txt")

	if err != nil {
		panic(err)
	}

	labels := make(map[string]int)

	for it.Next() {
		for _, s := range [2]string{it.Sample().String1, it.Sample().String2} {
			doc, _ := c.Annotate(s)
			tr, _ := dec.Decode(doc.Sentences[0].Parse)

			tr.Walk(func(st *tree.Tree) {
				if len(st.Children) > 0 {
					if _, ok := labels[st.Label]; ok {
						labels[st.Label]++
					} else {
						labels[st.Label] = 1
					}
				}
			})
		}
	}

	for k, v := range labels {
		fmt.Printf("\"%s\" x %d \n", k, v)
	}
}
