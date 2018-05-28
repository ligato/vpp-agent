#!/bin/bash

VPPDIR=${VPPDIR:-vpp}

find ${VPPDIR} -name \*.api -printf 'echo %p && ${VPPDIR}/src/tools/vppapigen/vppapigen --includedir ${VPPDIR}/src --input %p --output /usr/share/vpp/api/%f.json JSON\n' | xargs -0 sh -c
