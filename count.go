package main

import (
	"sync"
)

type Count struct {
	val map[string]map[string]float64
	rwm sync.RWMutex
}

func NewCount() *Count {
	return &Count{
		val: make(map[string]map[string]float64),
		rwm: sync.RWMutex{},
	}
}

func (c *Count) Add(feature, key string, value float64) {
	c.rwm.Lock()
	defer c.rwm.Unlock()

	if _, ok := c.val[feature]; !ok {
		c.val[feature] = make(map[string]float64)
	}

	if _, ok := c.val[feature][key]; !ok {
		c.val[feature][key] = 0
	}

	c.val[feature][key] += value
}

func (c *Count) Get(feature, key string) float64 {
	return c.val[feature][key]
}

func (c *Count) ForEach(p map[string]map[string]float64, f func(string, string) (float64, bool)) {
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

func (c *Count) Sum(feature string) float64 {
	sum := float64(0)

	for _, value := range c.val[feature] {
		sum += value
	}

	return sum
}

func (c *Count) Reset() {
	for feature, keys := range c.val {
		for key := range keys {
			c.val[feature][key] = 0
		}
	}
}
