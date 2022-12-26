package main

import (
	"errors"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const (
	ErrHttpClientDoFailed  = ErrCategoryHttp + 1
	ErrIOUtilReadAllFailed = ErrCategoryHttp + 2
	ErrCloseHttpResp       = ErrCategoryHttp + 3
	ErrReadHttpRespTimeout = ErrCategoryHttp + 4
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   300 * time.Second,
		KeepAlive: 1200 * time.Second,
	}).DialContext,
	MaxIdleConns:          1200,
	IdleConnTimeout:       600 * time.Second,
	ExpectContinueTimeout: 600 * time.Second,
	MaxIdleConnsPerHost:   300,
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
	resp, err := httpClient.Do(req)
	r.E = time.Now()

	if p.SyncConcurrency {
		atomic.AddInt32(&concurrency, -1)
	}

	r.Latency = r.E.Sub(r.S).Microseconds()

	if err != nil {
		r.Ret = ErrHttpClientDoFailed
		r.Err = err
		return r
	}

	r.HttpStatusCode = resp.StatusCode

	respBodyChan := make(chan string, 1)
	go func() {
		body, readAllErr := io.ReadAll(resp.Body)
		if readAllErr != nil {
			r.Ret = ErrIOUtilReadAllFailed
			r.Err = readAllErr
		}
		respBodyChan <- string(body)
	}()

	select {
	case <-time.After(time.Duration(p.ReadHttpRespTimeout) * time.Second):
		r.Ret = ErrReadHttpRespTimeout
		r.Err = errors.New("read http response timeout")
	case respBody := <-respBodyChan:
		if p.Verbose {
			logger.Debug("http response", zap.String("body", respBody))
		}

		if !dropHttpResp {
			r.Resp = respBody
		}
	}

	err = resp.Body.Close()
	if err != nil {
		r.Ret = ErrCloseHttpResp
		r.Err = err
	}

	return r
}
