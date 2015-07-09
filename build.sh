#! /bin/bash

# $PREFIX to allow namespacing instances of the app. Default is empty

# run tendermint (with data container)
docker run --name "${PREFIX}mintdata" eris/mint mkdir -p /home/eris/.eris/blockchains/tendermint
if [ "$VC" != "" ]; then
	cd $VC
	tar cf - . | docker run -i --rm --volumes-from "${PREFIX}mintdata" --user eris eris/mint tar xvf - -C /home/eris/.eris/blockchains/tendermint
fi
docker run --name "${PREFIX}mint" --volumes-from "${PREFIX}mintdata" -d -p 46657:46657 -e FAST_SYNC=$FAST_SYNC eris/mint

# build and run tinydns
cd $GOPATH/src/github.com/eris-ltd/mindy/tinydns
docker build -t "${PREFIX}tinydns" .
docker run --name "${PREFIX}tinydns" -d -p 53:53/udp "${PREFIX}tinydns"

# build and run mindy
cd $GOPATH/src/github.com/eris-ltd/mindy
docker build -t "${PREFIX}mindy" .
docker run -d --name "${PREFIX}mindy" --volumes-from "${PREFIX}tinydns" --link "${PREFIX}mint":mint -e UPDATE_EVERY=$UPDATE_EVERY "${PREFIX}mindy"

