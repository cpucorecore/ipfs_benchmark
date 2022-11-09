fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=4K -rw=randwrite -numjobs=1
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=4K -rw=randwrite -numjobs=2
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=4K -rw=randwrite -numjobs=3
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=4K -rw=randwrite -numjobs=4
echo
echo
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=8K -rw=randwrite -numjobs=1
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=16K -rw=randwrite -numjobs=1
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=64K -rw=randwrite -numjobs=1
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=128K -rw=randwrite -numjobs=1
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=256K -rw=randwrite -numjobs=1
fio -direct=1 -group_reporting -name=rw -ioengine=psync -runtime=120 -size=4G -bs=512K -rw=randwrite -numjobs=1
