package main

import (
	"fmt"
	"net/url"
)

type IpfsDhtFindprovsInput struct {
	IterHttpParams
	Verbose_ bool
}

func (i IpfsDhtFindprovsInput) paramsUrl(it string) string {
	values := url.Values{
		"arg":     {it},
		"Verbose": {fmt.Sprintf("%v", i.Verbose)},
	}

	return "?" + values.Encode()
}

type IpfsDagStatInput struct {
	IterHttpParams
	Progress bool
}

func (i IpfsDagStatInput) paramsUrl(it string) string {
	values := url.Values{
		"arg":      {it},
		"progress": {fmt.Sprintf("%v", i.Progress)},
	}

	return "?" + values.Encode()
}

type IpfsCatInput struct {
	IterHttpParams
	Offset   int
	Length   int
	Progress bool
}

func (i IpfsCatInput) paramsUrl(it string) string {
	values := url.Values{
		"arg":    {it},
		"offset": {fmt.Sprintf("%d", i.Offset)},
		//"length":   {fmt.Sprintf("%d", i.Length)},
		"progress": {fmt.Sprintf("%v", i.Progress)},
	}

	return "?" + values.Encode()
}
