package main

import (
	"fmt"
	"time"
)

var _ ISampleInput = CompareInput{}

type CompareInput struct {
	SampleInput
	Tag         string
	Timestamp   bool
	SortTps     bool
	SortLatency bool
}

func (i CompareInput) name() string {
	n := fmt.Sprintf("compare_%s_%s", i.Tag, i.paramsStr())
	if i.Timestamp {
		n += fmt.Sprintf("_%d", time.Now().Unix())
	}
	return n
}

func (i CompareInput) check() bool {
	return true
}

func (i CompareInput) paramsStr() string {
	return fmt.Sprintf("sort_tps-%v_sort_latency-%v", i.SortTps, i.SortLatency)
}