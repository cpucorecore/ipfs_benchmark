package main

import (
	"time"
)

type Result struct {
	Gid                   int
	Fid                   int
	Cid                   string
	Ret                   int
	S                     time.Time
	E                     time.Time
	LatenciesMicroseconds int64
	Err                   error  `json:"-"`
	Resp                  string `json:"-"`
}

type ErrResult struct {
	R    Result
	Err  error
	Resp string
}
