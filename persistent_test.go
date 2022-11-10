package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	test = Test{
		Input: Input{
			TestCase:       "tc",
			Goroutines:     10,
			From:           0,
			To:             100,
			BlockSize:      1024 * 1024,
			ReplicationMin: 2,
			ReplicationMax: 2,
			HostPort:       "127.0.0.1:9094",
		},
		ResultsSummary: ResultsSummary{
			StartTime:     time.Time{},
			EndTime:       time.Time{},
			Samples:       100,
			Errs:          11,
			ErrPercentage: 0,
			TPS:           0,
			ErrCounter:    nil,
			WindowTPSes:   nil,
			Results: []Result{
				{
					Gid: 0,
					Fid: 0,
					Ret: 0,
					S:   time.Time{},
					E:   time.Time{},
					Cid: "Cid1",
				},
				{
					Gid: 1,
					Fid: 1,
					Ret: 0,
					S:   time.Time{},
					E:   time.Time{},
					Cid: "Cid2",
				},
			},
		},
	}
)

func TestSaveTest(t *testing.T) {
	e := saveTest(".", test, false)
	require.Nil(t, e)
	tst, e := loadTest(test.Input.info(false) + ".json")
	require.Nil(t, e)
	require.Equal(t, test, tst)
}
