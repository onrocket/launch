#!/usr/bin/env bash
curr_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"             
bindir=${curr_dir}/bin                                                          
etcd=$bindir/etcd     

kill `ps -ef | grep "$etcd" | grep -v grep | awk '{print $2}'`
