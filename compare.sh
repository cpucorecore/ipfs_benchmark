tests_dir=$1

for item in `cat compare_items`
do

nm=`echo $item | awk -F ',' '{print $1}'`
total=`echo $item | awk -F ',' '{print $2}'`
from_p=`echo $item | awk -F ',' '{print $3}'`
to_p=`echo $item | awk -F ',' '{print $4}'`
from=$((total*from_p/100))
to=$((total*to_p/100))

./ipfs_benchmark --from $from --to $to tool compare --tag ${tests_dir}_${nm} `ls ${tests_dir}/report/${nm}*` && echo

done
