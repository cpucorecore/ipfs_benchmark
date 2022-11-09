for thread in `cat threads`
do
echo "____thread${thread}_____"
./ipfs_benchmark -g ${thread} --to 20000 -w 100 --tag crdt --ts=false cluster add --bs $((1024*1024))
sleep 14400
date

./ipfs_benchmark -g 2 -w 500 --to 20000 --trf test_result/ClusterAdd_0-20000_g${thread}_bs1048576_r2-2_crdt.json --tag crdt --ts=false cluster pin rm
sleep 300

date
./ipfs_benchmark gc
sleep 720
echo "@@@@thread${i}@@@@"

done
