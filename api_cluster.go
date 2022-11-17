package main

import (
	"go.uber.org/zap"
	"io/ioutil"
)

func gc(hostPort string) error {
	url := HTTP + hostPort + "/ipfs/gc?local=false"
	resp, e := httpClient.Post(url, "", nil)
	if e != nil {
		logger.Error("httpClient post err", zap.String("err", e.Error()))
		return e
	}

	_, e = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if e != nil {
		logger.Error("ioutil ReadAll err", zap.String("err", e.Error()))
		return e
	}

	logger.Info("gc finished")
	return nil
}
