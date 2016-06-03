#!/bin/sh




if [ ! -r /root/.docker/config.json ]
then
    echo "No file /root/.docker/config.json exists.  Checking for k8s secret"
    
    if [ ! -r /root/k8s-secret/.dockerconfigjson ]
    then
         echo "No file /root/k8s-secret/.dockerconfigjson exists.  Cannot start service"
         exit 1
    fi 
    
    # if we get here, we just need to copy the file over

    echo "Creating directory /root/.docker"
    mkdir -p /root/.docker
    echo "Copying /root/k8s-secret/.dockerconfigjson to /root/.docker/config.json "
    cp /root/k8s-secret/.dockerconfigjson /root/.docker/config.json 
fi

echo "File in place, starting shipyard"

./shipyard