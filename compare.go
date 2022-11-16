package main

import (
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"gonum.org/v1/plot/plotter"
)

func CompareTests(tag string, sortTps, sortLatency, timestamp bool, testFiles ...string) error {
	title := fmt.Sprintf("compare_st-%v_sl-%v", sortTps, sortLatency)
	if len(tag) > 0 {
		title += fmt.Sprintf("_%s", tag)
	}
	if timestamp {
		title += fmt.Sprintf("_%d", time.Now().Unix())
	}

	linesTps := make([]Line, len(testFiles))
	linesLatency := make([]Line, len(testFiles))

	for i, testFile := range testFiles {
		rs, e := countResultsFile(testFile, sortLatency, sortTps)
		if e != nil {
			logger.Error("countResultsFile err", zap.String("testFile", testFile), zap.String("err", e.Error()))
			return e
		}

		xysTPS := make(plotter.XYs, 0, len(rs.TPSes))
		for j, v := range rs.TPSes {
			xysTPS = append(xysTPS, plotter.XY{X: float64(j + 1), Y: v})
		}
		linesTps[i] = Line{name: testFile, xys: xysTPS}

		xysLatency := make(plotter.XYs, 0, len(rs.LatencySummary.Latencies))
		for j, v := range rs.LatencySummary.Latencies {
			xysLatency = append(xysLatency, plotter.XY{X: float64(j + 1), Y: v})
		}
		linesLatency[i] = Line{name: testFile, xys: xysLatency}
	}

	e := DrawLines(title, XLabel, TpsYLabel, filepath.Join(CompareTpsDir, title+PngSuffix), linesTps)
	if e != nil {
		return e
	}

	return DrawLines(title, XLabel, LatencyYLabel, filepath.Join(CompareLatencyDir, title+PngSuffix), linesLatency)
}
