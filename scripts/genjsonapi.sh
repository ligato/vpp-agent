#!/bin/bash

find ./vpp -name \*.api -printf 'echo %p && ./vpp/src/tools/vppapigen/vppapigen --includedir ./vpp/src --input %p --output /usr/share/vpp/api/%f.json JSON\n' | xargs -0 sh -c
