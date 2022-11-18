package main

import (
	"fmt"
	"net/url"
)

type ClusterAddInput struct {
	IterHttpParams
	BlockSize int
	Replica   int
	Pin       bool
}

const (
	MinBlockSize = 1024 * 256
	MaxBlockSize = 1024 * 1024
)

func (i ClusterAddInput) info() string {
	return fmt.Sprintf("%s_bs%d_replica%d_pin-%v", i.IterHttpParams.info(), i.BlockSize, i.Replica, i.Pin)
}

func (i ClusterAddInput) check() bool {
	return i.IterHttpParams.check() && i.BlockSize >= MinBlockSize && i.BlockSize <= MaxBlockSize && i.Replica > 0
}

func (i ClusterAddInput) paramsUrl() string {
	chunker := fmt.Sprintf("size-%d", i.BlockSize)
	noPin := fmt.Sprintf("%t", !i.Pin)
	min := fmt.Sprintf("%d", i.Replica)
	max := fmt.Sprintf("%d", i.Replica)
	values := url.Values{
		"chunker":         {chunker},
		"cid-version":     {"0"},
		"format":          {"unixfs"},
		"local":           {"false"},
		"mode":            {"recursive"},
		"no-pin":          {noPin},
		"replication-min": {min},
		"replication-max": {max},
	}

	return "?" + values.Encode()
}
