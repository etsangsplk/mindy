#! /bin/bash

export ROOT=./test/porcelain

# build tinydns
echo "###################### BUILD tinydns container #########################"
cd $GOPATH/src/github.com/eris-ltd/mindy/tinydns
docker build -t eris/tinydns .

# build mindy
echo "###################### BUILD mindy container #########################"
cd $GOPATH/src/github.com/eris-ltd/mindy
docker build -t eris/mindy .

# build the test container
echo "###################### BUILD mindy test container #########################"
cd $GOPATH/src/github.com/eris-ltd/mindy
docker build -t mindy_test -f $ROOT/Dockerfile .

cd $ROOT
bash run.sh

