#! /bin/bash

# import the priv_validator and set vars
ADDR=`mintkey eris /priv_validator.json`
echo addr $ADDR
export MINTX_NODE_ADDR="http://mint:46657/"
export MINTX_SIGN_ADDR="http://localhost:4767"
export MINTX_CHAINID=mindy_test

# start key daemon
echo "run the keys daemon ..."
eris-keys server &

# let the daemon start
sleep 2

export MINTX_PUBKEY=`eris-keys pub --addr $ADDR`
echo pub $MINTX_PUBKEY

# get mint ip for dns registration
IP=`cat /etc/hosts | grep mint | awk 'NR==1{print \$1}'` 
echo ip "$IP"

# register some dns entries
DATA="[{\"fqdn\":\"interblock.io\", \"address\":\"$IP\", \"type\":\"NS\"}, {\"fqdn\":\"interblock.io\", \"address\":\"$IP\", \"type\":\"A\"}]"
echo $DATA
mintx name --name=interblock.io --data="$DATA" --amt=1000000 --fee=0 --sign --broadcast

DATA="{\"fqdn\":\"pinkpenguin.interblock.io\", \"address\":\"$IP\", \"type\":\"A\"}"
echo $DATA
mintx name --name=pinkpgenuin.interblock.io --data="$DATA" --amt=100000 --fee=0  --sign --broadcast --wait

# let tinydns get updated
sleep 5

A1=`dig +short @tiny interblock.io`
echo $A1
if [ "$A1" != "$IP" ]; then
	echo "Resolved wrong ip for interblock.io. Got $A1, expected $IP"
	exit 1
fi
A2=`dig +short @tiny pinkpenguin.interblock.io`
echo $A2
if [ "$A2" != "$IP" ]; then
	echo "Resolved wrong ip for interblock.io. Got $A2, expected $IP"
	exit 1
fi

echo PASS


