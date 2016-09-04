#!/usr/bin/env bash
curr_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
datadir=$curr_dir/datadir
logdir=$curr_dir/logs
bindir=$curr_dir/bin
echo "datadir : "$datadir
echo " logdir : "$logdir
echo " bindir : "$bindir
mkdir -p $datadir
mkdir -p $logdir
mkdir -p $bindir

etcd=$bindir/etcd

if [ ! -f $etcd ]; then
    echo $etcd" :: File not found - dowloading ..."
    cd $bindir
    curl -L https://github.com/coreos/etcd/releases/download/v3.0.7/etcd-v3.0.7-linux-amd64.tar.gz -o etcd-v3.0.7-linux-amd64.tar.gz
    tar xzvf etcd-v3.0.7-linux-amd64.tar.gz 
    mv etcd-v3.0.7-linux-amd64/etcd .
    rm etcd-v3.0.7-linux-amd64.tar.gz
    cd $curr_dir
fi

etcd_cmd=$etcd" --data-dir $datadir"
$etcd_cmd > ${logdir}/etcd.log 2>&1 &
