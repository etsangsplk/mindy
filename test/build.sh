#! /bin/sh

cd $GOPATH/src/github.com/eris-ltd/mindy
VC=./test UPDATE_EVERY=5 ./build.sh

cd $GOPATH/src/github.com/eris-ltd/mindy
docker build -t mindy_test -f test/Dockerfile .
docker run --rm -t --link mint:mint --link tinydns:tiny --name mindy_test mindy_test

echo "----------------------------------"
echo "cleanup ..."
docker rm -f mindy tinydns mint mintdata
