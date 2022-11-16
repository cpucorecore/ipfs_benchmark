package main

import "net/url"

var paramsFunctionsIterative = map[string]func(string) string{
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
