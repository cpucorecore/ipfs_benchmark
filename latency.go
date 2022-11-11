package main

import (
	"gonum.org/v1/plot/plotter"
)

type LatenciesSummary struct {
	Quantity              int
	Min                   float64
	Max                   float64
	Mean                  float64
	SumLatency            float64
	LatenciesMicroseconds plotter.Values `json:"LatenciesMicroseconds,omitempty"`
}
