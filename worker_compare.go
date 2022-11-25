package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
	"gonum.org/v1/plot/plotter"
)

func CompareTests(input CompareParams, testFiles ...string) error {
	title := input.name() + "_" + input.info()

	linesTps := make([]Line, len(testFiles))
	linesLatency := make([]Line, len(testFiles))

	for i, testFile := range testFiles {
		rs, e := countResultsFile(testFile, input.SortTps, input.SortLatency)
		if e != nil {
			logger.Error("countResultsFile err", zap.String("testReport", testFile), zap.String("err", e.Error()))
			return e
		}

		if input.To > len(rs.TPSes) {
			input.To = len(rs.TPSes)
			if input.From >= input.To {
				return errors.New(fmt.Sprintf("wrong from, to range, from:%d, to:%d, max to:%d", from, to, input.To))
			}
		}

		xysTPS := make(plotter.XYs, 0, input.To-input.From)
		for j, v := range rs.TPSes[input.From:input.To] {
			xysTPS = append(xysTPS, plotter.XY{X: float64(j + 1), Y: v})
		}
		linesTps[i] = Line{name: testFile, xys: xysTPS}

		xysLatency := make(plotter.XYs, 0, input.To-input.From)
		for j, v := range rs.LatencySummary.Latencies[input.From:input.To] {
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
