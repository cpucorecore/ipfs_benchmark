package main

import (
	"net/http"
	"sync"

	"go.uber.org/zap"
)

func doIterUrlHttpInput(input IIterUrlHttpInput) error {
	baseUrl := input.baseUrl()

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for {
				fid2Cid, ok := <-chFid2Cids
				if !ok {
					break
				}

				url := baseUrl + "/" + fid2Cid.Cid + input.urlParams()

				if verbose {
					logger.Debug("http req", zap.String("url", url))
				}

				req, _ := http.NewRequest(input.method(), url, nil)

				r := doHttpRequest(req)
				r.Cid = fid2Cid.Cid
				r.Fid = fid2Cid.Fid

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
