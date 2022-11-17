package main

import (
	"io/ioutil"

	"go.uber.org/zap"
)

func gc() error {
	url := "http://" + hostPort + "/ipfs/gc?local=false"
	resp, e := httpClient.Post(url, "", nil)
	if e != nil {
		logger.Error("httpClient post err", zap.String("err", e.Error()))
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return e
	}

	_, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		logger.Error("read response body err", zap.String("err", e.Error()))
		resp.Body.Close()
		return e
	}

	return nil
}
