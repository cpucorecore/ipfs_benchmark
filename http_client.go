package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const (
	ErrCreateRequest       = ErrCategoryHttp + 1
	ErrHttpClientDoFailed  = ErrCategoryHttp + 2
	ErrIoutilReadAllFailed = ErrCategoryHttp + 3
	ErrCloseHttpResp       = ErrCategoryHttp + 4
	ErrReadHttpRespTimeout = ErrCategoryHttp + 5
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

func doHttpRequest(req *http.Request, dropHttpResp bool) Result {
	var r Result

	if p.SyncConcurrency {
		atomic.AddInt32(&concurrency, 1)
		r.Concurrency = concurrency
	} else {
		r.Concurrency = int32(p.Goroutines)
	}

	r.S = time.Now()
	resp, e := httpClient.Do(req)
	r.E = time.Now()

	if p.SyncConcurrency {
		atomic.AddInt32(&concurrency, -1)
	}

	r.Latency = r.E.Sub(r.S).Microseconds()

	if e != nil {
		r.Ret = ErrHttpClientDoFailed
		r.Err = e
		return r
	}

	r.HttpStatusCode = resp.StatusCode

	respBodyChan := make(chan string, 1)
	go func() {
		body, readAllErr := ioutil.ReadAll(resp.Body)
		if readAllErr != nil {
			r.Ret = ErrIoutilReadAllFailed
			r.Err = readAllErr
		}
		respBodyChan <- string(body)
	}()

	select {
	case <-time.After(time.Duration(p.ReadHttpRespTimeout) * time.Second):
		{
			r.Ret = ErrReadHttpRespTimeout
		}
	case respBody := <-respBodyChan:
		{
			if p.Verbose {
				logger.Debug("http response", zap.String("body", respBody))
			}

			if !dropHttpResp {
				r.Resp = respBody
			}
		}
	}

	e = resp.Body.Close()
	if e != nil {
		r.Ret = ErrCloseHttpResp
		r.Err = e
	}

	return r
}
