#!/bin/bash

etcdKey=""
while read line || [[ -n "$line" ]]; do
    if [ "${line:0:1}" != "/" ] ; then
        value="$(echo "$line"|tr -d '\r\n')"
        docker exec etcd etcdctl put $etcdKey "$value"
    else
        etcdKey="$(echo "$line"|tr -d '\r\n')"
    fi
done < "$1"
