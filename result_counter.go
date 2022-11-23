package main

import (
	"encoding/json"
	"fmt"
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
	StartTime     time.Time
	EndTime       time.Time
	Samples       int
	Errs          int
	ErrPercentage float32
	ErrCounter    map[int]int

	TPS            float64
	ConcurrencyAvg float64
	ConcurrencySum uint64
	TPSes          plotter.Values

	LatencySummary LatencySummary

	Results []Result
}

func countResults(wg *sync.WaitGroup) {
	defer wg.Done()

	rs := processResults(chResults)
	outputSummary(rs)
}

func processResults(in <-chan Result) ResultsSummary {
	rs := ResultsSummary{
		StartTime:  time.Now(),
		EndTime:    time.Now(),
		ErrCounter: make(map[int]int),
		TPSes:      make(plotter.Values, 0, 10000),
		Results:    make([]Result, 0, 10000),
	}

	latencies := make(plotter.Values, 0, 10000)

	var samples int
	if repeat > 0 {
		samples = p.Goroutines * repeat
	} else {
		samples = to - from
	}

	window := samples / 100
	if window == 0 {
		window = 10
	}

	cnt := 0
	for {
		r, ok := <-in
		if !ok {
			break
		}

		cnt++
		if cnt%window == 0 {
			errsInfo, _ := json.Marshal(rs.ErrCounter)
			logger.Info(fmt.Sprintf("progress:%d/%d %g%%, concurrency:%d, errs:%d, ErrCounter:%s",
				cnt,
				samples,
				float32(cnt*100)/float32(samples),
				r.Concurrency,
				rs.Errs,
				string(errsInfo)),
			)
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
		} else if len(p.Path) > 0 && r.HttpStatusCode != 200 {
			rs.Errs++
			rs.ErrCounter[r.Ret]++

			logger.Debug(fmt.Sprintf("result: %+v", r))
		} else {
			rs.ConcurrencySum += uint64(r.Concurrency)
			rs.TPSes = append(rs.TPSes, float64(r.Concurrency)*(float64(MillisecondPerSecond*MicrosecondPerMillisecond)/float64(r.Latency)))
			latencies = append(latencies, float64(r.Latency))
		}

		if p.Verbose {
			if r.Ret != 0 {
			} else {
				logger.Debug(
					"req summary",
					zap.Float64("seconds elapsed", time.Since(rs.StartTime).Seconds()),
					zap.Int("samples", rs.Samples),
					zap.Int("errs", rs.Errs),
					zap.Float64("Concurrency", float64(r.Concurrency)),
					zap.Float64("TPS", rs.TPSes[len(rs.TPSes)-1]),
				)
			}
		}
	}

	rs.LatencySummary = countLatencies(latencies)

	samplesSuccess := rs.Samples - rs.Errs
	if samplesSuccess > 0 {
		rs.ConcurrencyAvg = float64(rs.ConcurrencySum) / float64(samplesSuccess)
		tpsAvg := float64(MillisecondPerSecond*MicrosecondPerMillisecond) / (rs.LatencySummary.SumLatency / float64(samplesSuccess))
		rs.TPS = rs.ConcurrencyAvg * tpsAvg
	}
	rs.ErrPercentage = float32(rs.Errs) / float32(rs.Samples) * 100

	return rs
}
