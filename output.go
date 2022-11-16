package main

import (
	"path/filepath"

	"go.uber.org/zap"
	"gonum.org/v1/plot/plotter"
)

const (
	XLabel        = "samples"
	TpsYLabel     = "tps"
	LatencyYLabel = "latency"

	PngSuffix = ".png"
)

func outputSummary(rs ResultsSummary) {
	title := input.info()
	e := DrawValues(
		title,
		XLabel,
		LatencyYLabel,
		filepath.Join(ImagesLatencyDir, title+PngSuffix),
		rs.LatencySummary.Latencies,
	)
	if e != nil {
		logger.Error("DrawValues err", zap.String("err", e.Error()))
	}

	xyz := make(plotter.XYs, 0, len(rs.TPSes))
	for i, tps := range rs.TPSes {
		xyz = append(xyz, plotter.XY{X: float64(i + 1), Y: tps})
	}

	e = DrawXYs(
		title,
		XLabel,
		TpsYLabel,
		filepath.Join(ImagesTpsDir, title+PngSuffix),
		xyz,
	) // TODO fix execute slowly
	if e != nil {
		logger.Error("DrawXYs err", zap.String("err", e.Error()))
	}

	saveTest(rs)
}
