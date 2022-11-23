package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/buger/jsonparser"
)

var postFileUrl string

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

	req, e := http.NewRequest(http.MethodPost, postFileUrl, &b)
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
			continue
		}

		r = doHttpRequest(req, false)
		r.Fid = fid
		if r.Ret == 0 {
			cid, parseJsonErr := jsonparser.GetString([]byte(r.Resp), "cid")
			if parseJsonErr != nil {
				r.Ret = ErrJsonParse
				r.Err = parseJsonErr
				logger.Info(fmt.Sprintf("fid ret:%d, err:%s, resp:%s, retry:%d", r.Ret, r.Err.Error(), r.Resp, retry))
				continue
			}

			r.Cid = cid
			return r
		} else {
			logger.Info(fmt.Sprintf("fid ret:%d, err:%s, resp:%s, retry:%d", r.Ret, r.Err.Error(), r.Resp, retry))
			continue
		}
	}

	return r
}

func postFiles(input ClusterAddInput) error {
	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	postFileUrl = baseUrl() + input.paramsUrl()
	if p.Verbose {
		logger.Debug(postFileUrl)
	}

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
