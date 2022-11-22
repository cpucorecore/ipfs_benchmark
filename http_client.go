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
	ErrHttpBase            = 100
	ErrHttpClientDoFailed  = ErrHttpBase + 1
	ErrIoutilReadAllFailed = ErrHttpBase + 2
	ErrCloseHttpResp       = ErrHttpBase + 3
	ErrReadHttpRespTimeout = ErrHttpBase + 4
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

	if syncConcurrency {
		atomic.AddInt32(&concurrency, 1)
		r.Concurrency = concurrency
	} else {
		r.Concurrency = int32(goroutines)
	}
	r.S = time.Now()

	resp, e := httpClient.Do(req)

	r.E = time.Now()
	if syncConcurrency {
		atomic.AddInt32(&concurrency, -1)
	}

	r.Latency = r.E.Sub(r.S).Microseconds()

	if e != nil {
		r.Ret = ErrHttpClientDoFailed
		r.Err = e
		return r
	}

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
	case <-time.After(time.Duration(timeout) * time.Second):
		{
			r.Ret = ErrReadHttpRespTimeout
		}
	case respBody := <-respBodyChan:
		{
			if verbose {
				logger.Debug("http response", zap.String("body", r.Resp))
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
