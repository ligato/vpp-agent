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

# install golint, gvt & Glide
go get -u github.com/golang/lint/golint
go get -u github.com/FiloSottile/gvt
curl https://glide.sh/get | sh

# checkout agent code
mkdir -p $GOPATH/src/gitlab.cisco.com/ctao
cd $GOPATH/src/gitlab.cisco.com/ctao
git clone http://gitlab.cisco.com/ctao/vnf-agent

# build the agent
cd $GOPATH/src/gitlab.cisco.com/ctao/vnf-agent
git checkout $1
make
make install
#make test
#make generate
