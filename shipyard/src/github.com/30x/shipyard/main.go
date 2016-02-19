package main

import (
    "github.com/30x/shipyard/docker"
    "log"
    "os"
)

func main() {

    imageCreator, error := docker.NewImageCreator()

    if(error != nil){
        log.Fatal( error)
        os.Exit(1)
    }

    imageCreator.ListImages()


}
