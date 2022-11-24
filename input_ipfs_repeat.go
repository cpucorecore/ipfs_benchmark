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

func (i IpfsSwarmPeersInput) info() string {
	return i.RepeatHttpParams.info() + "_" + i.Tag
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

func (i IpfsIdInput) info() string {
	return i.RepeatHttpParams.info() + "_" + i.Tag
}

func (i IpfsIdInput) paramsUrl() string {
	return ""
}

type IpfsRepoStat struct {
	RepeatHttpParams
	SizeOnly, Human bool
}

func (i IpfsRepoStat) info() string {
	return i.RepeatHttpParams.info() + "_" + i.Tag
}

func (i IpfsRepoStat) paramsUrl() string {
	return "?size-only=false&human=false"
}
