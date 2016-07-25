#!/bin/bash
set -e
set -x

#
# This script helps a contributor to cut a new release of resourced-master.
#
# Prerequisites:
# - Ensure you(contributor) has Go 1.6.x or newer.
# - Ensure godep is installed.
#
# Arguments:
# $VERSION: semantic version number
#

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=$(dirname $CURRENT_DIR)

cd $ROOT_DIR

cp -r tests/config-files conf

GOOS=darwin godep go build
tar cvzf resourced-master-darwin-$VERSION.tar.gz resourced-master static/ templates/ migrations/ conf/

GOOS=linux godep go build
tar cvzf resourced-master-linux-$VERSION.tar.gz resourced-master static/ templates/ migrations/ conf/

rm -rf $ROOT_DIR/conf

rm -f $ROOT_DIR/resourced-master