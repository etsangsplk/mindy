# mindy

Mindy extends djb's tinydns by updating the zone file with records fetched from a blockchain 

Run a test mindy on a one man blockchain across four containers with `./test/build.sh`

Updating dns records with blockchain transactions:
---------

First thing we need the `eris-keys` daemon to generate private keys and handle signing.

```
go get github.com/eris-ltd/eris-keys
eris-keys &
eris-keys gen
```

This will install the software, start the daemon, and generate a new key (keys are stored in ~/.eris/keys/).
The final output is your tendermint address. We need to also know the public key:

```
eris-keys pub <address>
```

where `<address>` is the address you just generated with `eris-keys gen`

Now, to simplify our lives, lets set a few environment variables.
Open your `~/.bashrc` and paste in the following

```
export MINTX_SIGN_ADDR=http://localhost:4767
export MINTX_NODE_ADDR=http://pinkpenguin.chaintest.net:46657/
export MINTX_PUBKEY=<PUBKEY>
export MINTX_CHAINID=tendermint_testnet_5e
```

where `<PUBKEY>` is the output of the `eris-keys pub` command.

Run `source ~/.bashrc` to make the new variables take effect.

Now we can send transactions!

Well, technically, someone has to send you a transaction first so you have some funds. 
Contact someone with funds on the chain and give them your address so they can send you funds. 
Sending funds is easy as:

```
mintx send --to <addr> --amt <amount> --sign --broadcast
```

You can monitor the accounts at `http://pinkpenguin.chaintest.net:46657/list_accounts`
Alternatively, run

```
mintinfo accounts
```

Now, lets say we want to create an A record for the subdomain `magma.interblock.io` at `1.2.3.4`. 
We would send the following tx:

```
mintx name --name magma.interblock.io --data '{"fqdn":"magma.interblock.io", "address":"1.2.3.4", "type":"A"}' --amt 100000 --fee 0 --sign --broadcast
```

The data you pass must be specially formatted (as shown) to be recognized as a dns entry.
It's just json with an `fqdn`, `address`, and `type` argument.
The `--amt` flag determines how long your entry will remain on the chain (or at least how long until it can be overwritten,
but it is not automatically deleted). Wait about 30 seconds, and you should be able to resolve the new domain with dig:

```
dig @ns1.interblock.io magma.interblock.io
```

Of course this uses our name server directly, and it might take a little longer before the new record propogates 
to other dns caches.

To run your own nameservers linked to the blockchain, send a NameTx with the `"type"` argument set to `"NS"`, eg:

```
mintx name --name newdomain.io --data '{"fqdn":"newdomain.io", "address":"4.3.2.1", "type":"NS"}' --amt 100000 --fee 0 --sign --broadcast
```

Of course this will only work with the real DNS system if you own the domain `newdomain.io`, and you tell your registrar to point at your own nameservers (presumably `4.3.2.1`). Then just run our docker container and you'll be serving DNS straight from the blockchain!

# Run the containers yourself

A full mindy node uses 4 containers: one for each of tendermint, tendermint's data, tinydns, and mindy. They can be used all together 
or in varying combinations. See the `tendermint` and `tinydns` sub-directories for details on running those containers.

Once tinydns and tendermint are running, 

```
docker build -t mindy .
docker run -d --name mindy --volumes-from tinydns --link mint:mint mindy 
```