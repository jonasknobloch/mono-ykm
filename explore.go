package main

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"os"
	"reflect"
	"sort"
)

type pair struct {
	key   string
	value *big.Float
}

type pairs []pair

func (ps pairs) Len() int {
	return len(ps)
}

func (ps pairs) Less(i, j int) bool {
	return ps[i].value.Cmp(ps[j].value) == -1
}

func (ps pairs) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func Explore() {
	if m, err := importModel(Config.InitModelPath); err != nil {
		log.Fatal(err)
	} else {
		model = m
	}

	sum := func(m map[string]*big.Float) *big.Float {
		sum := new(big.Float)

		for _, p := range m {
			sum.Add(sum, p)
		}

		return sum
	}

	var t map[string]map[string]*big.Float

	scanner := bufio.NewScanner(os.Stdin)

	var table string
	var feature string

	for {
		if table == "" {
			fmt.Print("table=")
		} else if feature == "" {
			fmt.Print("feature=")
		} else {
			fmt.Printf("key=")
		}

		scanner.Scan()

		text := scanner.Text()

		if text == "$$$" {
			if table == "" {
				return
			}

			if feature == "" {
				table = ""
			} else {
				feature = ""
			}

			continue
		}

		if table == "" {
			if text == "$$" {
				t := reflect.ValueOf(model).Elem().Type()

				for i := 0; i < t.NumField(); i++ {
					fmt.Println(t.Field(i).Name)
				}

				continue
			}

			switch text {
			case "n":
				t = model.n
			case "r":
				t = model.r
			case "t":
				t = model.t
			case "l":
				t = model.l
			case "f":
				t = model.f
			default:
				fmt.Println("unknown table")
				continue
			}

			table = text
			continue
		}

		if feature == "" {
			if text == "$$" {
				for f := range t {
					fmt.Printf("[%e]\t%s\n", sum(t[f]), f)
				}

				continue
			}

			if _, ok := t[text]; ok {
				feature = text
			} else {
				fmt.Println("unknown feature")
			}

			continue
		}

		if text == "$$" {
			ps := make(pairs, 0, len(t[feature]))

			for k, p := range t[feature] {
				ps = append(ps, pair{
					key:   k,
					value: p,
				})
			}

			sort.Sort(ps)

			for _, p := range ps {
				fmt.Printf("[%e]\t%s : %s\n", p.value, feature, p.key)
			}

			continue
		}

		if val, ok := t[feature][text]; ok {
			fmt.Println(val)
		} else {
			fmt.Println("unknown key")
		}
	}
}
