#!/bin/bash

set +e
sudo docker rmi -f dev_vpp_agent 2>/dev/null
set -e

CURRENT_FOLDER=`pwd`
AGENT_COMMIT=`git rev-parse HEAD`
cd ../..
VPP_COMMIT=`git submodule status|grep vpp|cut -c2-41`
cd $CURRENT_FOLDER
echo "repo agent commit number: "$AGENT_COMMIT
echo "repo vpp commit number: "$VPP_COMMIT

while [ "$1" != "" ]; do
    case $1 in
        -a | --agent )          shift
                                AGENT_COMMIT=$1
                                ;;
        -v | --vpp )            shift
                                VPP_COMMIT=$1
                                ;;
        * )                     echo "invalid parameter "$1
                                exit 1
    esac
    shift
done

echo "build agent commit number: "$AGENT_COMMIT
echo "build vpp commit number: "$VPP_COMMIT

sudo docker build -t dev_vpp_agent --build-arg AGENT_COMMIT=$AGENT_COMMIT --build-arg VPP_COMMIT=$VPP_COMMIT --no-cache .

