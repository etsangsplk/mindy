#! /bin/bash

# run tendermint (with data container)
docker run --name mintdata eris/mint mkdir -p /home/eris/.eris/blockchains/tendermint
if [ "$VC" != "" ]; then
	cd $VC
	tar cf - . | docker run -i --rm --volumes-from mintdata --user eris eris/mint tar xvf - -C /home/eris/.eris/blockchains/tendermint
fi
docker run --name mint --volumes-from mintdata -d -p 46657:46657 -e FAST_SYNC=$FAST_SYNC eris/mint

# build and run tinydns
cd $GOPATH/src/github.com/eris-ltd/mindy/tinydns
docker build -t tinydns .
docker run --name tinydns -d -p 53:53/udp tinydns

# build and run mindy
cd $GOPATH/src/github.com/eris-ltd/mindy
docker build -t mindy .
docker run -d --name mindy --volumes-from tinydns --link mint:mint -e UPDATE_EVERY=$UPDATE_EVERY mindy 

