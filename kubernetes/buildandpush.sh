#!/usr/bin/env bash


if [ -z ${DOCKER_MACHINE_NAME+x} ]; then
  echo "DOCKER_MACHINE_NAME must be set, otherwise ssh tunneling for image push will not work as expected";
  exit 1;
fi

## Remove services in kubectl, missing errors can be ignored
kubectl delete rc app1
kubectl delete svc app1

kubectl delete rc app2
kubectl delete svc app2


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

docker build -t tnine/app2 .
docker tag -f tnine/app2 localhost:5000/tnine/app2
docker push  localhost:5000/tnine/app2

popd

######
# Launch the replication controller and serices in kubernetes
####

kubectl create -f app1/k8s/app1.yaml
kubectl create -f app1/k8s/service1.yaml

kubectl create -f app2/k8s/app2.yaml
kubectl create -f app2/k8s/service2.yaml
