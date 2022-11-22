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
	req, e := createPostFileRequest(fid)
	if e != nil {
		return Result{Fid: fid, Ret: ErrCreateRequest, Err: e}
	}

	return doHttpRequest(req, false)
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

				r := postFile(fid)
				r.Fid = fid
				if r.Ret == 0 {
					cid, e := jsonparser.GetString([]byte(r.Resp), "cid")
					if e != nil {
						r.Ret = ErrCategoryJson
						r.Err = e
					}
					r.Cid = cid
				}

				chResults <- r
			}
		}()
	}
	wg.Wait()

	close(chResults)

	countResultsWg.Wait()
	return nil
}
