package main

import (
	"fmt"
	"strings"
)

type Range struct {
	From, To int
}

func (r Range) info() string {
	return fmt.Sprintf("from%d_to%d", r.From, r.To)
}

func (r Range) check() bool {
	return r.From >= 0 && r.To >= 0 && r.To >= r.From
}

type CompareParams struct {
	Tag                  string
	SortTps, SortLatency bool
}

func (p CompareParams) name() string {
	return "compare"
}

func (p CompareParams) info() string {
	return fmt.Sprintf("%s_st-%v_sl-%v", p.Tag, p.SortTps, p.SortLatency)
}

func (p CompareParams) check() bool {
	return true
}

type Params struct {
	Verbose         bool
	Goroutines      int
	SyncConcurrency bool
}

func (p Params) info() string {
	return fmt.Sprintf("g%d_sc-%v", p.Goroutines, p.SyncConcurrency)
}

func (p Params) check() bool {
	return p.Goroutines > 0
}

type GenFileParams struct {
	Params
	Range
	Size int
}

func (p GenFileParams) name() string {
	return "gen_file"
}

func (p GenFileParams) info() string {
	return fmt.Sprintf("%s_%s_%d", p.Params.info(), p.Range.info(), p.Size)
}

const MinFileSize = 1024 * 256

func (p GenFileParams) check() bool {
	return p.Params.check() && p.Range.check() && p.Size >= MinFileSize
}

type HttpParams struct {
	Params
	Host, Port, Method, Path string
	Timeout, MaxRetry        int
	DropHttpResp             bool
	Tag                      string
}

func baseUrl() string {
	return "http://" + p.Host + ":" + p.Port + p.Path
}

func (p HttpParams) name() string {
	return getNameByHttpMethodAndPath(p.Method, p.Path)
}

func (p HttpParams) info() string {
	return p.Params.info()
}

func (p HttpParams) check() bool {
	if len(p.Path) > 0 {
		if strings.Count(p.Path, "/api/v0") > 0 && p.Port != "5001" {
			logger.Warn("default ipfs api port is 5001s")
		}
	}

	return p.Params.check() &&
		len(p.Host) > 0 && len(p.Port) > 0 && len(p.Method) > 0 && len(p.Path) > 0 && p.Path[0] == '/' && p.Timeout > 0
}

type RepeatHttpParams struct {
	HttpParams
	Repeat int
}

func (p RepeatHttpParams) info() string {
	return fmt.Sprintf("%s_repeat%d", p.HttpParams.info(), p.Repeat)
}

func (p RepeatHttpParams) check() bool {
	return p.HttpParams.check() && p.Repeat > 0
}

type IterHttpParams struct {
	HttpParams
	TestReport string
	Range
}

func (p IterHttpParams) info() string {
	return fmt.Sprintf("%s_%s", p.HttpParams.info(), p.Range.info())
}

func (p IterHttpParams) check() bool {
	return p.HttpParams.check() && p.Range.check() && len(p.TestReport) > 0 // TODO check file exist
}
