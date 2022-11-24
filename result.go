package main

import (
	"time"
)

type Result struct {
	Fid            int
	Cid            string
	Ret            int
	S              time.Time
	E              time.Time
	Latency        int64
	Concurrency    int32
	HttpStatusCode int
	Err            error  `json:"-"`
	Resp           string `json:"-"`
}

type ErrResult struct {
	R      Result
	ErrMsg string
	Resp   string
}
