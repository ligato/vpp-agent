#!/bin/bash

VPPDIR=${VPPDIR:-vpp}
APIDIR=${APIDIR:-/usr/share/vpp/api}

mkdir -p ${APIDIR}

find ${VPPDIR} -name \*.api -printf "echo %p - ${APIDIR}/%f.json && ${VPPDIR}/src/tools/vppapigen/vppapigen --includedir ${VPPDIR}/src --input %p --output ${APIDIR}/%f.json JSON\n" | xargs -0 sh -c
