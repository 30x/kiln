package main

import (
    "fmt"
       "github.com/docker/engine-api/client"
       "github.com/docker/engine-api/types"
    //"github.com/fsouza/go-dockerclient"
)

func main() {
    defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
        cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
        if err != nil {
            panic(err)
        }

        options := types.ContainerListOptions{All: true}
        containers, err := cli.ContainerList(options)
        if err != nil {
            panic(err)
        }

        for _, c := range containers {
            fmt.Println(c.ID)
        }
    //client, _ := docker.NewClientFromEnv()
    //// use client
	//imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})
    //for _, img := range imgs {
    //    fmt.Println("ID: ", img.ID)
    //    fmt.Println("RepoTags: ", img.RepoTags)
    //    fmt.Println("Created: ", img.Created)
    //    fmt.Println("Size: ", img.Size)
    //    fmt.Println("VirtualSize: ", img.VirtualSize)
    //    fmt.Println("ParentId: ", img.ParentID)
    //}
}
