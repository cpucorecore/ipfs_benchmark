# 200G
echo "200G add"
./ipfs_benchmark --from 10000 --to 20000 -g 32 --sc=true api --host 192.168.0.87 -p 9094 --max_retry 30 --dto 300 --rto 1200 -d=false --tag crdt_200G cluster add --bs $((1024*1024)) -r 3 -p

sleep 6000

echo "200G repeat repo stat"
./ipfs_benchmark -g 20 --sc=true api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_200G ipfs repeat_test -r 500 repo_stat

sleep 120

echo "200G cat"
./ipfs_benchmark -g 100 --sc=true --from 0 --to 10000 api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=true --tag crdt_200G ipfs iter_test -tr tests/report/cluster_add_g32_sc-true_from10000_to20000_bs1048576_replica3_pin-true_crdt_200G.json cat

sleep 120

echo "200G dag stat"
./ipfs_benchmark -g 100 --sc=true --from 0 --to 10000 api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_200G ipfs iter_test -tr tests/report/cluster_add_g32_sc-true_from10000_to20000_bs1048576_replica3_pin-true_crdt_200G.json dag_stat

sleep 120

echo "200G dht findprovs"
./ipfs_benchmark -g 100 --sc=true --from 0 --to 10000 api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_200G ipfs iter_test -tr tests/report/cluster_add_g32_sc-true_from10000_to20000_bs1048576_replica3_pin-true_crdt_200G.json dht_findprovs

echo "200G repeat id"
./ipfs_benchmark -g 100 --sc=true api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_200G ipfs repeat_test -r 500 id

echo "200G repeat swarm peers"
./ipfs_benchmark -g 100 --sc=true api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_200G ipfs repeat_test -r 500 swarm_peers

echo "200G cluster pin get"
./ipfs_benchmark --from 0 --to 10000 -g 100 --sc=true api --host 192.168.0.87 -p 9094 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_200G cluster pin --tr tests/report/cluster_add_g32_sc-true_from10000_to20000_bs1048576_replica3_pin-true_crdt_200G.json get

echo "200G cluster pin rm"
./ipfs_benchmark --from 0 --to 10000 -g 10 --sc=true api --host 192.168.0.87 -p 9094 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_200G cluster pin --tr tests/report/cluster_add_g32_sc-true_from10000_to20000_bs1048576_replica3_pin-true_crdt_200G.json rm

sleep 3600

echo "200G cluster pin add"
./ipfs_benchmark --from 0 --to 10000 -g 10 --sc=true api --host 192.168.0.87 -p 9094 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_200G cluster pin --tr tests/report/cluster_add_g32_sc-true_from10000_to20000_bs1048576_replica3_pin-true_crdt_200G.json add -r 3

sleep 10800

# 300G
echo add
./ipfs_benchmark --from 20000 --to 30000 -g 32 --sc=true api --host 192.168.0.87 -p 9094 --max_retry 30 --dto 300 --rto 1200 -d=false --tag crdt_300G cluster add --bs $((1024*1024)) -r 3 -p

sleep 6000

echo "300G repeat repo stat"
./ipfs_benchmark -g 20 --sc=true api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_300G ipfs repeat_test -r 500 repo_stat

sleep 120

echo "300G cat"
./ipfs_benchmark -g 100 --sc=true --from 0 --to 10000 api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=true --tag crdt_300G ipfs iter_test -tr tests/report/cluster_add_g32_sc-true_from20000_to30000_bs1048576_replica3_pin-true_crdt_300G.json cat

sleep 120

echo "300G dag stat"
./ipfs_benchmark -g 100 --sc=true --from 0 --to 10000 api --host 192.168.0.87 -p 5001 --dto 300 --rto 1200 --max_retry 20 -d=false --tag crdt_300G ipfs iter_test -tr tests/report/cluster_add_g32_sc-true_from20000_to30000_bs1048576_replica3_pin-true_crdt_300G.json dag_stat
