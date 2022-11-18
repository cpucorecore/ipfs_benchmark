package main

import (
	"fmt"
	"net/url"
)

var _ IHttpInputRepeat = IpfsSwarmPeersInput{}

type IpfsSwarmPeersInput struct {
	RepeatHttpInput
	Verbose   bool
	Streams   bool
	Latency   bool
	Direction bool
}

func (i IpfsSwarmPeersInput) urlParams() string {
	values := url.Values{
		"verbose":   {fmt.Sprintf("%v", i.Verbose)},
		"streams":   {fmt.Sprintf("%v", i.Streams)},
		"latency":   {fmt.Sprintf("%v", i.Latency)},
		"direction": {fmt.Sprintf("%v", i.Direction)},
	}
	return "?" + values.Encode()
}

func (i IpfsSwarmPeersInput) url(_ string) string {
	return i.baseUrl() + i.urlParams()
}
