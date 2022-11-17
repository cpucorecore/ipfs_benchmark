package main

import (
	"sort"
)

func countResultsFile(file string, sortTps, sortLatency bool) (rs ResultsSummary, e error) {
	t, e := loadTest(file)
	if e != nil {
		return rs, e
	}

	in := make(chan Result, 10000)
	go func() {
		for _, r := range t.ResultsSummary.Results {
			in <- r
		}
		close(in)
	}()

	rs = processResults(in)

	if sortLatency {
		sort.Float64s(rs.LatencySummary.Latencies)
	}

	if sortTps {
		sort.Float64s(rs.TPSes)
	}

	return rs, nil
}
