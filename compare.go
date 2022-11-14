package main

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
	"gonum.org/v1/plot/plotter"
)

func CompareTests(tag string, sortTps, sortLatency bool, files ...string) error {
	title := fmt.Sprintf("compare_st-%v_sl-%v", sortTps, sortLatency)
	if len(tag) > 0 {
		title += fmt.Sprintf("_%s", tag)
	}

	linesTps := make([]Line, len(files))
	linesLatency := make([]Line, len(files))

	for i, file := range files {
		rs, e := countResultsFile(file, sortLatency, sortTps)
		if e != nil {
			logger.Error("countResultsFile err", zap.String("file", file), zap.String("err", e.Error()))
			return e
		}

		xysTPS := make(plotter.XYs, 0, len(rs.TPSes))
		for j, v := range rs.TPSes {
			xysTPS = append(xysTPS, plotter.XY{X: float64(j + 1), Y: v})
		}
		linesTps[i] = Line{name: file, xys: xysTPS}

		xysLatency := make(plotter.XYs, 0, len(rs.LatencySummary.Latencies))
		for j, v := range rs.LatencySummary.Latencies {
			xysLatency = append(xysLatency, plotter.XY{X: float64(j + 1), Y: v})
		}
		linesLatency[i] = Line{name: file, xys: xysLatency}
	}

	e := DrawLines(title, XLabel, TpsYLabel, filepath.Join(CompareTpsDir, title+PngSuffix), linesTps)
	if e != nil {
		return e
	}

	return DrawLines(title, XLabel, LatencyYLabel, filepath.Join(CompareLatencyDir, title+PngSuffix), linesLatency)
}
