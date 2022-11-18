package main

import (
	"fmt"
	"net/url"
)

var _ IHttpInputIterParams = IpfsDhtFindprovsInput{}

type IpfsDhtFindprovsInput struct {
	IterHttpInput
	Verbose bool
}

func (i IpfsDhtFindprovsInput) iterUrlParams(it string) string {
	values := url.Values{
		"arg":     {it},
		"verbose": {fmt.Sprintf("%v", i.Verbose)},
	}

	return "?" + values.Encode()
}

func (i IpfsDhtFindprovsInput) url(it string) string {
	return i.baseUrl() + i.iterUrlParams(it)
}

var _ IHttpInputIterParams = IpfsDagStatInput{}

type IpfsDagStatInput struct {
	IterHttpInput
	Progress bool
}

func (i IpfsDagStatInput) iterUrlParams(it string) string {
	values := url.Values{
		"arg":      {it},
		"progress": {fmt.Sprintf("%v", i.Progress)},
	}

	return "?" + values.Encode()
}

func (i IpfsDagStatInput) url(it string) string {
	return i.baseUrl() + i.iterUrlParams(it)
}

var _ IHttpInputIterParams = IpfsCatInput{}

type IpfsCatInput struct {
	IterHttpInput
	Offset   int
	Length   int
	Progress bool
}

func (i IpfsCatInput) iterUrlParams(it string) string {
	values := url.Values{
		"arg":      {it},
		"offset":   {fmt.Sprintf("%d", i.Offset)},
		"length":   {fmt.Sprintf("%d", i.Length)},
		"progress": {fmt.Sprintf("%v", i.Progress)},
	}

	return "?" + values.Encode()
}

func (i IpfsCatInput) url(it string) string {
	return i.baseUrl() + i.iterUrlParams(it)
}
