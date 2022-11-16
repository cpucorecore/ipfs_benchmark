package main

import (
	"fmt"
	"net/url"
)

var paramsFunctionsRepetitive = map[string]func() string{
	"/pin/add":            clusterPinAddParams,
	"/add":                clusterAddParams,
	"/api/v0/swarm/peers": swarmPeersParams,
}

func clusterPinAddParams() string {
	min := fmt.Sprintf("%d", input.ReplicationMin)
	max := fmt.Sprintf("%d", input.ReplicationMax)
	values := url.Values{
		"mode":            {"recursive"},
		"replication-min": {min},
		"replication-max": {max},
	}

	return "?" + values.Encode()
}

func clusterAddParams() string {
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

func swarmPeersParams() string {
	return "?verbose=true&streams=false&latency=true&direction=true"
}
