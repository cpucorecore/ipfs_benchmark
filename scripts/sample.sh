./ipfs_benchmark -p 5001 -v -g 10 -sc=true ipfs -m POST -p "/api/v0/id" repeat_test -r 10
 /ipfs_benchmark -v -g 10 --sc=true cluster add --to 1000 -s $((1024*1024)) -r 1
./ipfs_benchmark -p 5001 -g 3  --sc=true ipfs -m POST -p "/api/v0/dht/findprovs" iter_test --trf tests/report/cluster_add.json
./ipfs_benchmark -g 100 --sc=true cluster pin --trf tests/report/cluster_add.json get
./ipfs_benchmark -g 100 --sc=true cluster pin --trf tests/report/cluster_add.json rm
./ipfs_benchmark -g 100 --sc=true cluster pin --trf tests/report/cluster_add.json add --replica 1
