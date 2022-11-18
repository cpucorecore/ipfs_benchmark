package main

import (
	"fmt"
	"net/url"
)

type ClusterPinInput struct {
	HttpInput
	TestFile string
}

func (i ClusterPinInput) iterUrl(it string) string {
	return "/" + it
}

var _ IHttpInputIterUrl = ClusterPinAddInput{}

type ClusterPinAddInput struct {
	ClusterPinInput
	Replica int
}

func (i ClusterPinAddInput) urlParams() string {
	min := fmt.Sprintf("%d", i.Replica)
	max := fmt.Sprintf("%d", i.Replica)
	values := url.Values{
		"mode":            {"recursive"},
		"replication-min": {min},
		"replication-max": {max},
	}

	return "?" + values.Encode()
}

var _ IHttpInputIterUrl = ClusterPinRmInput{}

type ClusterPinRmInput struct {
	ClusterPinInput
}

func (i ClusterPinRmInput) urlParams() string {
	return ""
}

var _ IHttpInputIterUrl = ClusterPinGetInput{}

type ClusterPinGetInput struct {
	ClusterPinInput
}

func (i ClusterPinGetInput) urlParams() string {
	return "?local=false"
}
