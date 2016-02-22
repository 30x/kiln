#About

This project is a simple POC that will take a node application from a zip file, extract it, then create a docker image
and push it to a remote repository


#Generating assets for use within the code

Assets such as the DockerFile are embedded using go-bindata.  https://github.com/jteeuwen/go-bindata

When modifying the docker file, ensure you re-generate the source.  Install and edit with the following from the project root

```
go get -u github.com/jteeuwen/go-bindata/...

go-bindata resources
```




