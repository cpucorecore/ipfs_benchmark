package main

import (
	"fmt"
)

var (
	verbose         bool
	goroutines      int
	syncConcurrency bool
)

func baseParamsStr() string {
	return fmt.Sprintf("v-%v_g-%d_sync-%v", verbose, goroutines, syncConcurrency)
}

var _ IHttpInput = HttpInput{}

type HttpInput struct {
	Sample           int
	Host             string
	Port             string
	Path             string
	Method           string
	DropHttpRespBody bool
}

func (i HttpInput) setSample(sample int) {
	i.Sample = sample
}

func (i HttpInput) getSample() int {
	return i.Sample
}

func (i HttpInput) name() string {
	return getNameByHttpMethodAndPath(i.Method, i.Path)
}

func (i HttpInput) check() bool {
	return true // TODO check
}

func (i HttpInput) paramsStr() string {
	return baseParamsStr() // TODO check
}

func (i HttpInput) host() string {
	return i.Host
}

func (i HttpInput) port() string {
	return i.Port
}

func (i HttpInput) method() string {
	return i.Method
}

func (i HttpInput) path() string {
	return i.Path
}

func (i HttpInput) baseUrl() string {
	return HTTP + i.Host + ":" + i.Port + i.Path
}

func (i HttpInput) dropHttpRespBody() bool {
	return shouldDropHttpRespBody(i.Path)
}
