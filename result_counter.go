package main

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"gonum.org/v1/plot/plotter"
)

type ResultsSummary struct {
	StartTime        time.Time
	EndTime          time.Time
	Samples          int
	Errs             int
	ErrPercentage    float32
	TPS              float64
	ErrCounter       map[int]int
	WindowTPSes      plotter.XYs
	Results          []Result
	LatenciesSummary LatenciesSummary
}

func countResults(wg *sync.WaitGroup) {
	time.Now().UnixMilli()
	defer wg.Done()

	rs := processResults(chResults, input.Goroutines, params.Window)
	outputSummary(rs)
}

func processResults(in <-chan Result, goroutines, window int) ResultsSummary {
	rs := ResultsSummary{
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		ErrCounter:  make(map[int]int),
		WindowTPSes: make(plotter.XYs, 0, 1000),
		Results:     make([]Result, 0, 10000),
	}

	latencies := make(plotter.Values, 0, 10000)
	var latency int64
	var intervalSuccessCount, intervalSumLatency int64
	for {
		r, ok := <-in
		if !ok {
			rs.LatenciesSummary = countLatencies(latencies)
			rs.TPS = float64(goroutines) * 1000 / (rs.LatenciesSummary.SumLatency / float64(rs.Samples-rs.Errs))
			break
		}

		rs.Samples++
		rs.Results = append(rs.Results, r)

		if !r.S.IsZero() && r.S.Before(rs.StartTime) {
			rs.StartTime = r.S
		}
		if !r.E.IsZero() && r.E.After(rs.EndTime) {
			rs.EndTime = r.E
		}

		if r.Ret != 0 {
			rs.Errs++
			rs.ErrCounter[r.Ret]++
		} else {
			intervalSuccessCount++

			latency = r.E.Sub(r.S).Milliseconds()
			intervalSumLatency += latency
			latencies = append(latencies, float64(latency))
		}

		if rs.Samples%window == 0 {
			tps := float64(goroutines*1000) / (float64(intervalSumLatency) / float64(intervalSuccessCount))
			logger.Info(
				"window summary",
				zap.Float64("seconds elapsed", time.Since(rs.StartTime).Seconds()),
				zap.Int("samples", rs.Samples),
				zap.Int("errs", rs.Errs),
				zap.Float64("tps", tps),
			)

			rs.WindowTPSes = append(rs.WindowTPSes, plotter.XY{X: float64(rs.Samples), Y: tps})
			intervalSuccessCount = 0
			intervalSumLatency = 0
		}
	}

	rs.ErrPercentage = float32(rs.Errs) / float32(rs.Samples)

	return rs
}