package main

import (
	"fmt"
	"net/url"
)

type NameInput interface {
	name() string
}
type Input interface {
	check() bool
	paramsStr() string
	samples() int
}

type LocalInput interface {
	NameInput
	Input
}

type HttpInput interface {
	NameInput
	Input
	baseUrl() string
	getMethod() string
}

// RepeatHttpInput
// url: 		const
// urlParams: 	const
// final url = baseUrl() + urlParams()
type RepeatHttpInput interface {
	HttpInput
	urlParams() string
	repeat() int
}

// IterUrlHttpInput
// url: 		iterative
// urlParams: 	const
// final url = baseUrl() + "/cid" + urlParams()
type IterUrlHttpInput interface {
	HttpInput
	urlParams() string
}

// IterParamsHttpInput
// url: 		const
// urlParams: 	iterative
// final url = baseUrl() + iterParams("cid")
type IterParamsHttpInput interface {
	HttpInput
	iterParams(it string) string
}

type BaseInput struct {
	Verbose         bool
	Goroutines      int
	SyncConcurrency bool
}

func (i BaseInput) check() bool {
	return i.Goroutines > 0
}

func (i BaseInput) paramsStr() string {
	return fmt.Sprintf("g-%d_sc-%v", i.Goroutines, i.SyncConcurrency)
}

type BaseHttpInput struct {
	BaseInput
	HostPort         string // [192.168.0.85:9094, 192.168.0.85:5001]
	Method           string // [POST, DELETE, GET]
	Path             string // [/api/v0/id, /api/v0/repo/stat, /pins/ipfs, /add...]
	DropHttpRespBody bool
}

func (i BaseHttpInput) name() string {
	return getNameByHttpMethodAndPath(i.Method, i.Path)
}

func (i BaseHttpInput) baseUrl() string {
	return HTTP + i.HostPort + i.Path
}

func (i BaseHttpInput) getMethod() string {
	return i.Method
}

type ClusterAddInput struct {
	BaseHttpInput
	From      int
	To        int
	Replica   int
	BlockSize int
	Pin       bool
}

func (i ClusterAddInput) samples() int {
	return i.To - i.From
}

func (i ClusterAddInput) repeat() int {
	// useless for cluster add
	//TODO implement me
	panic("implement me")
}

const (
	MinBlockSize = 1024 * 256
	MaxBlockSize = 1024 * 1024
)

var _ RepeatHttpInput = ClusterAddInput{}

func (i ClusterAddInput) check() bool {
	return i.BaseHttpInput.check() &&
		fromToCheck(i.From, i.To) &&
		i.Replica > 0 &&
		i.BlockSize >= MinBlockSize && i.BlockSize <= MaxBlockSize
}

func (i ClusterAddInput) paramsStr() string {
	return fmt.Sprintf("%s_%s_replica-%d_block_size-%d_pin-%v",
		i.BaseHttpInput.paramsStr(), fromToParamsStr(i.From, i.To), i.Replica, i.BlockSize, i.Pin)
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

var _ IterUrlHttpInput = ClusterPinInput{}

type ClusterPinInput struct { // for ClusterPinRm and ClusterPinGet
	BaseHttpInput
	TestFile string
}

func (i ClusterPinInput) samples() int {
	return 0
}

func (i ClusterPinInput) urlParams() string {
	return ""
}

var _ RepeatHttpInput = IpfsRepeatInput{}

type IpfsRepeatInput struct {
	BaseHttpInput
	Repeat int
}

func (i IpfsRepeatInput) samples() int {
	return i.Goroutines * i.Repeat
}

func (i IpfsRepeatInput) repeat() int {
	return i.Repeat
}

func (i IpfsRepeatInput) urlParams() string {
	if pf, ok := repetitive[i.Path]; ok {
		return pf()
	}
	return ""
}

var _ IterParamsHttpInput = IpfsIterInput{}

type IpfsIterInput struct {
	BaseHttpInput
	TestFile string
}

func (i IpfsIterInput) samples() int {
	return 0
}

func (i IpfsIterInput) iterParams(it string) string {
	if pf, ok := iterative[i.Path]; ok {
		return pf(it)
	}
	return ""
}
