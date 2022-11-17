package main

import "net/url"

var iterative = map[string]func(string) string{
	"/api/v0/dag/stat":          dagStatParams,
	"/api/v0/dht/findprovs":     routingFindprovsParams, // deprecated in new version in ipfs
	"/api/v0/routing/findprovs": routingFindprovsParams,
	"/api/v0/cat":               catParams,
}

func dagStatParams(cid string) string {
	values := url.Values{
		"arg":      {cid},
		"progress": {"true"},
	}

	return "?" + values.Encode()
}

func routingFindprovsParams(cid string) string {
	values := url.Values{
		"arg": {cid},
	}

	return "?" + values.Encode()
}

func catParams(cid string) string {
	values := url.Values{
		"arg":      {cid},
		"offset":   {"0"},
		"progress": {"true"},
	}

	return "?" + values.Encode()
}

var repetitive = map[string]func() string{
	"/api/v0/swarm/peers": swarmPeersParams,
}

func swarmPeersParams() string {
	return "?verbose=true&streams=false&latency=true&direction=true"
}
