#!/bin/sh

go build -v
MICROSERVICE_LABEL=example-agent sudo -E ./localclient_with_etcd
