./ipfs_benchmark -g 10 -v --sc=true gen_file --to 1000

 /ipfs_benchmark -v -g 10 --sc=true cluster add --to 1000 -s $((1024*1024)) -r 1
./ipfs_benchmark --host 192.168.0.220 -v -g 20 --sc=true cluster add --to 1000 -s $((1024*1024)) -r 1

./ipfs_benchmark -p 5001 -g 3  --sc=true ipfs -m POST -p "/api/v0/dht/findprovs" iter_test --trf tests/report/cluster_add.json
./ipfs_benchmark --host 192.168.0.220 -p 5001 -g 3  --sc=true ipfs -m POST -p "/api/v0/dht/findprovs" iter_test --trf tests/report/cluster_add_v-true_g-20_sync-true_from-0_to-1000_block_size-1048576_replica-1_pin-true.json

./ipfs_benchmark -g 100 --sc=true cluster pin --trf tests/report/cluster_add.json get
./ipfs_benchmark -g 100 --sc=true cluster pin --trf tests/report/cluster_add.json rm
./ipfs_benchmark -g 100 --sc=true cluster pin --trf tests/report/cluster_add.json add --replica 1

./ipfs_benchmark -p 5001 -v -g 10 -sc=true ipfs -m POST -p "/api/v0/id" repeat_test -r 10
