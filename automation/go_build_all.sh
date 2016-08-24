#!/bin/bash

# CURRPWD=`pwd`
# cd ../chco2/
# go build chco2.go
# cd $CURRPWD

# for (( peer_id=1; $peer_id<"$NUM_PEERS"; peer_id++ ))

# for name in ${*} {

echo -e "go build all the *.go files in current directory:"
for name in $( /bin/ls  | grep \.go$ )
do
  echo -e "$name"
  go build $name
done
echo -e "done."
