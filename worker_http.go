package main

import (
	"net/http"
	"sync"
)

func doRepeatHttpInput(pu ParamsUrl) error {
	url := baseUrl() + pu.paramsUrl()

	if p.Verbose {
		logger.Debug(url)
	}

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(p.Goroutines)
	for i := 0; i < p.Goroutines; i++ {
		go func() {
			defer wg.Done()

			req, _ := http.NewRequest(p.Method, url, nil)

			c := 0
			for c < repeat {
				c++
				r := doHttpRequest(req, p.DropHttpResp)
				chResults <- r
			}
		}()
	}

	wg.Wait()
	close(chResults)

	countResultsWg.Wait()
	return nil
}
