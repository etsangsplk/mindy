#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for mindy. It will run the testing
# sequence for mindy using docker.

# ----------------------------------------------------------
# REQUIREMENTS

# eris installed locally

# ----------------------------------------------------------
# USAGE

# test.sh

# ----------------------------------------------------------
# Set defaults

# Where are the Things?

name=mindy
base=github.com/eris-ltd/$name
repo=$GOPATH/src/$base
if [ "$CIRCLE_BRANCH" ]
then
  repo=${GOPATH%%:*}/src/$base
  ci=true
  linux=true
elif [ "$TRAVIS_BRANCH" ]
then
  ci=true
  osx=true
elif [ "$APPVEYOR_REPO_BRANCH" ]
then
  ci=true
  win=true
else
  ci=false
fi

branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}
branch=${branch/\//_}

# Other variables
if [[ "$(uname -s)" == "Linux" ]]
then
  uuid=$(cat /proc/sys/kernel/random/uuid | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
elif [[ "$(uname -s)" == "Darwin" ]]
then
  uuid=$(uuidgen | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1  | tr '[:upper:]' '[:lower:]')
else
  uuid="62d1486f0fe5"
fi

was_running=0
test_exit=0
chains_dir=$HOME/.eris/chains
chain_name=mindy-tests-$uuid
name_full="$chain_name"_full_000
chain_dir=$chains_dir/$chain_name

export ERIS_PULL_APPROVE="true"
export ERIS_MIGRATE_APPROVE="true"

# ---------------------------------------------------------------------------
# Needed functionality

ensure_running(){
  if [[ "$(eris services ls -qr | grep $1)" == "$1" ]]
  then
    echo "$1 already started. Not starting."
    was_running=1
  else
    echo "Starting service: $1"
    eris services start $1 1>/dev/null
    early_exit
    sleep 3 # boot time
  fi
}

early_exit(){
  if [ $? -eq 0 ]
  then
    return 0
  fi

  echo "There was an error duing setup; keys were not properly imported. Exiting."
  if [ "$was_running" -eq 0 ]
  then
    if [ "$ci" = true ]
    then
      eris services stop keys
    else
      eris services stop -rx keys
    fi
  fi
  exit 1
}

test_setup(){
  echo "Getting Setup"
  if [ "$ci" = true ]
  then
    eris init --yes --pull-images=true --testing=true 1>/dev/null
  fi
  ensure_running keys

  # make a chain
  eris chains make --account-types=Full:1 mindy-tests-$uuid 1>/dev/null
  key1_addr=$(cat $chain_dir/accounts.csv | grep $name_full | cut -d ',' -f 1)
  echo -e "Default Key =>\t\t\t\t$key1_addr"
  eris chains new $chain_name --dir $chain_dir 1>/dev/null
  sleep 5 # boot time
  echo "Setup complete"
}

perform_tests(){
  echo
  eris services start tinydns --publish
  eris services start mindy --chain=$chain_name
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi

  # get mint ip for dns registration
  IP=$(eris services inspect keys NetworkSettings.IPAddress)
  tiny=$(eris services inspect tinydns NetworkSettings.IPAddress)
  chain_ip=$(eris chains inspect $chain_name NetworkSettings.IPAddress)
  echo -e "IP is =>\t\t\t\t$IP"
  echo

  DATA="[{\"fqdn\":\"interblock.io\", \"address\":\"$IP\", \"type\":\"NS\"}, {\"fqdn\":\"interblock.io\", \"address\":\"$IP\", \"type\":\"A\"}]"
  echo -e "Setting Data =>\t\t\t\t$DATA"
  eris chains exec $chain_name -- mintx name --chainID=$chain_name --sign-addr=$IP:4767 --node-addr=$chain_ip:46657 --pubkey=$key1_addr --name=interblock.io --data="$DATA" --amt=100000 --fee=0 --sign --broadcast 1>/dev/null
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi

  DATA="{\"fqdn\":\"pinkpenguin.interblock.io\", \"address\":\"$IP\", \"type\":\"A\"}"
  echo -e "Setting Data =>\t\t\t\t$DATA"
  eris chains exec $chain_name -- mintx name --chainID=$chain_name --sign-addr=$IP:4767 --node-addr=$chain_ip:46657 --pubkey=$key1_addr --name=pinkpgenuin.interblock.io --data="$DATA" --amt=100000 --fee=0  --sign --broadcast --wait 1>/dev/null
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi
  echo

  sleep 15 # let tinydns get updated

  if [ $ci = true ]
  then
    tiny=$(docker-machine ip $(docker-machine active))
    port=$(eris services inspect tinydns NetworkSettings.Ports | cut -d ' ' -f 4 | sed 's/[^0-9]*//g')
  fi

  # check
  if [ $ci = true ]
  then
    A1=`dig +short @$tiny -p $port interblock.io`
  else
    A1=`dig +short @$tiny interblock.io`
  fi
  echo -e "First record =>\t\t\t\t$A1"
  if [ "$A1" != "$IP" ]; then
    echo "Resolved wrong ip for interblock.io. Got $A1, expected $IP"
    test_exit=1
    return 1
  fi
  if [ $ci = true ]
  then
    A2=`dig +short @$tiny -p $port pinkpenguin.interblock.io`
  else
    A2=`dig +short @$tiny pinkpenguin.interblock.io`
  fi
  echo -e "Second record =>\t\t\t$A2"
  if [ "$A2" != "$IP" ]; then
    echo "Resolved wrong ip for interblock.io. Got $A2, expected $IP"
    test_exit=1
    return 1
  fi
}

test_teardown(){
  if [ "$ci" = false ]
  then
    echo
    eris services stop -rxf mindy 1>/dev/null
    eris chains stop -f $chain_name 1>/dev/null
    eris chains rm -x --file $chain_name 1>/dev/null
    if [ "$was_running" -eq 0 ]
    then
      eris services stop -rx keys &>/dev/null
    fi
    rm -rf $HOME/.eris/scratch/data/mindy-tests-*
    rm -rf $chain_dir
  else
    eris services stop -f mindy 1>/dev/null
    eris chains stop -f $chain_name 1>/dev/null
  fi
  echo
  if [ "$test_exit" -eq 0 ]
  then
    echo "Tests complete! Tests are Green. :)"
  else
    echo "Tests complete. Tests are Red. :("
  fi
  cd $start
  exit $test_exit
}

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on

echo "Hello! I'm the marmot that tests mindy."
start=`pwd`
cd $repo
echo ""
echo "Building mindy in a docker container."
set -e
tests/build_tool.sh 1>/dev/null
set +e
if [ $? -ne 0 ]
then
  echo "Could not build mindy. Debug via by directly running [`pwd`/tests/build_tool.sh]"
  exit 1
fi
echo "Build complete."
echo ""

# ---------------------------------------------------------------------------
# Setup

test_setup

# ---------------------------------------------------------------------------
# Go!

perform_tests

# ---------------------------------------------------------------------------
# Cleaning up

test_teardown

