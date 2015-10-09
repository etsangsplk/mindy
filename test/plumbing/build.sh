#! /bin/bash
set -e

PREFIX=mindy_test_

# build tinydns
echo "###################### BUILD tinydns container #########################"
cd $GOPATH/src/github.com/eris-ltd/mindy/tinydns
docker build -t "${PREFIX}tinydns" .

# build mindy
echo "###################### BUILD mindy container #########################"
cd $GOPATH/src/github.com/eris-ltd/mindy
docker build -t "${PREFIX}mindy" .

# build the test container
echo "###################### BUILD mindy test container #########################"
cd $GOPATH/src/github.com/eris-ltd/mindy
docker build -t mindy_test -f test/plumbing/Dockerfile .

cd ./test/plumbing
PREFIX=$PREFIX bash run.sh

