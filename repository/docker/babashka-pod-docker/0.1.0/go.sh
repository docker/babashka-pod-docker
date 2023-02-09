#!/bin/sh

export ATOMIST_LOG_LEVEL=warn; $(dirname "$0")/babashka-pod-docker
