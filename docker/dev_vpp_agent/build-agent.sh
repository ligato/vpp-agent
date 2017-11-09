#!/bin/bash

set -e

# setup Go paths
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
echo "export GOROOT=$GOROOT" >> ~/.bashrc
echo "export GOPATH=$GOPATH" >> ~/.bashrc
echo "export PATH=$PATH" >> ~/.bashrc
mkdir $GOPATH

# install gometalinter, golint, gvt & Glide
go get -u github.com/alecthomas/gometalinter
gometalinter --install
go get -u github.com/golang/lint/golint
go get -u github.com/FiloSottile/gvt
curl https://glide.sh/get | sh

# checkout agent code
mkdir -p $GOPATH/src/github.com/ligato
cd $GOPATH/src/github.com/ligato
git clone https://github.com/ligato/vpp-agent

# build the agent
cd $GOPATH/src/github.com/ligato/vpp-agent
git checkout $1
make
make install
#make test
#make generate
