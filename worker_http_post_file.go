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
	"time"

	"github.com/buger/jsonparser"
)

var inputParamsUrl string

const (
	ErrCreateFormFile    = ErrCategoryHttpRequest + 1
	ErrOpenFile          = ErrCategoryHttpRequest + 2
	ErrCopyFile          = ErrCategoryHttpRequest + 3
	ErrCloseFile         = ErrCategoryHttpRequest + 4
	ErrCloseWriter       = ErrCategoryHttpRequest + 5
	ErrCreateHttpRequest = ErrCategoryHttpRequest + 6

	ErrJsonParse = ErrCategoryJson + 1
)

func postFile(b *bytes.Buffer, fid int) Result {
	r := Result{Fid: fid}

	b.Reset()
	w := multipart.NewWriter(b)
	fileName := fmt.Sprintf("%d", fid)
	formFile, err := w.CreateFormFile("file", fileName)
	if err != nil {
		r.Ret = ErrCreateFormFile
		r.Err = err
		return r
	}

	fp := filepath.Join(PathFiles, fileName)
	f, err := os.Open(fp)
	if err != nil {
		r.Ret = ErrOpenFile
		r.Err = err
		return r
	}

	_, err = io.Copy(formFile, f)
	if err != nil {
		r.Ret = ErrCopyFile
		r.Err = err
		return r
	}

	err = f.Close()
	if err != nil {
		r.Ret = ErrCloseFile
		r.Err = err
		return r
	}

	err = w.Close()
	if err != nil {
		r.Ret = ErrCloseWriter
		r.Err = err
		return r
	}

	req, err := http.NewRequest(http.MethodPost, baseUrl()+inputParamsUrl, b)
	if err != nil {
		r.Ret = ErrCreateHttpRequest
		r.Err = err
		return r
	}
	req.Header.Add("Content-Type", w.FormDataContentType())

	r = doHttpRequest(req, false)
	if r.Ret == 0 {
		//r.Cid, r.Err = jsonparser.GetString([]byte(r.Resp), "cid")
		//if r.Err != nil {
		//	r.Ret = ErrJsonParse
		//}

		_, r.Err = jsonparser.ArrayEach([]byte(r.Resp), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			r.Cid, r.Err = jsonparser.GetString(value, "cid")
		})
		if r.Err != nil {
			r.Ret = ErrJsonParse
		}

	}
	r.Fid = fid

	return r
}

func postFileWithRetry(b *bytes.Buffer, fid int) Result {
	retry := 0
	var r Result

	for retry < p.MaxRetry {
		retry++

		r = postFile(b, fid)
		if r.Ret == 0 {
			return r
		}

		logger.Debug(fmt.Sprintf("fid:%d, ret:%d, resp:%s, retry:%d, err:%s", r.Fid, r.Ret, r.Resp, retry, r.Err.Error()))
		time.Sleep(time.Second * 2 * time.Duration(retry))
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

			fileBuffer := make([]byte, 0, 1024*1024*fileBufferSize)
			buffer := bytes.NewBuffer(fileBuffer)
			for {
				fid, ok := <-chFids
				if !ok {
					return
				}

				chResults <- postFileWithRetry(buffer, fid)
			}
		}()
	}
	wg.Wait()

	close(chResults)

	countResultsWg.Wait()
	return nil
}
