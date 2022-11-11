to=1000
tag=crdt
bs=$((1024*1024))

for thread in `cat threads`
do
echo "____thread${thread}_____"

./ipfs_benchmark -g ${thread} --to ${to} -w 100 --tag ${tag} cluster add --bs ${bs}

sleep 14400
date


./ipfs_benchmark --tag crdt --to ${to} -w 100 -g 2 cluster pin --trf tests/reports/ClusterAdd_0-${to}_g${thread}_bs${bs}_r2-2_${tag}.json rm
sleep 3000

date
./ipfs_benchmark gc
sleep 720
echo "@@@@thread${thread}@@@@"

done
