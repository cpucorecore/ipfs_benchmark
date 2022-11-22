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
	return fmt.Sprintf("%s_replica%d_%s", i.IterHttpParams.info(), i.Replica, i.Tag)
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

func (i ClusterPinRmInput) info() string {
	return i.IterHttpParams.info() + "_" + i.Tag
}

func (i ClusterPinRmInput) paramsUrl() string {
	return ""
}

type ClusterPinGetInput struct {
	IterHttpParams
}

func (i ClusterPinGetInput) info() string {
	return i.IterHttpParams.info() + "_" + i.Tag
}

func (i ClusterPinGetInput) paramsUrl() string {
	return "?local=false"
}

type ClusterUnpinByCidInput struct {
	HttpParams
	cidFile string
	Range
}

func (i ClusterUnpinByCidInput) info() string {
	return fmt.Sprintf("%s_%s_%s", i.HttpParams.info(), i.Range.info(), i.Tag)
}

func (i ClusterUnpinByCidInput) check() bool {
	return i.HttpParams.check() && len(i.cidFile) > 0 && i.Range.check() // TODO check cid file exist
}

func (i ClusterUnpinByCidInput) paramsUrl() string {
	return ""
}
