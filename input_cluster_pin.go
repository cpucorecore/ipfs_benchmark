package main

import (
	"fmt"
	"net/url"
)

type ClusterPinAddInput struct {
	IterHttpParams
	Replica int
}

func (i ClusterPinAddInput) info() string {
	return fmt.Sprintf("%s_replica%d", i.IterHttpParams.info(), i.Replica)
}

func (i ClusterPinAddInput) check() bool {
	return i.IterHttpParams.check() && i.Replica > 0
}

func (i ClusterPinAddInput) paramsUrl() string {
	min := fmt.Sprintf("%d", i.Replica)
	max := fmt.Sprintf("%d", i.Replica)
	values := url.Values{
		"mode":            {"recursive"},
		"replication-min": {min},
		"replication-max": {max},
	}

	return "?" + values.Encode()
}

type ClusterPinRmInput struct {
	IterHttpParams
}

func (i ClusterPinRmInput) paramsUrl() string {
	return ""
}

type ClusterPinGetInput struct {
	IterHttpParams
}

func (i ClusterPinGetInput) paramsUrl() string {
	return "?local=false"
}
