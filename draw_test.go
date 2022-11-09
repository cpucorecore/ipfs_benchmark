package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot/plotter"
)

func TestDrawValues(t *testing.T) {
	e := DrawValues("TestDrawValues", "x", "y", plotter.Values{
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
	e := DrawXYs("TestDrawXYs", "x", "y", plotter.XYs{
		plotter.XY{X: 1, Y: 1},
		plotter.XY{X: 2, Y: 2},
		plotter.XY{X: 3, Y: 3},
		plotter.XY{X: 5, Y: 4},
		plotter.XY{X: 10, Y: 5},
	})

	require.Nil(t, e)
}
