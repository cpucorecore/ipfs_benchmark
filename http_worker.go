package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/buger/jsonparser"
	"go.uber.org/zap"
)

type Fid2Cid struct {
	Fid int
	Cid string
}

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
var concurrency int32

func doHttpRequest(req *http.Request) (time.Time, time.Time, int32, error, string) {
	currentConcurrency := int32(input.Goroutines)
	if params.Sync {
		atomic.AddInt32(&concurrency, 1)
		currentConcurrency = concurrency
	}

	startTime := time.Now()
	resp, e := httpClient.Do(req)
	endTime := time.Now()

	if params.Sync {
		atomic.AddInt32(&concurrency, -1)
	}

	if e != nil {
		if params.Verbose {
			logger.Info("httpClient do err", zap.String("err", e.Error()))
		}
		return startTime, endTime, currentConcurrency, e, ""
	}

	respBody, e := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if e != nil {
		if params.Verbose {
			logger.Error("ioutil ReadAll err", zap.String("err", e.Error()))
		}
		return startTime, endTime, currentConcurrency, e, string(respBody)
	}

	return startTime, endTime, currentConcurrency, nil, string(respBody)
}

func doIPFSRequest(gid int, method, baseUrl string, fid2cid Fid2Cid, paramsStr string) {
	r := Result{
		Gid: gid,
		Fid: fid2cid.Fid,
		Cid: fid2cid.Cid,
	}

	url := baseUrl + fid2cid.Cid + paramsStr
	if params.Verbose {
		logger.Debug("request url", zap.String("url", url))
	}

	req, _ := http.NewRequest(method, url, nil)

	r.S, r.E, r.Concurrency, r.Err, r.Resp = doHttpRequest(req)
	r.Latency = r.E.Sub(r.S).Microseconds()
	if r.Err != nil {
		r.Ret = -6
		r.ErrMsg = r.Err.Error()
		chResults <- r
		return
	}

	if params.Verbose {
		logger.Debug("http response", zap.String("body", r.Resp))
	}

	chResults <- r
}

func doRequestRepeat(method, path string, pf func(string) string, repeat int) error {
	url := "http://" + input.HostPort + path + pf("")
	if params.Verbose {
		logger.Debug(url)
	}

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(input.Goroutines)
	for i := 0; i < input.Goroutines; i++ {
		go func(gid int) {
			defer wg.Done()

			req, _ := http.NewRequest(method, url, nil)

			c := 0
			for c < repeat {
				c++
				r := Result{Gid: gid}

				r.S, r.E, r.Concurrency, r.Err, r.Resp = doHttpRequest(req)
				r.Latency = r.E.Sub(r.S).Microseconds()

				if params.Verbose {
					logger.Debug("http response", zap.String("body", r.Resp))
				}

				if r.Err != nil {
					r.Ret = -1
					r.ErrMsg = r.Err.Error()
					chResults <- r
					continue
				}

				chResults <- r
			}
		}(i)
	}

	wg.Wait()
	close(chResults)

	countResultsWg.Wait()
	return nil
}

func doRequestIter(method, path string, pf func(cid string) string) error {
	baseUrl := "http://" + input.HostPort + path

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(input.Goroutines)
	for i := 0; i < input.Goroutines; i++ {
		go func(gid int) {
			defer wg.Done()

			for {
				it, ok := <-chFid2Cids
				if !ok {
					break
				}

				url := baseUrl
				if pf == nil {
					url += "/" + it.Cid
				} else {
					url += pf(it.Cid)

				}

				if params.Verbose {
					logger.Debug(url)
				}

				req, _ := http.NewRequest(method, url, nil)

				r := Result{Gid: gid}

				r.S, r.E, r.Concurrency, r.Err, r.Resp = doHttpRequest(req)
				r.Latency = r.E.Sub(r.S).Microseconds()

				if params.Verbose {
					logger.Debug("http response", zap.String("body", r.Resp))
				}

				if r.Err != nil {
					r.Ret = -1
					r.ErrMsg = r.Err.Error()
					chResults <- r
					continue
				}

				chResults <- r
			}
		}(i)
	}

	wg.Wait()
	close(chResults)

	countResultsWg.Wait()
	return nil
}

func doIPFSRequests(method, path string, pf func() string) error {
	baseUrl := "http://" + input.HostPort + path

	var prsWg sync.WaitGroup
	prsWg.Add(1)
	go countResults(&prsWg)

	paramsStr := pf()
	var wg sync.WaitGroup
	wg.Add(input.Goroutines)
	for i := 0; i < input.Goroutines; i++ {
		go func(gid int) {
			defer wg.Done()

			for {
				cid, ok := <-chFid2Cids
				if !ok {
					return
				}

				doIPFSRequest(gid, method, baseUrl, cid, paramsStr)
			}
		}(i)
	}

	wg.Wait()
	close(chResults)

	prsWg.Wait()
	return nil
}

var urlAdd string

func postFile(tid int, fid int) {
	r := Result{
		Gid: tid,
		Fid: fid,
		Ret: 0,
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fidStr := fmt.Sprintf("%d", fid)
	ff, e := w.CreateFormFile("file", fidStr)
	if e != nil {
		r.Ret = -1
		r.Err = e
		r.ErrMsg = r.Err.Error()
		chResults <- r
		return
	}

	fp := filepath.Join(params.FilesDir, fidStr)
	f, e := os.Open(fp)
	if e != nil {
		r.Ret = -2
		r.Err = e
		r.ErrMsg = r.Err.Error()
		chResults <- r
		return
	}
	defer f.Close()

	n, e := io.Copy(ff, f)
	if e != nil {
		logger.Error("read file err", zap.String("err", e.Error()), zap.Int64("read bytes", n))
		r.Ret = -3
		r.Err = e
		r.ErrMsg = r.Err.Error()
		chResults <- r
		return
	}

	e = w.Close()
	if e != nil {
		r.Ret = -4
		r.Err = e
		r.ErrMsg = r.Err.Error()
		chResults <- r
		return
	}

	req, e := http.NewRequest(http.MethodPost, urlAdd, &b)
	if e != nil {
		r.Ret = -5
		r.Err = e
		r.ErrMsg = r.Err.Error()
		chResults <- r
		return
	}

	req.Header.Add("Content-Type", w.FormDataContentType())

	r.S, r.E, r.Concurrency, r.Err, r.Resp = doHttpRequest(req)
	r.Latency = r.E.Sub(r.S).Microseconds()
	if r.Err != nil {
		logger.Error("doHttpRequest err", zap.String("err", r.Err.Error()))
		r.Ret = -6
		r.ErrMsg = r.Err.Error()
		chResults <- r
		return
	}

	if params.Verbose {
		logger.Info("http response", zap.String("body", r.Resp))
	}

	cid, e := jsonparser.GetString([]byte(r.Resp), "cid")
	if e != nil {
		r.Ret = -7
		r.Err = e
		r.ErrMsg = r.Err.Error()
		chResults <- r
		return
	}

	r.Cid = cid
	chResults <- r
}

func postFiles() error {
	logger.Info("postFiles",
		zap.String("FilesDir", params.FilesDir),
		zap.Int("From", input.From),
		zap.Int("To", input.To),
		zap.Int("BlockSize", input.BlockSize),
		zap.Bool("Pin", input.Pin),
		zap.Int("ReplicationMin", input.ReplicationMin),
		zap.Int("ReplicationMax", input.ReplicationMax),
	)

	var prsWg sync.WaitGroup
	prsWg.Add(1)
	go countResults(&prsWg)

	urlAdd = "http://" + input.HostPort + "/add" + clusterAdd("")
	if params.Verbose {
		logger.Debug(urlAdd)
	}

	chFids := make(chan int, 10000)
	go func() {
		for i := input.From; i < input.To; i++ {
			chFids <- i
		}
		close(chFids)
	}()

	var wg sync.WaitGroup
	wg.Add(input.Goroutines)
	for i := 0; i < input.Goroutines; i++ {
		go func(gid int) {
			defer wg.Done()

			for {
				fid, ok := <-chFids
				if !ok {
					return
				}

				postFile(gid, fid)
			}
		}(i)
	}
	wg.Wait()

	close(chResults)

	prsWg.Wait()
	return nil
}
