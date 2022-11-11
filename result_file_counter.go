package main

import (
	"sort"
)

func countResultsFile(file string, sortLatency bool, window int) (rs ResultsSummary, e error) {
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

	rs = processResults(in, t.Input.Goroutines, window)

	if sortLatency {
		sort.Float64s(rs.LatenciesSummary.LatenciesMicroseconds)
	}

	return rs, nil
}
