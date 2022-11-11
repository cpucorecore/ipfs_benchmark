package main

import (
	"go.uber.org/zap"
	"path/filepath"
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
		rs.LatenciesSummary.LatenciesMicroseconds,
	)
	if e != nil {
		logger.Error("DrawValues err", zap.String("err", e.Error()))
	}

	e = DrawXYs(
		title,
		XLabel,
		TpsYLabel,
		filepath.Join(ImagesTpsDir, title+PngSuffix),
		rs.WindowTPSes,
	)
	if e != nil {
		logger.Error("DrawXYs err", zap.String("err", e.Error()))
	}

	saveTest(rs)
}
