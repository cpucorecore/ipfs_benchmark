package main

import (
	"net/http"
	"sync"
)

func doRepeatHttpInput(input ConstHttpInput, repeat int) error {
	url := input.baseUrl() + input.urlParams()
	if verbose {
		logger.Debug(url)
	}

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(gid int) {
			defer wg.Done()

			logger.Info(input.method())
			logger.Info(url)
			req, _ := http.NewRequest(input.method(), url, nil)

			c := 0
			for c < repeat {
				c++
				r := doHttpRequest(req)
				chResults <- r
			}
		}(i)
	}

	wg.Wait()
	close(chResults)

	countResultsWg.Wait()
	return nil
}
