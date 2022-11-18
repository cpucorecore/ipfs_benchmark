package main

import (
	"errors"
	"fmt"
	"net/url"
)

var _ IConstHttpInput = ClusterAddInput{}

type ClusterAddInput struct {
	HttpInput
	From      int
	To        int
	BlockSize int
	Replica   int
	Pin       bool
}

const (
	MinBlockSize = 1024 * 256
	MaxBlockSize = 1024 * 1024
)

func (i ClusterAddInput) check() error {
	if i.BlockSize < MinBlockSize || i.BlockSize > MaxBlockSize {
		return errors.New(fmt.Sprintf("wrong block size, must >= %d and <= %d", MinBlockSize, MaxBlockSize))
	}

	if i.Replica <= 0 {
		return errors.New("wrong replica, replica must > 0")
	}

	return checkFromTo(i.From, i.To)
}

func (i ClusterAddInput) paramsStr() string {
	return fmt.Sprintf("%s_%s_block_size-%d_replica-%d_pin-%v",
		baseParamsStr(), fromToParamsStr(i.From, i.To), i.BlockSize, i.Replica, i.Pin)
}

func (i ClusterAddInput) urlParams() string {
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
