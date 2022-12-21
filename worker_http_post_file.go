package main

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/buger/jsonparser"
)

var inputParamsUrl string

const (
	ErrJsonParse = ErrCategoryJson + 1
)

func createPostFileRequest(fid int) (*http.Request, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fidStr := fmt.Sprintf("%d", fid)
	ff, e := w.CreateFormFile("file", fidStr)
	if e != nil {
		return nil, e
	}

	fp := filepath.Join(PathFiles, fidStr)
	f, e := os.Open(fp)
	if e != nil {
		return nil, e
	}

	defer f.Close()

	_, e = io.Copy(ff, f)
	if e != nil {
		return nil, e
	}

	e = w.Close()
	if e != nil {
		return nil, e
	}

	u := baseUrl() + inputParamsUrl
	if p.Verbose {
		logger.Debug("http req", zap.String("url", u))
	}
	req, e := http.NewRequest(http.MethodPost, u, &b)
	if e != nil {
		return nil, e
	}

	req.Header.Add("Content-Type", w.FormDataContentType())

	return req, nil
}

func postFile(fid int) Result {
	retry := 0
	var r Result

	for retry < p.MaxRetry {
		retry++

		req, e := createPostFileRequest(fid)
		if e != nil {
			r.Fid = fid
			r.Ret = ErrCreateRequest
			r.Err = e
			logger.Info(fmt.Sprintf("create request err, fid:%d, retry:%d", r.Fid, retry))
			time.Sleep(time.Second * 2 * time.Duration(retry))
			continue
		}

		r = doHttpRequest(req, false)
		r.Fid = fid
		if r.Ret == 0 {
			cid, parseJsonErr := jsonparser.GetString([]byte(r.Resp), "cid")
			if parseJsonErr != nil {
				r.Ret = ErrJsonParse
				r.Err = parseJsonErr
				if r.Err == nil {
					logger.Info(fmt.Sprintf("fid:%d, ret:%d, resp:%s, retry:%d", r.Fid, r.Ret, r.Resp, retry))
				} else {
					logger.Info(fmt.Sprintf("fid:%d, ret:%d, resp:%s, retry:%d, err:%s", r.Fid, r.Ret, r.Resp, retry, r.Err.Error()))
				}
				time.Sleep(time.Second * 2 * time.Duration(retry))
				continue
			}

			r.Cid = cid
			return r
		} else {
			if r.Err == nil {
				logger.Info(fmt.Sprintf("fid:%d, ret:%d, resp:%s, retry:%d", r.Fid, r.Ret, r.Resp, retry))
			} else {
				logger.Info(fmt.Sprintf("fid:%d, ret:%d, resp:%s, retry:%d, err:%s", r.Fid, r.Ret, r.Resp, retry, r.Err.Error()))
			}
			time.Sleep(time.Second * 2 * time.Duration(retry))
			continue
		}
	}

	return r
}

func postFiles(input ClusterAddInput) error {
	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	inputParamsUrl = input.paramsUrl()

	chFids := make(chan int, 10000)
	go func() {
		for i := input.From; i < input.To; i++ {
			chFids <- i
		}
		close(chFids)
	}()

	var wg sync.WaitGroup
	wg.Add(p.Goroutines)
	for i := 0; i < p.Goroutines; i++ {
		go func() {
			defer wg.Done()

			for {
				fid, ok := <-chFids
				if !ok {
					return
				}

				chResults <- postFile(fid)
			}
		}()
	}
	wg.Wait()

	close(chResults)

	countResultsWg.Wait()
	return nil
}
