#!/usr/bin/env bash


if [ -z ${DOCKER_MACHINE_NAME+x} ]; then
  echo "DOCKER_MACHINE_NAME must be set, otherwise ssh tunneling for image push will not work as expected";
  exit 1;
fi

## Remove services in kubectl, missing errors can be ignored
kubectl delete namespaces app1
kubectl delete namespaces app2


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

kubectl create -f app1/k8s/namespace1.yaml
kubectl create -f app1/k8s/app1.yaml --namespace=app1
kubectl create -f app1/k8s/service1.yaml --namespace=app1

kubectl create -f app2/k8s/namespace2.yaml
kubectl create -f app2/k8s/app2.yaml --namespace=app2
kubectl create -f app2/k8s/service2.yaml --namespace=app2


#Wait for pod names to come online
while :
do
   APP1_POD_NAME=$(kubectl get po --namespace=app1 | grep app1 | awk '{print $1}')
   APP2_POD_NAME=$(kubectl get po --namespace=app2 | grep app2 | awk '{print $1}')


   if [[ "$APP1_POD_NAME" != "" && "$APP2_POD_NAME" != "" ]] ; then
	  break       	   #Abandon the loop. we're ready
   fi

   echo "Waiting for pods names to come online"
   sleep 1
done



echo "APP1_POD_NAME is $APP1_POD_NAME"
echo "APP2_POD_NAME is $APP2_POD_NAME"


#Now get the service ip for namespace 1

APP1_SERVICE_IP=$(kubectl get svc --namespace=app1 | grep app1 | awk '{print $2}')

echo "SERVICE_IP to test is $APP1_SERVICE_IP"



#Now we wait until the pods are ready
while :
do
   APP1_STATUS=$(kubectl get po --namespace=app1 | grep app1 | awk '{print $3}')
   APP2_STATUS=$(kubectl get po --namespace=app2 | grep app2 | awk '{print $3}')

   if [[ "$APP1_STATUS" = "Running" && "$APP2_STATUS" = "Running" ]] ; then
	  break       	   #Abandon the loop. we're ready
   fi

   echo "Waiting for pods to come online"
   sleep 1
done

echo "Waiting for nginx to initialize"

sleep 5

echo "About to execute command kubectl exec --namespace=app2 $APP2_POD_NAME -- curl --fail $APP1_SERVICE_IP"

kubectl exec --namespace=app2 $APP2_POD_NAME -- curl --fail $APP1_SERVICE_IP

STATUS=$?

if [[ STATUS == 0 ]] ; then
  echo "FAILURE: Curl was successful, the pod in app 2 should not be allowed to call the service in the namespace app1"
else
  echo "PASS:  The pod in app 2 cannot see the serice in namespace app 1"
fi
