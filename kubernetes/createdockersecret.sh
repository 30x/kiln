#!/bin/bash


####
# Moves the existing docker file out of the way.  Creates the new one, then
# adds the secret.  After complete moves the original back
###

SECRET_NAME="docker-config"
ORIG_CONF_FILE="${HOME}/.docker/config.json"
TEMP_CONF_FILE="${ORIG_CONF_FILE}.orig"

function test {
    local status=$?
    if [ $status -ne 0 ]; then
        echo "error with $1" >&2
        exit 3
    fi
}

#Copy the temp file back to the original file
function finish {
  if [ -f $TEMP_CONF_FILE ];
  then
    mv ${TEMP_CONF_FILE} ${ORIG_CONF_FILE}
  fi
}



function show_help {
    echo "Usage is $0 -k {AWS_KEY} -s {AWS_SECRET} -r {AWS REGION} -n {K8s namespace}"

}


ACCESS_KEY=""
SECRET_KEY=""
NAMESPACE=""
REGION=""


#get opts
while getopts "s:k:n:r:" opt; do
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
    r)
        echo "REGION $OPTARG"
        REGION=$OPTARG
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

if [ -z "${REGION}" ]; then
    show_help
    exit 1
fi

#Set the trap to copy back if we fail somewhere
trap finish EXIT

export AWS_ACCESS_KEY=${ACCESS_KEY}
export AWS_SECRET_KEY=${SECRET_KEY}
#Now call kubectl and set the secrets

#Copy the file into the temp
mv ${ORIG_CONF_FILE} ${TEMP_CONF_FILE}

LOGIN_COMMAND="$(aws ecr get-login --region ${REGION})"

#Get the hostname to set into the secret
HOSTNAME="$(echo ${LOGIN_COMMAND} | awk '{print $9}')"
HOSTNAME="${HOSTNAME//https:\/\//}"


#Run the login command to update the docker config
echo "Setting docker host to ${HOSTNAME}"
echo "Executing login command"
echo  "${LOGIN_COMMAND}"

eval ${LOGIN_COMMAND}

test "Could not create login file"

#Delete, ignore errors if it doesn't exist
kubectl  --namespace=${NAMESPACE} delete secret ${SECRET_NAME}
kubectl  --namespace=${NAMESPACE} create secret generic ${SECRET_NAME} --from-file=${HOME}/.docker/config.json --from-literal=host=${HOSTNAME}

test "Could not create secret in kubernetes"
