# Purpose

The purpose of this system is to create 2 simple nginx containers.  These containers
can then be tested for visibility between containers once running in a kuberentes
environment.

# Building images

```sh
./buildandpush.sh
```

This will remove any existing kubernetes services.  Build and push the latest docker image, then start
both services.  You can then hit the following urls.

http://<your docker ip>:30010 for app1
http://<your docker ip>:30011 for app2
