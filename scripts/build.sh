#!/bin/bash
set -x

#
# Call this script from the root of resourced-master.
# Arguments:
# $1: architecture
# $2: semantic version number

GOOS=$1 godep go build
tar cvzf resourced-master-$1-$2.tar.gz resourced-master static/ templates/ migrations/