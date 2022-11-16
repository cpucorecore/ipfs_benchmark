package main

import (
	"net/http"
	"sync"

	"go.uber.org/zap"
)

func doHttpRequests(method, path string, iterativePath bool) error {
	baseUrl := HTTP + input.HostPort + path

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(input.Goroutines)
	for i := 0; i < input.Goroutines; i++ {
		go func() {
			defer wg.Done()

			for {
				fid2Cid, ok := <-chFid2Cids
				if !ok {
					break
				}

				var url string
				if iterativePath {
					url = baseUrl + "/" + fid2Cid.Cid
					paramsFunction := paramsFunctionsRepetitive[path]
					if paramsFunction != nil {
						url = baseUrl + paramsFunction()
					}
				} else {
					paramsFunction := paramsFunctionsIterative[path]
					if paramsFunction != nil {
						url = baseUrl + paramsFunction(fid2Cid.Cid)
					}
				}

				if params.Verbose {
					logger.Debug("http req", zap.String("url", url))
				}

				req, _ := http.NewRequest(method, url, nil)

				r := doHttpRequest(req)
				r.Cid = fid2Cid.Cid
				r.Fid = fid2Cid.Fid

				if params.Verbose {
					logger.Debug("http response", zap.String("body", r.Resp))
				}

				if r.Err != nil {
					r.Ret = -1
					chResults <- r
					continue
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

func doRequestsRepeat(method, path string, repeat int) error {
	baseUrl := "http://" + input.HostPort + path

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(input.Goroutines)
	for i := 0; i < input.Goroutines; i++ {
		go func(gid int) {
			defer wg.Done()

			url := baseUrl
			paramsFunction := paramsFunctionsRepetitive[path]
			if paramsFunction != nil {
				url += paramsFunction()
			}

			if params.Verbose {
				logger.Debug(url)
			}
			req, _ := http.NewRequest(method, url, nil)

			c := 0
			for c < repeat {
				c++
				r := doHttpRequest(req)

				if params.Verbose {
					logger.Debug("http response", zap.String("body", r.Resp))
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
