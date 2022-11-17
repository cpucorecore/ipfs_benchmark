package main

import (
	"go.uber.org/zap"
	"gonum.org/v1/plot/plotter"
	"path/filepath"
)

func CompareTests(input CompareInput, testFiles ...string) error {
	name := input.name()

	linesTps := make([]Line, len(testFiles))
	linesLatency := make([]Line, len(testFiles))

	for i, testFile := range testFiles {
		rs, e := countResultsFile(testFile, input.SortTps, input.SortLatency)
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

	e := DrawLines(name, XLabel, TpsYLabel, filepath.Join(CompareTpsDir, name+PngSuffix), linesTps)
	if e != nil {
		return e
	}

	return DrawLines(name, XLabel, LatencyYLabel, filepath.Join(CompareLatencyDir, name+PngSuffix), linesLatency)
}
