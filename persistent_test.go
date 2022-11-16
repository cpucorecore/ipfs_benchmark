package main

import (
	"testing"
	"time"
)

var (
	rs = ResultsSummary{
		StartTime:     time.Time{},
		EndTime:       time.Time{},
		Samples:       100,
		Errs:          11,
		ErrPercentage: 0,
		TPS:           0,
		ErrCounter:    nil,
		Results: []Result{
			{
				Fid: 0,
				Ret: 0,
				S:   time.Time{},
				E:   time.Time{},
				Cid: "Cid1",
			},
			{
				Fid: 1,
				Ret: 0,
				S:   time.Time{},
				E:   time.Time{},
				Cid: "Cid2",
			},
		},
		LatencySummary: LatencySummary{
			Samples:    10,
			Min:        1,
			Max:        10,
			Mean:       4,
			SumLatency: 40,
		},
	}

	testInput = Input{
		TestCase:       "tc",
		Goroutines:     10,
		From:           0,
		To:             100,
		BlockSize:      1024 * 1024,
		ReplicationMin: 2,
		ReplicationMax: 2,
		HostPort:       "127.0.0.1:9094",
	}

	testParams = Params{
		Verbose:        false,
		TestResultFile: "",
		FilesDir:       "",
		FileSize:       0,
	}
)

func TestSaveTest(t *testing.T) {
	input = testInput
	params = testParams

	saveTest(rs)
}
