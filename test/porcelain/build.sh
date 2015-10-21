#! /bin/bash
set -e

REPO=github.com/eris-ltd/mindy
ROOT=$GOPATH/src/$REPO/test/porcelain

# build tinydns
echo "###################### BUILD tinydns container #########################"
cd $GOPATH/src/$REPO/tinydns
docker build -t eris/tinydns .

# build mindy
echo "###################### BUILD mindy container #########################"
cd $GOPATH/src/$REPO
docker build -t eris/mindy .

# build the test container
echo "###################### BUILD mindy test container #########################"
cd $GOPATH/src/$REPO
docker build -t mindy_test -f $ROOT/Dockerfile .

cd $GOPATH/src/$REPO

# eris/eris container in which we run the tests
docker run -it --rm -v $GOPATH/src/$REPO:/go/src/$REPO -v /var/run/docker.sock:/var/run/docker.sock --entrypoint bash quay.io/eris/eris /go/src/$REPO/test/porcelain/run.sh
