package main

import (
	"fmt"
	"net/url"
)

var _ IIterUrlHttpInput = ClusterPinAddInput{}

type ClusterPinAddInput struct {
	HttpInput
	Replica int
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
