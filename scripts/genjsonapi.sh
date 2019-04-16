#!/bin/bash

VPPDIR=${VPPDIR:-/opt/vpp}
APIDIR=${APIDIR:-/usr/share/vpp/api}

VPPAPIGEN=${VPPDIR}/src/tools/vppapigen/vppapigen

[ -d "${VPPDIR}" ] || {
    echo >&2 "vpp directory not found at: ${VPPDIR}";
    exit 1;
}
mkdir -p ${APIDIR}/core ${APIDIR}/plugins

APIDST=${APIDIR}/core

find ${VPPDIR}/src -name \*.api -not -path '*src/plugins/*' \
	-printf "echo %p - ${APIDST}/%f.json \
    && ${VPPAPIGEN} --includedir ${VPPDIR}/src \
    --input %p --output ${APIDST}/%f.json JSON\n" | xargs -0 sh -c

APIDST=${APIDIR}/plugins

find ${VPPDIR}/src/plugins -name \*.api \
	-printf "echo %p - ${APIDST}/%f.json \
    && ${VPPAPIGEN} --includedir ${VPPDIR}/src \
    --input %p --output ${APIDST}/%f.json JSON\n" | xargs -0 sh -c
