# gen file
./ipfs_benchmark -f 30200 -t 30300 -g 10 tool gen_file --size $((1024*1024))

# cluster gc
./ipfs_benchmark api --hosts 127.0.0.1 -p 9094 cluster gc

# cluster add
./ipfs_benchmark --from 0 --to 10000 -g 3 --sc=true api --hosts 192.168.0.87 -p 9094 --max_retry 6 --timeout 300 -d=false --tag crdt_100G cluster add --bs $((1024*1024)) -r 3 -p

# cluster pin get
./ipfs_benchmark --from 0 --to 100 -g 1 --sc=true api --hosts 127.0.0.1 -p 9094 -d=false --tag crdt cluster pin --tr tests/report/cluster_add_g10_sc-true_from0_to100_bs1048576_replica3_pin-true_crdt.json get

# cluster pin rm
./ipfs_benchmark --from 0 --to 100 -g 1 --sc=true api --hosts 127.0.0.1 -p 9094 -d=false --tag crdt cluster pin --tr tests/report/cluster_add_g10_sc-true_from0_to100_bs1048576_replica3_pin-true_crdt.json rm

# cluster pin add
./ipfs_benchmark --from 0 --to 100 -g 1 --sc=true api --hosts 127.0.0.1 -p 9094 -d=false --tag crdt cluster pin --tr tests/report/cluster_add_g10_sc-true_from0_to100_bs1048576_replica3_pin-true_crdt.json add -r 3

# cluster unpin by cid
./ipfs_benchmark -g 100 --to 1000 --sc=true api --hosts 127.0.0.1 -p 9094 -d=false cluster unpin_by_cid -c cids

# ipfs dht findprovs
./ipfs_benchmark --from 0 --to 100 -g 1 --sc=true api --hosts 127.0.0.1 -p 5001 -d=false --tag crdt ipfs iter_test --tr tests/report/cluster_add_g10_sc-true_from0_to100_bs1048576_replica3_pin-true_crdt.json dht_findprovs

# ipfs dag stat
./ipfs_benchmark --from 0 --to 100 -g 1 --sc=true api --hosts 127.0.0.1 -p 5001 -d=false --tag crdt ipfs iter_test --tr tests/report/cluster_add_g10_sc-true_from0_to100_bs1048576_replica3_pin-true_crdt.json dag_stat

# ipfs cat
./ipfs_benchmark --from 0 --to 100 -g 1 --sc=true api --hosts 127.0.0.1 -p 5001 -d=true --tag crdt ipfs iter_test --tr tests/report/cluster_add_g10_sc-true_from0_to100_bs1048576_replica3_pin-true_crdt.json cat

# ipfs id
./ipfs_benchmark -g 1 --sc=true api --hosts 127.0.0.1 -p 5001 -d=true --tag crdt ipfs repeat_test -r 10 id

# ipfs swarm_peers
./ipfs_benchmark -g 1 --sc=true api --hosts 127.0.0.1,192.168.0.85 -p 5001 -d=true --tag crdt ipfs repeat_test -r 10 swarm_peers

# compare
./ipfs_benchmark -f 0 -t 100000 tool compare --tag test tests/report/ipfs_id_g100_s-true_repeat1000.json tests/report/ipfs_swarm_peers_g100_s-true_repeat1000.json
