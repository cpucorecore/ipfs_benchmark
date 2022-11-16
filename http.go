package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	ErrHttpClientDoFailed  = -101
	ErrIoutilReadAllFailed = -102
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   300 * time.Second,
		KeepAlive: 1200 * time.Second,
	}).DialContext,
	MaxIdleConns:          2000,
	IdleConnTimeout:       600 * time.Second,
	ExpectContinueTimeout: 600 * time.Second,
	MaxIdleConnsPerHost:   2000,
}

var httpClient = &http.Client{Transport: transport}

func doHttpRequest(req *http.Request) Result {
	var r Result

	if params.Sync {
		atomic.AddInt32(&concurrency, 1)
		r.Concurrency = concurrency
	} else {
		r.Concurrency = int32(input.Goroutines)
	}
	r.S = time.Now()

	resp, e := httpClient.Do(req)

	r.E = time.Now()
	if params.Sync {
		atomic.AddInt32(&concurrency, -1)
	}

	r.Latency = r.E.Sub(r.S).Microseconds()

	if e != nil {
		r.Ret = ErrHttpClientDoFailed
		r.Err = e
		return r
	}

	body, e := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if e != nil {
		r.Err = e
		r.Ret = ErrIoutilReadAllFailed
		return r
	}
	if !params.Drop {
		r.Resp = string(body)
	}

	return r
}
