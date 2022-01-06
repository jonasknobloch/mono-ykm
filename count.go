package main

type Count map[string]map[string]float64

func NewCount() Count {
	return make(map[string]map[string]float64)
}

func (c Count) Add(feature, key string, value float64) {
	if _, ok := c[feature]; !ok {
		c[feature] = make(map[string]float64)
	}

	if _, ok := c[feature][key]; !ok {
		c[feature][key] = 0
	}

	c[feature][key] += value
}

func (c Count) Get(feature, key string) float64 {
	return c[feature][key]
}

func (c Count) ForEach(p map[string]map[string]float64, f func(string, string) float64) {
	for feature, keys := range p {
		for key := range keys {
			c.Add(feature, key, f(feature, key))
		}
	}
}

func (c Count) Sum(feature string) float64 {
	sum := float64(0)

	for _, value := range c[feature] {
		sum += value
	}

	return sum
}

func (c Count) Reset() {
	for feature, keys := range c {
		for key := range keys {
			c[feature][key] = 0
		}
	}
}
