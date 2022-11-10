package main

import (
	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func DrawValues(title, xLabel, yLabel, file string, values plotter.Values) error {
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

	return p.Save(30*vg.Inch, 10*vg.Inch, file)
}

func DrawXYs(title, xLabel, yLabel, file string, xys plotter.XYs) error {
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

	return p.Save(30*vg.Inch, 10*vg.Inch, file)
}

var colors = []color.RGBA{ // TODO load from cfg file
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

type Line struct {
	name string
	xys  plotter.XYs
}

func DrawLines(title, xLabel, yLabel, file string, lines []Line) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel
	p.Legend.Top = true

	for i, line := range lines {
		l, s, e := plotter.NewLinePoints(line.xys)
		if e != nil {
			return e
		}

		l.LineStyle.Color = colors[i%len(colors)]
		l.LineStyle.Width = 1
		p.Add(l)
		p.Legend.Add(line.name, l, s)
	}

	return p.Save(30*vg.Inch, 10*vg.Inch, file)
}
