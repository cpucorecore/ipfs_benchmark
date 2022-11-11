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
	//Error                 error  `json:"Error,omitempty"`
	Resp string // TODO impl
}

type ErrResult struct {
	Result
	Err error
}
