#!/bin/bash

set -e

# setup Go paths
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:${GOROOT}/bin:${GOPATH}/bin
echo "export GOROOT=$GOROOT" >> ~/.bashrc
echo "export GOPATH=$GOPATH" >> ~/.bashrc
echo "export PATH=$PATH" >> ~/.bashrc

# checkout agent code
mkdir -p ${GOPATH}/src/github.com/ligato
cd ${GOPATH}/src/github.com/ligato
git clone https://github.com/ligato/vpp-agent

# build the agent
cd ${GOPATH}/src/github.com/ligato/vpp-agent
git checkout $1
make install
