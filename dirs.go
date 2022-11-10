package main

import (
	"os"
	"path/filepath"
)

const (
	PathTests    = "tests"
	PathReports  = "reports"
	PathErrs     = "errs"
	PathImages   = "images"
	PathTps      = "tps"
	PathLatency  = "latency"
	PathCompares = "compares"
)

var (
	ReportsDir       = filepath.Join(PathTests, PathReports)
	ErrsDir          = filepath.Join(PathTests, PathErrs)
	ImagesDir        = filepath.Join(PathTests, PathImages)
	ImagesTpsDir     = filepath.Join(ImagesDir, PathTps)
	ImagesLatencyDir = filepath.Join(ImagesDir, PathLatency)

	CompareTpsDir     = filepath.Join(PathCompares, PathTps)
	CompareLatencyDir = filepath.Join(PathCompares, PathLatency)
)

func initDirs() int {
	ec := 0

	e := os.MkdirAll(ReportsDir, os.ModePerm)
	if e != nil {
		ec++
	}

	e = os.MkdirAll(ErrsDir, os.ModePerm)
	if e != nil {
		ec++
	}

	e = os.MkdirAll(ImagesTpsDir, os.ModePerm)
	if e != nil {
		ec++
	}

	e = os.MkdirAll(ImagesLatencyDir, os.ModePerm)
	if e != nil {
		ec++
	}

	e = os.MkdirAll(CompareTpsDir, os.ModePerm)
	if e != nil {
		ec++
	}

	e = os.MkdirAll(CompareLatencyDir, os.ModePerm)
	if e != nil {
		ec++
	}

	return ec
}
