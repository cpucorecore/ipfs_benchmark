package main

import (
	"fmt"
	"net/url"
)

var paramsFunctions = map[string]func(string) string{
	"/api/v0/dag/stat":          dagStat,
	"/api/v0/routing/findprovs": routingFindprovs,
	"/api/v0/cat":               cat,
}

func getParamsFunc(apiPath string) func(string) string {
	return paramsFunctions[apiPath]
}

func empty() string {
	return ""
}

// for ipfs
func dagStat(cid string) string {
	values := url.Values{
		"arg":      {cid},
		"progress": {"true"},
	}

	return "?" + values.Encode()
}

func routingFindprovs(cid string) string {
	values := url.Values{
		"arg": {cid},
	}

	return "?" + values.Encode()
}

func cat(cid string) string {
	values := url.Values{
		"arg":      {cid},
		"offset":   {"0"},
		"progress": {"true"},
	}

	return "?" + values.Encode()
}

func swarmPeers() string {
	return "?verbose=true&streams=false&latency=true&direction=true"
}

// for cluster
func clusterPinAdd(_ string) string {
	min := fmt.Sprintf("%d", input.ReplicationMin)
	max := fmt.Sprintf("%d", input.ReplicationMax)
	values := url.Values{
		"mode":            {"recursive"},
		"replication-min": {min},
		"replication-max": {max},
	}

	return "?" + values.Encode()
}

func clusterAdd(_ string) string {
	chunker := fmt.Sprintf("size-%d", input.BlockSize)
	noPin := fmt.Sprintf("%t", !input.Pin)
	min := fmt.Sprintf("%d", input.ReplicationMin)
	max := fmt.Sprintf("%d", input.ReplicationMax)
	values := url.Values{
		"chunker":         {chunker},
		"cid-version":     {"0"},
		"format":          {"unixfs"},
		"local":           {"false"},
		"mode":            {"recursive"},
		"no-pin":          {noPin},
		"replication-min": {min},
		"replication-max": {max},
	}

	return "?" + values.Encode()
}
