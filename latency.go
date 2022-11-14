package main

import (
	"gonum.org/v1/plot/plotter"
)

type LatencySummary struct {
	Samples    int
	Min        float64
	Max        float64
	Mean       float64
	SumLatency float64
	Latencies  plotter.Values
}
