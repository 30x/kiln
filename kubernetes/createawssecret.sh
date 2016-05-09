#!/bin/bash

###
# Writes a secret named "aws-access" and there are 2 fields in it.
# key={your access key}
# secret={your secret}
###

function test {
    local status=$?
    if [ $status -ne 0 ]; then
        echo "error with $1" >&2
        exit 3
    fi
}

function show_help {
    echo "Usage is $0 -k {AWS_KEY} -s {AWS_SECRET} -n {K8s namespace}"

}


ACCESS_KEY=""
SECRET_KEY=""
NAMESPACE=""


#get opts
while getopts "s:k:n:" opt; do
  case $opt in
    k)
        echo "ACCESS_KEY $OPTARG"
        ACCESS_KEY=$OPTARG
        ;;
    s)
        echo "SECRET_KEY $OPTARG"
        SECRET_KEY=$OPTARG
        ;;
    n)
        echo "NAMSPACE $OPTARG"
        NAMESPACE=$OPTARG
        ;;
    \?)
        show_help
        exit 1
        ;;
  esac
done


#Validate input
if [ -z "${ACCESS_KEY}" ]; then
    show_help
    exit 1
fi

if [ -z "${SECRET_KEY}" ]; then
    show_help
    exit 1
fi

if [ -z "${NAMESPACE}" ]; then
    show_help
    exit 1
fi



#Now call kubectl and set the secrets
kubectl delete secret aws-access --namespace=${NAMESPACE}
kubectl create secret generic aws-access --from-literal=key=${ACCESS_KEY} --from-literal=secret=${SECRET_KEY} --namespace=${NAMESPACE}

echo "Created kuberentes aws secret aws-access in namespace ${NAMESPACE}"
