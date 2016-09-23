#!/bin/bash

# This is an excerpt from the GOPATH/bin/local_fabric.sh file.

# Docker is not perfect; we need to unpause any paused containers, before we can kill them.
docker ps -aq -f status=paused | xargs docker unpause  1>/dev/null 2>&1

# kill all running docker containers and LOGFILES...This may need to be revisited.

docker kill $(docker ps -q) 1>/dev/null 2>&1
docker ps -aq -f status=exited | xargs docker rm -f 1>/dev/null 2>&1
# cd ../automation && rm -f LOG*
docker rm -f $(docker ps -aq)

docker rmi -f $(docker images | grep "<none>" | awk '{print $3}')
docker rmi -f $(docker images | grep "dev-" | awk '{print $3}')

