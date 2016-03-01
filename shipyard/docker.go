package shipyard

import (
	"github.com/fsouza/go-dockerclient"
	"fmt"
)

//DockerInfo is a struct that holds information for creating a docker container
type DockerInfo struct {
	TarFile   string
	RepoName  string
	ImageName string
	Version   string
}

//ImageCreator is a struct that holds our pointer to the docker client
type ImageCreator struct {
	//the client to docker
	client     *docker.Client
	//the remote repository url
	remoteRepo string
}

//NewImageCreator creates an instance of the ImageCreator from the docker environment variables, and returns the instance
func NewImageCreator(remoteRepo string) (*ImageCreator, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	imageCreator := &ImageCreator{
		client : client,
		remoteRepo: remoteRepo,
	}
	return imageCreator, nil
}

//listImages prints all images in the system, just here for show
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

//UploadImage uploads a docker tar from the specified dockerInfo to the specified repo, image, and version
func (imageCreator *ImageCreator) BuildImage(dockerInfo *DockerInfo) {

}

//Tag the remote image with the given dockerInfo
func (imageCreator *ImageCreator) TagImage(dockerInfo *DockerInfo ) {

}

//PushImage pushes the remotely tagged image to docker
func (imageCreator *ImageCreator) PushImage(dockerInfo *DockerInfo){

}

