package main

import (
	"io/ioutil"

	"go.uber.org/zap"
)

func gc() error {
	url := "http://" + input.HostPort + "/ipfs/gc?local=false"
	resp, err := httpClient.Post(url, "", nil)
	if err != nil {
		logger.Error("httpClient post err", zap.String("err", err.Error()))
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("read response body err", zap.String("err", err.Error()))
		resp.Body.Close()
		return err
	}

	if params.Verbose {
		logger.Info(string(respBody))
	}

	return nil
}
