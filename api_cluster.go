package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	ipfsPort    = "5001"
	clusterPort = "9094"
)

func buildIpfsUrl(ip, apiPath string) string {
	return "http://" + ip + ":" + ipfsPort + apiPath
}

func buildClusterUrl(ip, apiPath string) string {
	return "http://" + ip + ":" + clusterPort + apiPath
}

func callApi(method, url string) []byte {
	if p.Verbose {
		logger.Debug(fmt.Sprintf("callApi"), zap.String("method", method), zap.String("url", url))
	}

	req, e := http.NewRequest(method, url, nil)
	if e != nil {
		logger.Error("new request err")
		return nil
	}

	r := doHttpRequest(req, false)
	if r.Ret != 0 {
		logger.Error(fmt.Sprintf("do http request err:%s", e.Error()))
		return nil
	}

	return []byte(r.Resp)
}

//	IpfsInfo {
//	 "ID": "12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi",
//	 "PublicKey": "CAESIIk15/kGXzbWa9DF9VaOKZrikYVYXU3vYj8dzUw4Lt47",
//	 "Addresses": [
//	   "/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi",
//	   "/ip4/192.168.0.87/tcp/4001/p2p/12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi",
//	   "/ip6/::1/tcp/4001/p2p/12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi"
//	 ],
//	 "AgentVersion": "kubo/0.15.0/",
//	 "ProtocolVersion": "ipfs/0.1.0",
//	 "Protocols": [
//	   "/ipfs/bitswap",
//	   "/ipfs/bitswap/1.0.0",
//	   "/ipfs/bitswap/1.1.0",
//	   "/ipfs/bitswap/1.2.0",
//	   "/ipfs/id/1.0.0",
//	   "/ipfs/id/push/1.0.0",
//	   "/ipfs/lan/kad/1.0.0",
//	   "/ipfs/ping/1.0.0",
//	   "/libp2p/autonat/1.0.0",
//	   "/libp2p/circuit/relay/0.1.0",
//	   "/libp2p/circuit/relay/0.2.0/stop",
//	   "/p2p/id/delta/1.0.0",
//	   "/x/"
//	 ]
//	}
type IpfsInfo struct {
	ID string
}

func ipfsId() string {
	url := buildIpfsUrl(p.Hosts[0], "/api/v0/id")
	resp := callApi(http.MethodPost, url)
	if resp == nil {
		return ""
	}

	var ipfsInfo IpfsInfo
	e := json.Unmarshal(resp, &ipfsInfo)
	if e != nil {
		logger.Error(fmt.Sprintf("json parse IpfsInfo err:%s", e.Error()))
		return ""
	}

	return ipfsInfo.ID
}

type Peer struct {
	Addr string
	Peer string
}

type SwarmPeers struct {
	Peers []Peer
}

func ipfsSwarmPeers() (swarmPeers SwarmPeers) {
	url := buildIpfsUrl(p.Hosts[0], "/api/v0/swarm/peers")
	resp := callApi(http.MethodPost, url)
	if resp == nil {
		return
	}

	e := json.Unmarshal(resp, &swarmPeers)
	if e != nil {
		logger.Error(fmt.Sprintf("json parse Peers err:%s", e.Error()))
		return
	}

	for i, p := range swarmPeers.Peers {
		vs := strings.Split(p.Addr, "/") // p.Addr like "/ip4/192.168.0.86/tcp/4001"
		swarmPeers.Peers[i].Addr = vs[2]
	}

	id := ipfsId()

	swarmPeers.Peers = append(swarmPeers.Peers, Peer{Addr: p.Hosts[0], Peer: id})

	return
}

type RepoStat struct {
	RepoSize   int64
	StorageMax int64
	NumObjects int64
	RepoPath   string
	Version    string
}

const GB = 1024 * 1024 * 1024

func ipfsRepoStat(ip string) (repoStat RepoStat) {
	url := buildIpfsUrl(ip, "/api/v0/repo/stat")
	resp := callApi(http.MethodPost, url)
	if resp == nil {
		return
	}

	e := json.Unmarshal(resp, &repoStat)
	if e != nil {
		logger.Error(fmt.Sprintf("json parse RepoStat err:%s", e.Error()))
		return
	}

	return
}

type CidInfo struct {
	Cid string `json:"cid"`
}

func clusterPins(cidDetail bool) {
	//http://127.0.0.1:9094/pins?local=false&filter=&cids=

	//{"cid":"QmQDVZrti9ZAvZQHndy58TzfAAKQpkyWQZrSwrzn64qX5E","name":"","allocations":["12D3KooWB6cxrTahCGu4T1vLeJTsSU3fHnWJsrig6bNi8afVEekm","12D3KooWRh8hkpcCL6kk5LtQ7Fz1Kj5npsDSdHQSpnnEcq6WF7p4","12D3KooWJ7b5LSbZJmRvrgGQVoSyVM6bTQdjtSc6cBpLWoZTQKXH"],"origins":[],"created":"2022-11-21T01:33:20Z","metadata":null,"peer_map":{"12D3KooWB6cxrTahCGu4T1vLeJTsSU3fHnWJsrig6bNi8afVEekm":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi","ipfs_peer_addresses":["/ip4/192.168.0.87/tcp/4001/p2p/12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi"],"status":"pinned","timestamp":"2022-11-21T01:33:20Z","error":"","attempt_count":0,"priority_pin":false},"12D3KooWJ7b5LSbZJmRvrgGQVoSyVM6bTQdjtSc6cBpLWoZTQKXH":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWRC61AVD34UNGCcdutEadWzetteZDwv26f55s1zfmhFgC","ipfs_peer_addresses":["/ip4/192.168.0.85/tcp/4001/p2p/12D3KooWRC61AVD34UNGCcdutEadWzetteZDwv26f55s1zfmhFgC"],"status":"pinned","timestamp":"2022-11-21T09:33:20+08:00","error":"","attempt_count":0,"priority_pin":false},"12D3KooWRh8hkpcCL6kk5LtQ7Fz1Kj5npsDSdHQSpnnEcq6WF7p4":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWSJMSyjLGFJzkHRB7x7xp8jVfxCbAiT5oSNmEnQY6QbqB","ipfs_peer_addresses":["/ip4/192.168.0.86/tcp/4001/p2p/12D3KooWSJMSyjLGFJzkHRB7x7xp8jVfxCbAiT5oSNmEnQY6QbqB"],"status":"pinned","timestamp":"2022-11-21T01:33:20Z","error":"","attempt_count":0,"priority_pin":false}}}
	//{"cid":"QmXHFRWyW49vNzUrEo3Rr4wiU5LemBKYMCaG8gjB7oq7RN","name":"","allocations":["12D3KooWB6cxrTahCGu4T1vLeJTsSU3fHnWJsrig6bNi8afVEekm","12D3KooWRh8hkpcCL6kk5LtQ7Fz1Kj5npsDSdHQSpnnEcq6WF7p4","12D3KooWJ7b5LSbZJmRvrgGQVoSyVM6bTQdjtSc6cBpLWoZTQKXH"],"origins":[],"created":"2022-11-21T03:07:35Z","metadata":null,"peer_map":{"12D3KooWB6cxrTahCGu4T1vLeJTsSU3fHnWJsrig6bNi8afVEekm":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi","ipfs_peer_addresses":["/ip4/192.168.0.87/tcp/4001/p2p/12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi"],"status":"pinned","timestamp":"2022-11-21T03:07:35Z","error":"","attempt_count":0,"priority_pin":false},"12D3KooWJ7b5LSbZJmRvrgGQVoSyVM6bTQdjtSc6cBpLWoZTQKXH":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWRC61AVD34UNGCcdutEadWzetteZDwv26f55s1zfmhFgC","ipfs_peer_addresses":["/ip4/192.168.0.85/tcp/4001/p2p/12D3KooWRC61AVD34UNGCcdutEadWzetteZDwv26f55s1zfmhFgC"],"status":"pinned","timestamp":"2022-11-21T11:07:35+08:00","error":"","attempt_count":0,"priority_pin":false},"12D3KooWRh8hkpcCL6kk5LtQ7Fz1Kj5npsDSdHQSpnnEcq6WF7p4":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWSJMSyjLGFJzkHRB7x7xp8jVfxCbAiT5oSNmEnQY6QbqB","ipfs_peer_addresses":["/ip4/192.168.0.86/tcp/4001/p2p/12D3KooWSJMSyjLGFJzkHRB7x7xp8jVfxCbAiT5oSNmEnQY6QbqB"],"status":"pinned","timestamp":"2022-11-21T03:07:35Z","error":"","attempt_count":0,"priority_pin":false}}}
	//{"cid":"QmahnmYLjUWX8ek8oRMwmTeg7WnMY61foQdPf18VwWfaLY","name":"","allocations":["12D3KooWB6cxrTahCGu4T1vLeJTsSU3fHnWJsrig6bNi8afVEekm","12D3KooWRh8hkpcCL6kk5LtQ7Fz1Kj5npsDSdHQSpnnEcq6WF7p4","12D3KooWJ7b5LSbZJmRvrgGQVoSyVM6bTQdjtSc6cBpLWoZTQKXH"],"origins":[],"created":"2022-11-21T02:53:09Z","metadata":null,"peer_map":{"12D3KooWB6cxrTahCGu4T1vLeJTsSU3fHnWJsrig6bNi8afVEekm":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi","ipfs_peer_addresses":["/ip4/192.168.0.87/tcp/4001/p2p/12D3KooWK3yhTEZvvzZq5LM8FtvtAs2oX7iEj4Vjycx8eBdJRCQi"],"status":"pinned","timestamp":"2022-11-21T02:53:09Z","error":"","attempt_count":0,"priority_pin":false},"12D3KooWJ7b5LSbZJmRvrgGQVoSyVM6bTQdjtSc6cBpLWoZTQKXH":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWRC61AVD34UNGCcdutEadWzetteZDwv26f55s1zfmhFgC","ipfs_peer_addresses":["/ip4/192.168.0.85/tcp/4001/p2p/12D3KooWRC61AVD34UNGCcdutEadWzetteZDwv26f55s1zfmhFgC"],"status":"pinned","timestamp":"2022-11-21T10:53:09+08:00","error":"","attempt_count":0,"priority_pin":false},"12D3KooWRh8hkpcCL6kk5LtQ7Fz1Kj5npsDSdHQSpnnEcq6WF7p4":{"peername":"localhost.localdomain","ipfs_peer_id":"12D3KooWSJMSyjLGFJzkHRB7x7xp8jVfxCbAiT5oSNmEnQY6QbqB","ipfs_peer_addresses":["/ip4/192.168.0.86/tcp/4001/p2p/12D3KooWSJMSyjLGFJzkHRB7x7xp8jVfxCbAiT5oSNmEnQY6QbqB"],"status":"pinned","timestamp":"2022-11-21T02:53:09Z","error":"","attempt_count":0,"priority_pin":false}}}

	url := buildClusterUrl(p.Hosts[0], "/pins?local=false")
	resp := callApi(http.MethodGet, url)
	if resp == nil {
		return
	}

	var cidInfo CidInfo
	var cids []string
	lines := strings.Split(string(resp), "\n")
	for _, line := range lines {
		if len(line) <= 1 {
			continue
		}

		e := json.Unmarshal([]byte(line), &cidInfo)
		if e != nil {
			logger.Error(fmt.Sprintf("parse CidInfo err:%s", e.Error()))
			continue
		}

		cids = append(cids, cidInfo.Cid)
	}

	logger.Info(fmt.Sprintf("cluster track %d cids", len(cids)))
	if cidDetail {
		for _, cid := range cids {
			logger.Info(cid)
		}
	}
}

func ipfsPeersInfo(nodeDetail bool) (swarmPeers SwarmPeers, repoStats []RepoStat) {
	swarmPeers = ipfsSwarmPeers()

	if nodeDetail {
		for _, peer := range swarmPeers.Peers {
			repoStat := ipfsRepoStat(peer.Addr)
			repoStats = append(repoStats, repoStat)
		}
	}

	return
}

func clusterInfo(nodeDetail, cidDetail, print bool) (swarmPeers SwarmPeers, repoStats []RepoStat) {
	clusterPins(cidDetail)

	swarmPeers, repoStats = ipfsPeersInfo(nodeDetail)

	if print {
		if len(repoStats) > 0 {
			for i := range swarmPeers.Peers {
				fmt.Sprintf("%s-%s used/total: (%02fGB/%02fGB), objects:%d",
					swarmPeers.Peers[i].Addr,
					swarmPeers.Peers[i].Peer,
					float32(repoStats[i].RepoSize)/float32(GB),
					float32(repoStats[i].StorageMax)/float32(GB),
					repoStats[i].NumObjects,
				)
				logger.Info(fmt.Sprintf("%+v:\t%+v", swarmPeers.Peers[i], repoStats[i]))
			}
		} else {
			logger.Info(fmt.Sprintf("%+v", swarmPeers))
		}
	}

	return
}

func clusterGc() time.Duration {
	url := buildClusterUrl(p.Hosts[0], "/ipfs/gc?local=false")

	st := time.Now()
	logger.Info(fmt.Sprintf("gc started at: %s", st.String()))

	callApi(http.MethodPost, url)

	et := time.Now()
	logger.Info(fmt.Sprintf("gc finished at: %s", et.String()))
	logger.Info(fmt.Sprintf("gc time used:%s", et.Sub(st).String()))

	return et.Sub(st)
}

func gc() error {
	sp1, rs1 := clusterInfo(true, false, true)
	d := clusterGc()
	sp2, rs2 := clusterInfo(true, false, true)

	for i := range sp1.Peers {
		for j := range sp2.Peers {
			if sp1.Peers[i].Addr == sp2.Peers[i].Addr {
				s := float64(rs1[i].RepoSize-rs2[j].RepoSize) / float64(GB)
				os := rs1[i].NumObjects - rs2[j].NumObjects
				logger.Info(
					fmt.Sprintf("node[%s] gc repo size:%02f, objects:%d, gc speed: %02fGB/s, %02fObjects/s",
						sp1.Peers[i].Addr,
						s,
						os,
						s/d.Seconds(),
						float64(os)/d.Seconds(),
					),
				)
			}
		}
	}
	return nil
}
