package docker

import (
	"github.com/fsouza/go-dockerclient"
	"fmt"
)

//DockerInfo is a struct that holds information for creating a docker container
type DockerInfo struct {
	tmpDir    string
	tarFile   string
	repoName  string
	imageName string
	version   string
}

//ImageCreator is a struct that holds our pointer to the docker client
type ImageCreator struct {
	client *docker.Client
}

//NewImageCreator creates an instance of the ImageCreator from the docker environment variables, and returns the instance
func NewImageCreator() (*ImageCreator, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	imageCreator := &ImageCreator{client : client}
	return imageCreator, nil
}

//listImages prints all images in the system
func (imageCreator *ImageCreator) ListImages() {
	// use client
	imgs, _ := imageCreator.client.ListImages(docker.ListImagesOptions{All: false})
	for _, img := range imgs {
		fmt.Println("ID: ", img.ID)
		fmt.Println("RepoTags: ", img.RepoTags)
		fmt.Println("Created: ", img.Created)
		fmt.Println("Size: ", img.Size)
		fmt.Println("VirtualSize: ", img.VirtualSize)
		fmt.Println("ParentId: ", img.ParentID)
	}

}
