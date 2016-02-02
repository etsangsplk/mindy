#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for mind. It will build the tool
# into docker containers in a reliable and predicatable
# manner.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_tool.sh

# ----------------------------------------------------------
# Set defaults

if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  repo=$GOPATH/src/github.com/eris-ltd/mindy
fi

testimage=${testimage:="quay.io/eris/mindy"}
otherimage=${otherimage:="quay.io/eris/tinydns"}

cd $repo

# ---------------------------------------------------------------------------
# Go!

docker build -t $testimage:latest .
cd $repo/tinydns
docker build -t $otherimage:latest .
cd $repo