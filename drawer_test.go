package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot/plotter"
)

const (
	TestTitle      = "t"
	TestXLabel     = "x"
	TestYLabel     = "y"
	DrawValuesFile = "f1" // TODO remove "f1" file in teardown
	DrawXYsFile    = "f2" // TODO remove "f2" file in teardown
	DrawLinesFile  = "f3" // TODO remove "f3" file in teardown
)

func TestDrawValues(t *testing.T) {
	e := DrawValues(TestTitle, TestXLabel, TestYLabel, DrawValuesFile, plotter.Values{
		1,
		2,
		3,
		4,
		5,
		3,
		1,
		8,
	})

	require.Nil(t, e)
}

func TestDrawXYs(t *testing.T) {
	e := DrawXYs(TestTitle, TestXLabel, TestYLabel, DrawXYsFile, plotter.XYs{
		plotter.XY{X: 1, Y: 1},
		plotter.XY{X: 2, Y: 2},
		plotter.XY{X: 3, Y: 3},
		plotter.XY{X: 5, Y: 4},
		plotter.XY{X: 10, Y: 5},
	})

	require.Nil(t, e)
}

func TestDrawLines(t *testing.T) {
	e := DrawLines(TestTitle, TestXLabel, TestYLabel, DrawLinesFile,
		[]Line{
			{"line1",
				plotter.XYs{
					plotter.XY{
						X: 1,
						Y: 1,
					},
					plotter.XY{
						X: 2,
						Y: 2,
					},
					plotter.XY{
						X: 3,
						Y: 3,
					},
				},
			},
			{"line2",
				plotter.XYs{
					plotter.XY{
						X: 1,
						Y: 4,
					},
					plotter.XY{
						X: 2,
						Y: 1,
					},
					plotter.XY{
						X: 3,
						Y: 8,
					},
				},
			},
		},
	)
	require.Nil(t, e)
}
