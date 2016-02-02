#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for mindy to be ran from circle ci.
# It will run the testing sequence for mindy using docker.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally
# docker-machine installed locally
# eris installed locally

# ----------------------------------------------------------
# USAGE

# circle_test.sh

# ----------------------------------------------------------
# Set defaults

uuid=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
machine="eris-test-mindy-$uuid"
ver=$(cat version/version.go | tail -n 1 | cut -d ' ' -f 4 | tr -d '"')
start=`pwd`

# ----------------------------------------------------------
# Get machine sorted

echo "Setting up a Machine for EPM Testing"
docker-machine create --driver amazonec2 $machine 1>/dev/null
eval $(docker-machine env $machine)
echo "Machine setup."
echo
docker version
echo
echo "Pulling needed images"
docker pull quay.io/eris/data 1>/dev/null
docker pull quay.io/eris/keys 1>/dev/null
docker pull quay.io/eris/erisdb:$ver 1>/dev/null
echo "Pulling finished."
echo

# ----------------------------------------------------------
# Run tests

tests/test.sh
test_exit=$?

# ----------------------------------------------------------
# Clenup

echo
echo
echo "Cleaning up"
docker-machine rm --force $machine
cd $start
exit $test_exit