# gen file
./ipfs_benchmark -f 30200 -t 30300 -g 10 tool gen_file --size $((1024*1024))

# cluster add
./ipfs_benchmark -g 100 -f 0 --to 1000 api --host 127.0.0.1 -p 9094 -d=false cluster add --bs $((1024*1024)) -p -r 3

# cluster pin get
./ipfs_benchmark -g 100 --to 1000 --sc=true api --host 127.0.0.1 -p 9094 -d=false cluster pin --tr tests/report/cluster_add_g100_s-true_from0_to1000_bs1048576_replica3_pin-true.json get

# cluster pin rm
./ipfs_benchmark -g 100 --to 1000 --sc=true api --host 127.0.0.1 -p 9094 -d=false cluster pin --tr tests/report/cluster_add_g100_s-true_from0_to1000_bs1048576_replica3_pin-true.json rm

# cluster pin add
./ipfs_benchmark -g 100 --to 1000 --sc=true api --host 127.0.0.1 -p 9094 -d=false cluster pin --tr tests/report/cluster_add_g100_s-true_from0_to1000_bs1048576_replica3_pin-true.json add -r 1

# cluster unpin by cid
./ipfs_benchmark -g 100 --to 1000 --sc=true api --host 127.0.0.1 -p 9094 -d=false cluster unpin_by_cid -c cids

# ipfs dht findprovs
./ipfs_benchmark -g 100 --sc=true -f 0 -t 1000 api --host 127.0.0.1 -p 5001 -d=false ipfs iter_test --tr tests/report/cluster_add_g100_s-true_from0_to1000_bs1048576_replica3_pin-true.json dht_findprovs

# ipfs dag stat
./ipfs_benchmark -g 100 --sc=true -f 0 -t 1000 api --host 127.0.0.1 -p 5001 -d=false ipfs iter_test --tr tests/report/cluster_add_g100_s-true_from0_to1000_bs1048576_replica3_pin-true.json dag_stat -p=false

# ipfs cat
./ipfs_benchmark -g 100 --sc=true -f 0 -t 1000 api --host 127.0.0.1 -p 5001 -d=false ipfs iter_test --tr tests/report/cluster_add_g100_s-true_from0_to1000_bs1048576_replica3_pin-true.json cat

# ipfs id
./ipfs_benchmark -g 100 --sc=true -f 0 -t 1000 api --host 127.0.0.1 -p 5001 -d=false ipfs repeat_test -r 1000 id

# compare
./ipfs_benchmark -f 0 -t 100000 tool compare --tag test tests/report/ipfs_id_g100_s-true_repeat1000.json tests/report/ipfs_swarm_peers_g100_s-true_repeat1000.json
