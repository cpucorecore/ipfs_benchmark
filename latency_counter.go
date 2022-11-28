package main

import (
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

func countLatencies(latencies []float64) LatencySummary {
	var ls LatencySummary

	ls.Samples = len(latencies)
	if ls.Samples == 0 {
		return ls
	}

	ls.Min = floats.Min(latencies)
	ls.Max = floats.Max(latencies)
	ls.Mean = stat.Mean(latencies, nil)

	ls.Latencies = latencies

	for _, l := range latencies {
		ls.SumLatency += l
	}

	return ls
}
