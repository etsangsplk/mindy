#! /bin/bash

# download dns entries from blockchain
mindy catchup

# run daemon that listens for nametx events
# and updates tinydns
if [ $UPDATE_EVERY = "" ]; then
	UPDATE_EVERY=30
fi
mindy run -u $UPDATE_EVERY

