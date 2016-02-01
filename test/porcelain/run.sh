#! /bin/bash

export ERIS_PULL_APPROVE="true"
export ERIS_MIGRATE_APPROVE="true"
export UPDATE_EVERY=5 
ROOT=/home/eris/.eris/mindy/test/porcelain

# init the eris cli
eris init --yes --testing=true

# create new blockchain with files in $ROOT
eris chains new --dir $ROOT mindy_test

# cat the new chains genesis 
eris chains plop mindy_test genesis

# TODO: UPDATE_EVERY in the env
eris services start mindy --chain=mindy_test --debug --publish

echo "###################### RUN the mindy test #########################"
docker run --rm -t --link eris_chain_mindy_test_1:mint --link eris_service_tinydns_1:tiny --name mindy_test mindy_test

echo "----------------------------------"
echo "cleanup ..."
# eris services stop -rx mindy --chain=mindy_test ## TODO: fix
docker rm -vf eris_service_tinydns_1 eris_service_mindy_1 eris_chain_mindy_test_1 eris_service_keys_1 eris_data_keys_1
eris data rm tinydns mindy_test

