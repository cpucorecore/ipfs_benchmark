package main

import (
	"time"
)

type Result struct {
	Gid         int
	Fid         int
	Cid         string
	Ret         int
	S           time.Time
	E           time.Time
	Latency     int64
	Concurrency int32
	Err         error  `json:"-"`
	Resp        string `json:"-"`
}

type ErrResult struct {
	R    Result
	Err  error
	Resp string
}
