package main

import (
	"bytes"
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

type Fid2Cid struct {
	Fid int
	Cid string
}

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

func doRequest(gid int, method, url string, fid2cid Fid2Cid) {
	r := Result{
		Gid: gid,
		Fid: fid2cid.Fid,
		Cid: fid2cid.Cid,
	}

	var req *http.Request
	req, _ = http.NewRequest(method, url+fid2cid.Cid, nil)
	r.S = time.Now()
	resp, e := httpClient.Do(req)
	r.E = time.Now()
	r.LatenciesMicroseconds = r.E.Sub(r.S).Microseconds()
	if e != nil { // TODO retry
		logger.Error("httpClient do err", zap.String("err", e.Error()))
		r.Ret = -1
		//r.Error = e
		if resp != nil && resp.Body != nil {
			logger.Debug("err with response")
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("read response err", zap.String("err", e.Error()))
			}
			resp.Body.Close()
			r.Resp = string(body)
		}
		chResults <- r
		return
	}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		r.Ret = -2
		//r.Error = e
		chResults <- r
		return
	}

	if params.Verbose {
		logger.Info("http response", zap.String("body", string(body)))
	}

	if resp.StatusCode != 200 && resp.StatusCode != 404 { // TODO retry
		logger.Error("do http request failed",
			zap.Int("fid", fid2cid.Fid),
			zap.String("cid", fid2cid.Cid),
			zap.Int("status", resp.StatusCode),
			zap.String("resp", string(body)))

		r.Ret = -3
		//r.Error = e
		chResults <- r
	}

	chResults <- r
}

func doRequests(method, path string) error {
	url := "http://" + input.HostPort + path

	var prsWg sync.WaitGroup
	prsWg.Add(1)
	go countResults(&prsWg)

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

				doRequest(gid, method, url, cid)
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
		//r.Error = e
		chResults <- r
		return
	}

	fp := filepath.Join(params.FilesDir, fidStr)
	f, e := os.Open(fp)
	if e != nil {
		r.Ret = -2
		//r.Error = e
		chResults <- r
		return
	}
	defer f.Close()

	n, e := io.Copy(ff, f)
	if e != nil {
		logger.Error("read file err", zap.String("err", e.Error()), zap.Int64("read bytes", n))
		w.Close()
		r.Ret = -3
		//r.Error = e
		chResults <- r
		return
	}

	e = w.Close()
	if e != nil {
		r.Ret = -4
		//r.Error = e
		chResults <- r
		return
	}

	req, e := http.NewRequest(http.MethodPost, sendFileUrl, &b)
	if e != nil {
		r.Ret = -5
		//r.Error = e
		chResults <- r
		return
	}

	req.Header.Add("Content-Type", w.FormDataContentType())

	r.S = time.Now()
	resp, e := httpClient.Do(req)
	r.E = time.Now()
	r.LatenciesMicroseconds = r.E.Sub(r.S).Microseconds()
	if e != nil {
		r.Ret = -6
		//r.Error = e
		if resp != nil && resp.Body != nil {
			logger.Debug("err with response")
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("read response err", zap.String("err", e.Error()))
			}
			resp.Body.Close()
			r.Resp = string(body)
		}
		chResults <- r
		return
	}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		r.Ret = -7
		//r.Error = e
		chResults <- r
		return
	}

	if params.Verbose {
		logger.Info("http response", zap.String("body", string(body)))
	}

	cid, e := jsonparser.GetString(body, "cid")
	if e != nil {
		r.Ret = -8
		//r.Error = e
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

	chunker := fmt.Sprintf("size-%d", input.BlockSize)
	noPin := fmt.Sprintf("%t", !input.Pin)
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

	close(chResults)

	prsWg.Wait()
	return nil
}
