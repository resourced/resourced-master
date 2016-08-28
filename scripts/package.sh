#!/bin/bash
set -e
set -x

#
# This script helps a contributor to cut a new release of resourced-master.
#
# Prerequisites:
# - Ensure you(contributor) has Go 1.6.x or newer.
# - Ensure govendor is installed.
#
# Arguments:
# $VERSION: semantic version number (required)
#

: "${VERSION?You must set VERSION}"

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=$(dirname $CURRENT_DIR)

cd $ROOT_DIR

git checkout tests/config-files/*.toml

cp -r tests/config-files conf
rm -rf conf/config-files

govendor add +external

GOOS=darwin go build
tar cvzf resourced-master-darwin-$VERSION.tar.gz resourced-master static/ templates/ migrations/ conf/

GOOS=linux go build
tar cvzf resourced-master-linux-$VERSION.tar.gz resourced-master static/ templates/ migrations/ conf/

rm -rf $ROOT_DIR/conf

rm -f $ROOT_DIR/resourced-master