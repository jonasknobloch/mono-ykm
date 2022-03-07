package main

import (
	"math/big"
	"sync"
)

type Count struct {
	val map[string]map[string]*big.Float
	rwm sync.RWMutex
}

func NewCount() *Count {
	return &Count{
		val: make(map[string]map[string]*big.Float),
		rwm: sync.RWMutex{},
	}
}

func (c *Count) Add(feature, key string, value *big.Float) {
	c.rwm.Lock()
	defer c.rwm.Unlock()

	if _, ok := c.val[feature]; !ok {
		c.val[feature] = make(map[string]*big.Float)
	}

	if _, ok := c.val[feature][key]; !ok {
		c.val[feature][key] = new(big.Float)
	}

	c.val[feature][key].Add(c.val[feature][key], value)
}

func (c *Count) Get(feature, key string) *big.Float {
	return c.val[feature][key]
}

func (c *Count) ForEach(p map[string]map[string][]*Node, f func(string, string) (*big.Float, bool)) {
	for feature, keys := range p {
		for key := range keys {
			val, ok := f(feature, key)

			if !ok {
				continue
			}

			c.Add(feature, key, val)
		}
	}
}

func (c *Count) Sum(feature string) *big.Float {
	sum := new(big.Float)

	for _, value := range c.val[feature] {
		sum.Add(sum, value)
	}

	return sum
}

func (c *Count) Size(feature string) int {
	return len(c.val[feature])
}

func (c *Count) Reset() {
	for feature, keys := range c.val {
		for key := range keys {
			c.val[feature][key].SetFloat64(0)
		}
	}
}
