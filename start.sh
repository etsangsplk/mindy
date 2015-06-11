#! /bin/bash

## Start the tinydns server
cd /etc/service
svscan &

# download dns entries from blockchain
mindy catchup

# run daemon that listens for nametx events
# and updates tinydns
mindy run

