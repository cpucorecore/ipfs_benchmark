package main

import (
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

func countLatencies(latencies []float64) LatenciesSummary {
	var ls LatenciesSummary

	ls.Quantity = len(latencies)
	if ls.Quantity == 0 {
		return ls
	}

	ls.Min = floats.Min(latencies)
	ls.Max = floats.Max(latencies)
	ls.Mean = stat.Mean(latencies, nil)

	ls.LatenciesMicroseconds = latencies

	for _, latency := range latencies {
		ls.SumLatency += latency
	}

	return ls
}
