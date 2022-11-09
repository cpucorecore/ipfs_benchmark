package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"go.uber.org/zap"
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 120 * time.Second,
	}).DialContext,
	MaxIdleConns:          1000,
	IdleConnTimeout:       600 * time.Second,
	ExpectContinueTimeout: 60 * time.Second,
	MaxIdleConnsPerHost:   1000,
}

var httpClient = &http.Client{Transport: transport}

func doRequest(gid int, method, url, cid string) {
	r := Result{
		Gid: gid,
		Cid: cid,
	}

	var req *http.Request
	req, _ = http.NewRequest(method, url+cid, nil)
	r.S = time.Now()
	resp, e := httpClient.Do(req)
	r.E = time.Now()
	if e != nil { // TODO retry
		logger.Error("httpClient do err", zap.String("err", e.Error()))
		r.Ret = -1
		r.Error = e
		chResults <- r
		return
	}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		r.Ret = -2
		r.Error = e
		chResults <- r
		return
	}

	if params.Verbose {
		logger.Info("http response", zap.String("body", string(body)))
	}

	if resp.StatusCode != 200 && resp.StatusCode != 404 { // TODO retry
		logger.Error("do http request failed",
			zap.String("cid", cid),
			zap.Int("status", resp.StatusCode),
			zap.String("resp", string(body)))

		r.Ret = -3
		r.Error = e
		chResults <- r
	}

	chResults <- r
}

func doRequests(method, path string) error {
	defer close(chResults)

	url := "http://" + input.HostPort + path

	var wg sync.WaitGroup
	for i := 0; i < input.Goroutines; i++ {
		wg.Add(1)

		go func(gid int) {
			defer wg.Done()

			for {
				cid, ok := <-chCids
				if !ok {
					return
				}

				doRequest(gid, method, url, cid)
			}
		}(i)
	}

	wg.Wait()
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
		r.Error = e
		chResults <- r
		return
	}

	fp := filepath.Join(params.Dir, fidStr)
	f, e := os.Open(fp)
	if e != nil {
		r.Ret = -2
		r.Error = e
		chResults <- r
		return
	}
	defer f.Close()

	_, e = io.Copy(ff, f)
	if e != nil {
		r.Ret = -3
		r.Error = e
		chResults <- r
		return
	}

	e = w.Close()
	if e != nil {
		r.Ret = -4
		r.Error = e
		chResults <- r
		return
	}

	req, e := http.NewRequest(http.MethodPost, sendFileUrl, &b)
	if e != nil {
		r.Ret = -5
		r.Error = e
		chResults <- r
		return
	}

	req.Header.Add("Content-Type", w.FormDataContentType())

	r.S = time.Now()
	resp, e := httpClient.Do(req)
	r.E = time.Now()
	if e != nil {
		r.Ret = -6
		r.Error = e
		chResults <- r
		return
	}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		r.Ret = -7
		r.Error = e
		chResults <- r
		return
	}

	if params.Verbose {
		logger.Info("http response", zap.String("body", string(body)))
	}

	cid, e := jsonparser.GetString(body, "cid")
	if e != nil {
		r.Ret = -8
		r.Error = e
		chResults <- r
		return
	}

	r.Cid = cid
	chResults <- r
}

func sendFiles(pin bool) error {
	defer close(chResults)

	logger.Info("sendFiles",
		zap.String("Dir", params.Dir),
		zap.Int("From", input.From),
		zap.Int("To", input.To),
		zap.Int("BlockSize", input.BlockSize),
		zap.Bool("pin", pin),
		zap.Int("ReplicationMin", input.ReplicationMin),
		zap.Int("ReplicationMax", input.ReplicationMax),
	)

	if input.BlockSize > 1024*1024 {
		return errors.New("BlockSize can not > 1MB")
	}

	chunker := fmt.Sprintf("size-%d", input.BlockSize)
	noPin := fmt.Sprintf("%t", !pin)
	replicationMin := fmt.Sprintf("%d", input.ReplicationMin)
	replicationMax := fmt.Sprintf("%d", input.ReplicationMax)
	httpParams := url.Values{
		"chunker":         {chunker},
		"cid-version":     {"0"},
		"format":          {"unixfs"},
		"local":           {"false"},
		"mode":            {"recursive"},
		"no-pin":          {noPin},
		"replication-min": {replicationMin},
		"replication-max": {replicationMax},
	}

	sendFileUrl = "http://" + input.HostPort + "/add?" + httpParams.Encode()
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
	close(chCids)
	return nil
}
