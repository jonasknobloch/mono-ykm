package main

import (
	"fmt"
	"strings"
	"time"
)

type Stopwatch struct {
	laps  []Lap
	last  time.Time
	start time.Time
	stop  time.Time
}

type Lap struct {
	Label    string
	Duration time.Duration
}

func NewStopWatch() *Stopwatch {
	sw := &Stopwatch{}
	sw.Reset()

	return sw
}

func (sw *Stopwatch) Start() {
	sw.start = time.Now()
	sw.last = sw.start
}

func (sw *Stopwatch) Stop() {
	sw.stop = time.Now()
}

func (sw *Stopwatch) Lap(label string) {
	lt := time.Now()

	sw.laps = append(sw.laps, Lap{
		Label:    label,
		Duration: lt.Sub(sw.last),
	})

	sw.last = lt
}

func (sw *Stopwatch) Result() time.Duration {
	return sw.stop.Sub(sw.start)
}

func (sw *Stopwatch) Laps() []Lap {
	return sw.laps
}

func (sw *Stopwatch) Reset() {
	sw.laps = make([]Lap, 0)
	sw.start = time.Time{}
	sw.stop = time.Time{}
}

func (sw *Stopwatch) String() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("Total time elapsed: %v\n", sw.Result()))

	for _, lap := range sw.Laps() {
		sb.WriteString(fmt.Sprintf("Lap: %v [%s]\n", lap.Duration, lap.Label))
	}

	return sb.String()
}
