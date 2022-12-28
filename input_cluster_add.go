package main

import (
	"fmt"
	"net/url"
)

const (
	MinBlockSize = 1024 * 256
	MaxBlockSize = 1024 * 1024
)

type ClusterAddInput struct {
	HttpParams
	Range
	FileBufferSize int
	BlockSize      int
	Replica        int
	Pin            bool
}

func (i ClusterAddInput) name() string {
	return "cluster_add"
}

func (i ClusterAddInput) info() string {
	return fmt.Sprintf("%s_%s_bs%d_replica%d_pin-%v_%s", i.HttpParams.info(), i.Range.info(), i.BlockSize, i.Replica, i.Pin, i.Tag)
}

func (i ClusterAddInput) check() bool {
	return i.HttpParams.check() && i.Range.check() && fileBufferSize > 0 && i.BlockSize >= MinBlockSize && i.BlockSize <= MaxBlockSize && i.Replica > 0
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
		"stream-channels": {"false"},
	}

	return "?" + values.Encode()
}
