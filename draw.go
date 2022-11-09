package main

import (
	"fmt"
	"image/color"
	"path/filepath"

	"go.uber.org/zap"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func DrawValues(title, xLabel, yLabel string, values plotter.Values) error {
	if len(values) == 0 {
		return nil
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	bar, e := plotter.NewBarChart(values, 1)
	if e != nil {
		return e
	}

	p.Add(bar)

	return p.Save(30*vg.Inch, 10*vg.Inch, filepath.Join(ImagesDir, title+".png"))
}

func DrawXYs(title, xLabel, yLabel string, xys plotter.XYs) error {
	if len(xys) == 0 {
		return nil
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	if e := plotutil.AddLinePoints(p, xys); e != nil {
		return e
	}

	return p.Save(30*vg.Inch, 10*vg.Inch, filepath.Join(ImagesDir, title+".png"))
}

var colors = []color.RGBA{
	{R: 255, G: 0, B: 0, A: 255},
	{R: 0, G: 255, B: 0, A: 255},
	{R: 0, G: 0, B: 255, A: 255},
	{R: 204, G: 102, B: 0, A: 255},
	{R: 255, G: 0, B: 255, A: 255},
	{R: 0, G: 255, B: 255, A: 255},
	{R: 128, G: 255, B: 0, A: 255},
	{R: 0, G: 128, B: 128, A: 255},
	{R: 100, G: 200, B: 0, A: 255},
	{R: 100, G: 128, B: 30, A: 255},
}

func Compare(tag string, sortLatency bool, window int, files ...string) error {
	title := fmt.Sprintf("compare_%s_sortLatency%v", tag, sortLatency)

	tpsPlot := plot.New()
	tpsPlot.Title.Text = title + ".tps"
	tpsPlot.Legend.Top = true

	latencyPlot := plot.New()
	latencyPlot.Title.Text = title + ".latency"
	latencyPlot.Legend.Top = true

	for i, file := range files {
		ls, rs, _, e := analyseResultsFile(file, sortLatency, window)
		if e != nil {
			logger.Error("analyseResultsFile err", zap.String("file", file), zap.String("err", e.Error()))
			return e
		}

		// tps line
		tpsLine, tpsScatter, e := plotter.NewLinePoints(rs.TpsInfos)
		if e != nil {
			logger.Error("NewLinePoints", zap.String("err", e.Error()))
			return e
		}

		tpsLine.LineStyle.Color = colors[i%len(colors)]
		tpsLine.LineStyle.Width = 1
		tpsPlot.Add(tpsLine)
		tpsPlot.Legend.Add(file, tpsLine, tpsScatter)

		// latency line
		var xys plotter.XYs
		for j, v := range ls.Latencies {
			xys = append(xys, plotter.XY{
				X: float64(j + 1),
				Y: v,
			})
		}
		latencyLine, latencyScatter, e := plotter.NewLinePoints(xys)
		if e != nil {
			logger.Error("NewLinePoints", zap.String("err", e.Error()))
			return e
		}

		latencyLine.LineStyle.Color = colors[i%len(colors)]
		latencyLine.LineStyle.Width = 1
		latencyPlot.Add(latencyLine)
		latencyPlot.Legend.Add(file, latencyLine, latencyScatter)
	}

	if e := tpsPlot.Save(30*vg.Inch, 10*vg.Inch, filepath.Join(CompareImagesDir, title+".tps.png")); e != nil {
		logger.Error("plotter save err", zap.String("err", e.Error()))
		return e
	}

	if e := latencyPlot.Save(30*vg.Inch, 10*vg.Inch, filepath.Join(CompareImagesDir, title+".latency.png")); e != nil {
		logger.Error("plotter save err", zap.String("err", e.Error()))
		return e
	}

	return nil
}
