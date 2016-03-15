package shipyard

import (
	"github.com/fsouza/go-dockerclient"
	"io"
)

//EcsImageCreator is a struct that holds our pointer to the docker client
type EcsImageCreator struct {
	//the client to docker

	//the remote repository url
	remoteRepo string
}

//NewEcsImageCreator creates an instance of the EcsImageCreator from the docker environment variables, and returns the instance
func NewEcsImageCreator() (ImageCreator, error) {
	//

	return &EcsImageCreator{}, nil
}

//ListImages prints all images in the system, just here for show
func (imageCreator EcsImageCreator) ListImages() ([]docker.APIImages, error) {
	return nil, nil
}

//SearchImages return all images with matching labels.  The label name is the key, the values are the value strings
func (imageCreator EcsImageCreator) SearchImages(search *ImageSearch) ([]docker.APIImages, error) {

	//initialize the filter map

	filters := []string{}

	//append filters as required based on the input
	if search.Repository != "" {
		newFilter := TAG_REPO + "=" + search.Repository
		filters = append(filters, newFilter)
	}

	if search.Application != "" {
		newFilter := TAG_APPLICATION + "=" + search.Application
		filters = append(filters, newFilter)
	}

	if search.Revision != "" {
		newFilter := TAG_REVISION + "=" + search.Revision
		filters = append(filters, newFilter)
	}

	return nil, nil
}

//BuildImage creates a docker tar from the specified dockerInfo to the specified repo, image, and version.  Returns the reader stream or an error
func (imageCreator EcsImageCreator) BuildImage(dockerInfo *DockerBuild, logs io.Writer) error {

	return nil

}

//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
func (imageCreator EcsImageCreator) PushImage(dockerInfo *DockerInfo, logs io.Writer) error {

	return nil
}

//PullImage pull the specified image to our the docker runtime
func (imageCreator EcsImageCreator) PullImage(dockerInfo *DockerInfo, logs io.Writer) error {

	return nil
}
