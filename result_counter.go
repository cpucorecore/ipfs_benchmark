package main

import (
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
	"gonum.org/v1/plot/plotter"
)

const (
	MillisecondPerSecond      = 1000
	MicrosecondPerMillisecond = 1000
)

type ResultsSummary struct {
	StartTime              time.Time
	EndTime                time.Time
	Samples                int
	Errs                   int
	ErrPercentage          float32
	TPS                    float64
	ConcurrentReqNumberSum uint64
	ErrCounter             map[int]int
	WindowTPSes            plotter.XYs
	Results                []Result
	LatenciesSummary       LatenciesSummary
}

func countResults(wg *sync.WaitGroup) {
	time.Now().UnixMilli()
	defer wg.Done()

	rs := processResults(chResults, input.Goroutines, params.Window)
	outputSummary(rs)
}

func processResults(in <-chan Result, _, window int) ResultsSummary {
	rs := ResultsSummary{
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		ErrCounter:  make(map[int]int),
		WindowTPSes: make(plotter.XYs, 0, 1000),
		Results:     make([]Result, 0, 10000),
	}

	latencies := make(plotter.Values, 0, 10000)
	var intervalSuccessCount, intervalSumLatency int64
	for {
		r, ok := <-in
		if !ok {
			rs.LatenciesSummary = countLatencies(latencies)
			rs.TPS = 0
			if rs.LatenciesSummary.SumLatency > 0 && (rs.Samples-rs.Errs) > 0 {
				rs.TPS = (float64(rs.ConcurrentReqNumberSum) / float64(rs.Samples-rs.Errs)) *
					(float64(MillisecondPerSecond*MicrosecondPerMillisecond) /
						(rs.LatenciesSummary.SumLatency / float64(rs.Samples-rs.Errs)))
			}
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
			intervalSumLatency += r.LatenciesMicroseconds
			rs.ConcurrentReqNumberSum += uint64(r.CurrentReqNumber)
			latencies = append(latencies, float64(r.LatenciesMicroseconds))
		}

		if rs.Samples%window == 0 {
			var tps float64 = 0
			if intervalSumLatency > 0 && intervalSuccessCount > 0 {
				tps = float64(r.CurrentReqNumber) *
					(float64(MillisecondPerSecond*MicrosecondPerMillisecond) /
						(float64(intervalSumLatency) / float64(intervalSuccessCount)))
			}

			logger.Info(
				"window summary",
				zap.Float64("seconds elapsed", time.Since(rs.StartTime).Seconds()),
				zap.Int("samples", rs.Samples),
				zap.Int("errs", rs.Errs),
				zap.Float64("CurrentReqNumber", float64(r.CurrentReqNumber)),
				zap.Float64("tps", tps),
			)

			if !math.IsNaN(tps) {
				rs.WindowTPSes = append(rs.WindowTPSes, plotter.XY{X: float64(rs.Samples), Y: tps})
			}

			intervalSuccessCount = 0
			intervalSumLatency = 0
		}
	}

	rs.ErrPercentage = float32(rs.Errs) / float32(rs.Samples) * 100

	return rs
}
