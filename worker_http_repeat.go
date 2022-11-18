package main

import (
	"net/http"
	"sync"
)

func doRepeatHttpInput(input IHttpInputRepeat) error {
	method := input.method()
	url := input.baseUrl() + input.urlParams()
	repeat := input.repeat()
	dropHttpResp := input.dropHttpResp()

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			req, _ := http.NewRequest(method, url, nil)

			c := 0
			for c < repeat {
				c++
				r := doHttpRequest(req, dropHttpResp)
				chResults <- r
			}
		}()
	}

	wg.Wait()
	close(chResults)

	countResultsWg.Wait()
	return nil
}
