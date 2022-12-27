package main

import (
	"fmt"
	"net/url"
)

type ClusterPinAddInput struct {
	IterHttpParams
	Replica int
}

func (i ClusterPinAddInput) name() string {
	return "cluster_pins_add"
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

func (i ClusterPinRmInput) name() string {
	return "cluster_pins_rm"
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

func (i ClusterPinGetInput) name() string {
	return "cluster_pins_get"
}

func (i ClusterPinGetInput) info() string {
	return i.IterHttpParams.info() + "_" + i.Tag
}

func (i ClusterPinGetInput) paramsUrl() string {
	return "?local=false"
}
