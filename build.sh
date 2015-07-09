#! /bin/bash

# build and run mindy containers


# env vars for tendermint docker.sh
export $NO_BUILD
export $VD
export $VC
export $FAST_SYNC

# build and run tendermint
cd $GOPATH/src/github.com/tendermint/tendermint
./DOCKER/docker.sh

# build and run tinydns
cd $GOPATH/src/github.com/eris-ltd/mindy/tinydns
docker build -t tinydns .
docker run --name tinydns -d -p 53:53/udp tinydns

# build and run mindy
cd $GOPATH/src/github.com/eris-ltd/mindy
docker build -t mindy .
docker run -d --name mindy --volumes-from tinydns --link mint:mint  mindy 

