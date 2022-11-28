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
	name := iInput.name() + "_" + iInput.info()
	e := DrawValues(
		name,
		XLabel,
		LatencyYLabel,
		filepath.Join(ImagesLatencyDir, name+PngSuffix),
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
		name,
		XLabel,
		TpsYLabel,
		filepath.Join(ImagesTpsDir, name+PngSuffix),
		xyz,
	) // TODO fix execute slowly when points > 10000
	if e != nil {
		logger.Error("DrawXYs err", zap.String("err", e.Error()))
	}

	saveTest(rs)
}
