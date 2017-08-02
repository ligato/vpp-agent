#!/bin/bash

source $(dirname "$0")/static_analysis.sh
SELECTOR="" static_analysis gofmt -w -s