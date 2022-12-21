package main

import (
	"net/http"
	"sync"

	"go.uber.org/zap"
)

func doIterUrlRequest(pu ParamsUrl) error {
	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(p.Goroutines)
	for i := 0; i < p.Goroutines; i++ {
		go func() {
			defer wg.Done()

			for {
				fid2Cid, ok := <-chFid2Cids
				if !ok {
					break
				}

				url := baseUrl() + "/" + fid2Cid.Cid + pu.paramsUrl()

				if p.Verbose {
					logger.Debug("http req", zap.String("url", url))
				}

				req, _ := http.NewRequest(p.Method, url, nil)

				r := doHttpRequest(req, p.DropHttpResp)
				r.Cid = fid2Cid.Cid
				r.Fid = fid2Cid.Fid

				if r.Err != nil {
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
