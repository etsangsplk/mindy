#! /bin/bash

if [ "$PREFIX" = "" ]; then
	PREFIX=mindy_test_
fi

cd $GOPATH/src/github.com/eris-ltd/mindy
VC=./test/plumbing
UPDATE_EVERY=5 

# $PREFIX to allow namespacing instances of the app. Default is empty

# run tendermint (with data container)
docker run --name "${PREFIX}mintdata" eris/mint mkdir -p /home/eris/.eris/blockchains/tendermint
if [ "$VC" != "" ]; then
	echo "copy data into eris/mint contract"
	cd $VC
	tar cf - . | docker run -i --rm --volumes-from "${PREFIX}mintdata" --user eris eris/mint tar xvf - -C /home/eris/.eris/blockchains/tendermint
fi
echo "###################### RUN eris/mint container #########################"
docker run --name "${PREFIX}mint" --volumes-from "${PREFIX}mintdata" -d -p 46657:46657 -e FAST_SYNC=$FAST_SYNC eris/mint

# run tinydns
echo "###################### RUN tinydns container #########################"
docker run --name "${PREFIX}tinydns" -d -p 53:53/udp "${PREFIX}tinydns"

# run mindy
echo "###################### RUN mindy container #########################"
docker run -d --name "${PREFIX}mindy" --volumes-from "${PREFIX}tinydns" --link "${PREFIX}mint":mint -e UPDATE_EVERY=$UPDATE_EVERY "${PREFIX}mindy"


echo "###################### RUN the mindy test #########################"
docker run --rm -t --link "${PREFIX}mint":mint --link "${PREFIX}tinydns":tiny --name mindy_test mindy_test

echo "----------------------------------"
echo "cleanup ..."
docker rm -f "${PREFIX}mindy" "${PREFIX}tinydns" "${PREFIX}mint" "${PREFIX}mintdata"

