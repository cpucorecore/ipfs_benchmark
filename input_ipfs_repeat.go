package main

import (
	"fmt"
	"net/url"
)

type IpfsSwarmPeersInput struct {
	RepeatHttpParams
	Verbose_  bool
	Streams   bool
	Latency   bool
	Direction bool
}

func (i IpfsSwarmPeersInput) paramsUrl() string {
	values := url.Values{
		"Verbose":   {fmt.Sprintf("%v", i.Verbose_)},
		"streams":   {fmt.Sprintf("%v", i.Streams)},
		"latency":   {fmt.Sprintf("%v", i.Latency)},
		"direction": {fmt.Sprintf("%v", i.Direction)},
	}
	return "?" + values.Encode()
}

type IpfsIdInput struct {
	RepeatHttpParams
}

func (i IpfsIdInput) paramsUrl() string {
	return ""
}
