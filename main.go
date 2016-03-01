package main

import (
    "log"
    "os"
    "github.com/30x/shipyard/shipyard"
)

func main() {

    sourceInfo, err := shipyard.CreateNewWorkspace()

    if(err != nil){
        shipyard.Log.Fatalf("Could not create a new workspace, returning", err)
        os.Exit(1)
    }




    dockerImage :=  &shipyard.DockerInfo{
        TarFile: "/tmp/test.tar",
        RepoName: "test",
        ImageName: "test",
        Version: "v1.0",
    }

    const remoteUrl = "http://localhost:5000"

    imageCreator, error := shipyard.NewImageCreator(remoteUrl)

    if(error != nil){
        log.Fatal( error)
        os.Exit(1)
    }

    imageCreator.ListImages()

    imageCreator.BuildImage(dockerImage)
    imageCreator.TagImage(dockerImage)
    imageCreator.PushImage(dockerImage)


}
