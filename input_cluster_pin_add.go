package main

import (
	"fmt"
	"net/url"
)

var _ IterUrlHttpInput = ClusterPinAddInput{}

type ClusterPinAddInput struct {
	BaseHttpInput
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
