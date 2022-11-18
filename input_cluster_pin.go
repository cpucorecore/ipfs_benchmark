package main

import (
	"fmt"
	"net/url"
)

var _ IIterUrlHttpInput = ClusterPinAddInput{}

type ClusterPinAddInput struct {
	HttpInput
	TestFile string
	Replica  int
}

func (i ClusterPinAddInput) iterUrl(it string) string {
	return "/" + it
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

var _ IIterUrlHttpInput = ClusterPinRmInput{}

type ClusterPinRmInput ClusterPinAddInput

func (i ClusterPinRmInput) iterUrl(it string) string {
	return "/" + it
}

func (i ClusterPinRmInput) urlParams() string {
	return ""
}

var _ IIterUrlHttpInput = ClusterPinGetInput{}

type ClusterPinGetInput ClusterPinAddInput

func (i ClusterPinGetInput) iterUrl(it string) string {
	return "/" + it
}

func (i ClusterPinGetInput) urlParams() string {
	return "?local=false"
}
