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
	MaxIdleConns:          1000,
	IdleConnTimeout:       600 * time.Second,
	ExpectContinueTimeout: 600 * time.Second,
	MaxIdleConnsPerHost:   1000,
}

var httpClient = &http.Client{Transport: transport}
var activeRequest int32

func doHttpRequest(req *http.Request) (startTime, endTime time.Time, currentActiveRequest int32, e error, body string) {
	if params.Sync {
		atomic.AddInt32(&activeRequest, 1)
		currentActiveRequest = activeRequest
	} else {
		currentActiveRequest = int32(input.Goroutines)
	}
	startTime = time.Now()
	resp, e := httpClient.Do(req)
	endTime = time.Now()
	if params.Sync {
		atomic.AddInt32(&activeRequest, -1)
	}

	if e != nil {
		return
	}

	defer resp.Body.Close()
	respBody, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return
	}
	body = string(respBody)

	return
}

func doRequest(gid int, method, baseUrl string, fid2cid Fid2Cid, paramsStr string) {
	r := Result{
		Gid: gid,
		Fid: fid2cid.Fid,
		Cid: fid2cid.Cid,
	}

	u := baseUrl + fid2cid.Cid + paramsStr
	if params.Verbose {
		logger.Debug("request url", zap.String("url", u))
	}

	req, _ := http.NewRequest(method, u, nil)

	r.S, r.E, r.Concurrency, r.Err, r.Resp = doHttpRequest(req)
	r.Latency = r.E.Sub(r.S).Microseconds()
	if r.Err != nil {
		r.Ret = -6
		chResults <- r
		return
	}

	if params.Verbose {
		logger.Debug("http response", zap.String("body", r.Resp))
	}

	chResults <- r
}

func doRequests(method, path string, pf func() string) error {
	baseUrl := "http://" + input.HostPort + path

	var prsWg sync.WaitGroup
	prsWg.Add(1)
	go countResults(&prsWg)

	paramsStr := pf()
	var wg sync.WaitGroup
	for i := 0; i < input.Goroutines; i++ {
		wg.Add(1)

		go func(gid int) {
			defer wg.Done()

			for {
				cid, ok := <-chFid2Cids
				if !ok {
					return
				}

				doRequest(gid, method, baseUrl, cid, paramsStr)
			}
		}(i)
	}

	wg.Wait()
	close(chResults)

	prsWg.Wait()
	return nil
}

var sendFileUrl string

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
		chResults <- r
		return
	}

	fp := filepath.Join(params.FilesDir, fidStr)
	f, e := os.Open(fp)
	if e != nil {
		r.Ret = -2
		r.Err = e
		chResults <- r
		return
	}
	defer f.Close()

	n, e := io.Copy(ff, f)
	if e != nil {
		logger.Error("read file err", zap.String("err", e.Error()), zap.Int64("read bytes", n))
		r.Ret = -3
		r.Err = e
		chResults <- r
		return
	}

	e = w.Close()
	if e != nil {
		r.Ret = -4
		r.Err = e
		chResults <- r
		return
	}

	req, e := http.NewRequest(http.MethodPost, sendFileUrl, &b)
	if e != nil {
		r.Ret = -5
		r.Err = e
		chResults <- r
		return
	}

	req.Header.Add("Content-Type", w.FormDataContentType())

	r.S, r.E, r.Concurrency, r.Err, r.Resp = doHttpRequest(req)
	r.Latency = r.E.Sub(r.S).Microseconds()
	if r.Err != nil {
		r.Ret = -6
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
		chResults <- r
		return
	}

	r.Cid = cid
	chResults <- r
}

func sendFiles() error {
	logger.Info("sendFiles",
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

	sendFileUrl = "http://" + input.HostPort + "/add?" + genHttpParamsClusterAdd()
	if params.Verbose {
		logger.Debug(sendFileUrl)
	}

	chFids := make(chan int, 10000)
	go func() {
		for i := input.From; i < input.To; i++ {
			chFids <- i
		}
		close(chFids)
	}()

	var wg sync.WaitGroup
	for i := 0; i < input.Goroutines; i++ {
		wg.Add(1)

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
