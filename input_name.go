package main

// POST		http://127.0.0.1:9094/pins/ipfs/QmWbhVQdcStoUT7d7U6t5NhHJxoevY3GuaBxVxbK8krKuU?mode=recursive&name=&replication-max=0&replication-min=0&shard-size=0&user-allocations=
// DELETE	http://127.0.0.1:9094/pins/ipfs/QmWbhVQdcStoUT7d7U6t5NhHJxoevY3GuaBxVxbK8krKuU
// GET		http://127.0.0.1:9094/pins/QmWbhVQdcStoUT7d7U6t5NhHJxoevY3GuaBxVxbK8krKuU?local=false
var httpInputNames = map[string]map[string]string{
	"POST": {
		"/api/v0/id":                "ipfs_id",
		"/api/v0/cat":               "ipfs_cat",
		"/api/v0/dag/stat":          "ipfs_dag_stat",
		"/api/v0/repo/stat":         "ipfs_repo_stat",
		"/api/v0/swarm/peers":       "ipfs_swarm_peers",
		"/api/v0/dht/findprovs":     "ipfs_dht_findprovs",
		"/api/v0/routing/findprovs": "ipfs_routing_findprovs",

		"/add":       "cluster_add",
		"/pins/ipfs": "cluster_pins_add",
	},
	"DELETE": {
		"/pins/ipfs": "cluster_pins_rm",
	},
	"GET": {
		"/pins": "cluster_pins_get",
	},
}

func getNameByHttpMethodAndPath(method, path string) string {
	name, ok := httpInputNames[method][path]
	if ok {
		return name
	} else {
		// TODO
		// e.g.:
		// POST /ipfs/v0/id		--> post-ipfs_id
		// DELETE /pins/ipfs	--> delete-pins_ipfs
		return ""
	}
}
