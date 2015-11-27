#! /bin/bash
set -e

REPO=github.com/eris-ltd/mindy
ROOT=$GOPATH/src/$REPO/test/porcelain
if [[ "$ERIS_VERSION" == "" ]]; then
	ERIS_VERSION=0.11
fi

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

# eris container in which we run the tests
docker run --name eris-data eris/data echo "Data-container for testing with eris-cli"

docker cp $GOPATH/src/$REPO/ eris-data:/home/eris/.eris/mindy/

docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock --volumes-from eris-data --entrypoint bash quay.io/eris/eris:$ERIS_VERSION /home/eris/.eris/mindy/test/porcelain/run.sh

docker rm -vf eris-data
