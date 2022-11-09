package main

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/plot/plotter"
)

type Result struct {
	Gid   int // goroutine id
	Fid   int // file id
	Ret   int
	S     time.Time
	E     time.Time
	Cid   string
	Error error
}

type LatenciesSummary struct {
	Min        float64
	Max        float64
	Mean       float64
	SumLatency int64
	Latencies  plotter.Values
}

type ResultsSummary struct {
	S            time.Time
	E            time.Time
	Samples      int
	Errs         int
	TpsByLatency float64
	ErrCount     map[int]int
	TpsInfos     plotter.XYs
	Results      []Result
}

func analyseLatencies(values []float64) (min float64, max float64, mean float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	min = floats.Min(values)
	max = floats.Max(values)
	mean = stat.Mean(values, nil)
	return
}

func processResults(wg *sync.WaitGroup) {
	defer wg.Done()

	ls, rs := analyseResults(chResults, input.Goroutines, params.Window)
	outputLatenciesSummary(ls)
	outputResultsSummary(rs)
}

func analyseResultsFile(name string, sortLatency bool, window int) (LatenciesSummary, ResultsSummary, Input, error) {
	t, e := loadTest(name)
	if e != nil {
		logger.Error("loadTestResult err", zap.String("file", name), zap.String("err", e.Error()))
		return LatenciesSummary{}, ResultsSummary{}, Input{}, e
	}

	in := make(chan Result, 10000)
	go func() {
		for _, r := range t.Results {
			in <- r
		}
		close(in)
	}()

	ls, rs := analyseResults(in, t.Input.Goroutines, window)

	if sortLatency {
		sort.Float64s(ls.Latencies)
	}

	return ls, rs, t.Input, nil
}

func analyseResults(in <-chan Result, goroutines, window int) (LatenciesSummary, ResultsSummary) {
	ls := LatenciesSummary{
		Latencies: make(plotter.Values, 0, 10000),
	}

	rs := ResultsSummary{
		S:        time.Now(),
		E:        time.Now(),
		ErrCount: make(map[int]int),
		TpsInfos: make(plotter.XYs, 0, 1000),
		Results:  make([]Result, 0, 10000),
	}

	var latency, intervalSuccess, intervalSumLatency int64
	for {

		r, ok := <-in
		if !ok {
			rs.TpsByLatency = float64(goroutines) * 1000 / (float64(ls.SumLatency) / float64(rs.Samples-rs.Errs))
			break
		}

		rs.Samples++
		rs.Results = append(rs.Results, r)

		if !r.S.IsZero() && r.S.Before(rs.S) {
			rs.S = r.S
		}
		if !r.E.IsZero() && r.E.After(rs.E) {
			rs.E = r.E
		}

		if r.Ret != 0 {
			rs.Errs++
			rs.ErrCount[r.Ret]++
		} else {
			latency = r.E.Sub(r.S).Milliseconds()

			intervalSuccess++
			intervalSumLatency += latency

			ls.SumLatency += latency
			ls.Latencies = append(ls.Latencies, float64(latency))
		}

		if rs.Samples%window == 0 {
			tps := float64(goroutines*1000) / (float64(intervalSumLatency) / float64(intervalSuccess))
			logger.Info(
				"window summary",
				zap.Float64("seconds elapsed", time.Since(rs.S).Seconds()),
				zap.Int("samples", rs.Samples),
				zap.Int("errs", rs.Errs),
				zap.Float64("tps", tps))

			rs.TpsInfos = append(rs.TpsInfos, plotter.XY{X: float64(rs.Samples), Y: tps})
			intervalSuccess = 0
			intervalSumLatency = 0
		}
	}

	ls.Min, ls.Max, ls.Mean = analyseLatencies(ls.Latencies)

	return ls, rs
}

func outputLatenciesSummary(ls LatenciesSummary) {
	logger.Info("LatenciesSummary",
		zap.Float64("Min", ls.Min),
		zap.Float64("Max", ls.Max),
		zap.Float64("Mean", ls.Mean),
	)

	title := input.info(params.Timestamp)
	e := DrawValues(title+".latency", "samples", "latency", ls.Latencies)
	if e != nil {
		logger.Error("DrawValues err", zap.String("err", e.Error()))
	}
}

func outputResultsSummary(rs ResultsSummary) {
	b, e := json.Marshal(rs.ErrCount)
	if e != nil {
		logger.Error(e.Error())
	}

	logger.Info("ResultsSummary",
		zap.String("StartTime", rs.S.String()),
		zap.String("EndTime", rs.E.String()),
		zap.Int("Samples", rs.Samples),
		zap.Int("Errs", rs.Errs),
		zap.Float64("TpsByLatency", rs.TpsByLatency),
		zap.String("ErrCount", string(b)),
	)

	title := input.info(params.Timestamp)
	e = saveTest(Test{Params: params, Input: input, Results: rs.Results}, params.Timestamp)
	if e != nil {
		logger.Error("saveTest err", zap.String("err", e.Error()))
	}

	e = DrawXYs(title+".tps", "samples", "tps", rs.TpsInfos)
	if e != nil {
		logger.Error("DrawXYs err", zap.String("err", e.Error()))
	}
}
