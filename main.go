package shipyard

import (
    "log"
    "os"
)

func main() {

    dockerImage :=  &DockerInfo{
        TarFile: "/tmp/test.tar",
        RepoName: "test",
        ImageName: "test",
        Version: "v1.0",
    }

    const remoteUrl = "http://localhost:5000"

    imageCreator, error := NewImageCreator(remoteUrl)

    if(error != nil){
        log.Fatal( error)
        os.Exit(1)
    }

    imageCreator.ListImages()

    imageCreator.BuildImage(dockerImage)
    imageCreator.TagImage(dockerImage)
    imageCreator.PushImage(dockerImage)


}
