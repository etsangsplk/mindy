#! /bin/bash

cd $GOPATH/src/github.com/eris-ltd/mindy
UPDATE_EVERY=5 

if [ "$ROOT" = "" ]; then
	ROOT=./test/porcelain
fi



eris chains new --dir $ROOT mindy_test

# XXX: UPDATE_EVERY in the env
eris services start mindy --chain=mindy_test --debug


echo "###################### RUN the mindy test #########################"
docker run --rm -t --link eris_chain_mindy_test_1:mint --link eris_service_tinydns_1:tiny --name mindy_test mindy_test

echo "----------------------------------"
echo "cleanup ..."
# eris services stop -rx mindy --chain=mindy_test ## TODO: fix
docker rm -vf eris_service_tinydns_1 eris_service_mindy_1 eris_chain_mindy_test_1
eris data rm tinydns mindy_test

