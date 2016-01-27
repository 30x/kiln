#!/usr/bin/env bash


if [ -z ${DOCKER_MACHINE_NAME+x} ]; then
  echo "DOCKER_MACHINE_NAME must be set";
  exit 1;
fi

##
# This script assumes you've created a tunnel on port 5000 to a docker
# registry container
##

## Append the insecure-registry
pushd app1

docker build -t tnine/app1 .
docker tag -f tnine/app1 localhost:5000/tnine/app1
docker push  localhost:5000/tnine/app1

popd

pushd app2

docker tag -f tnine/app2 localhost:5000/tnine/app2
docker push  localhost:5000/tnine/app2
docker push  localhost:5000/tnine/app2

popd
