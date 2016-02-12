#K8s Parity Between Dev and Deploy

>Notes for getting images running on both local dev environment and K8s cluster spun up by Mamu.

##Goal

Starting with echo-test.zip file. We want to deploy the zipped Node.js app to a local K8s cluster as well as a remote AWS K8s cluster. 

This document should serve as a step by step guide to achieve this on both environments as well as a reference that points out the differences and sticking points with both use cases.

##Dev Environment

Extract `echo-test.zip` to a new directory.

```sh
mkdir echo-test && tar -xf echo-test.zip -C echo-test
```
Create `Dockerfile` in `echo-test` directory.

```Dockerfile
FROM mhart/alpine-node:4

WORKDIR /src
ADD . .

RUN npm install

EXPOSE 3000
CMD ["npm", "start"]
```

Build Docker image from directory.

```sh
docker build -t testuser/echo-test echo-test
```

Create a replication controller in the K8s cluster from the image. 

```sh
kubectl run echo --image=testuser/echo-test
```

Expose the replication controller as a service.

```sh
kubectl expose rc echo --port=80 --target-port=3000 --type=NodePort
```

Get the NodePort of the service.

```sh
kubectl describe svc echo | grep NodePort
```

The output should look like this:

```sh
Type:			NodePort
NodePort:		<unnamed>	31554/TCP
```

Now verify that you can get a response from the server. Replace <NodePort> with the port provided above and <vm-name> with the name of your active docker-machine vm.

```sh
curl $(docker-machine ip <vm-name>):<NodePort>
```

The response should look like this:

```sh
{"echo":"hello!"}%
```

##AWS Environment

For convenience the `echo-test` has been uploaded to dockerhub. It is available at `jbowen93/echo-test`.


Create a replication controller in the K8s cluster from the image. 

```sh
kubectl run echo --image=jbowen93/echo-test
```

Expose the replication controller as a service.

```sh
kubectl expose rc echo --port=80 --target-port=3000 --type=NodePort
```

