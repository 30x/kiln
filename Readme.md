# Shipyard

A system for creating and managing docker based node runtimes for node apps

## Overview

The purpose of this application is to allow developers to post a zip file of node.js source, and then have a docker images created, updloaded to an internal repository, and have a pod template spec built.


## Examples

The easiest form of client demonstration is the Postman example within the example directory.  Import this collection into postman locally with docker running.  Start the server with the following command.

```bash
SHUTDOWN_TIMEOUT=1 PORT=5280 DOCKER_PROVIDER=docker DOCKER_REGISTRY_URL=localhost:5000 POD_PROVIDER=local LOCAL_DIR=/tmp/storagedir go run main.go
```

You can then use the collection to experiment with the api.  The zip file within the example directory is a simple echo node application. This will echo the output of the request.