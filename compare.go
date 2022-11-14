package main

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
	"gonum.org/v1/plot/plotter"
)

func CompareTests(tag string, sortLatency, sortTps bool, window int, files ...string) error {
	title := fmt.Sprintf("compare_%s_wd%d_sortLatency%v", tag, window, sortLatency)

	linesTps := make([]Line, len(files))
	linesLatency := make([]Line, len(files))

	for i, file := range files {
		rs, e := countResultsFile(file, sortLatency, sortTps)
		if e != nil {
			logger.Error("countResultsFile err", zap.String("file", file), zap.String("err", e.Error()))
			return e
		}

		linesTps[i] = Line{name: file, xys: rs.WindowTPSes}

		var xys plotter.XYs
		for j, v := range rs.LatenciesSummary.Latencies {
			xys = append(xys, plotter.XY{X: float64(j), Y: v})
		}
		linesLatency[i] = Line{name: file, xys: xys}
	}

	e := DrawLines(title, XLabel, TpsYLabel, filepath.Join(CompareTpsDir, title+PngSuffix), linesTps)
	if e != nil {
		return e
	}

	return DrawLines(title, XLabel, LatencyYLabel, filepath.Join(CompareLatencyDir, title+PngSuffix), linesLatency)
}
