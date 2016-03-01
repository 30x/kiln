package shipyard

import (
	"github.com/fsouza/go-dockerclient"
	"bytes"
	"os"
)

//DockerInfo is a struct that holds information for creating a docker container
type DockerInfo struct {
	TarFile   string
	RepoName  string
	ImageName string
	Revision  string
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
		client:     client,
		remoteRepo: remoteRepo,
	}
	return imageCreator, nil
}

//listImages prints all images in the system, just here for show
func (imageCreator *ImageCreator) ListImages() ([]docker.APIImages, error) {
	// use client
	return imageCreator.client.ListImages(docker.ListImagesOptions{All: false})

}

//UploadImage uploads a docker tar from the specified dockerInfo to the specified repo, image, and version
func (imageCreator *ImageCreator) BuildImage(dockerInfo *DockerInfo) error {

	name := dockerInfo.RepoName + "/" + dockerInfo.ImageName + ":" + dockerInfo.Revision

	Log.Printf("Uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	inputReader, err := os.Open(dockerInfo.TarFile)

	if (err != nil) {
		Log.Fatal("Unable to open tar file " + dockerInfo.TarFile + "for input")
		return err
	}

	Log.Printf("Docker tar file is %s", inputReader)

	//make an output buffer with 1k
	outputBuffer := bytes.NewBuffer(make([]byte, 1024))

	buildImageOptions := docker.BuildImageOptions{
		Name: name,
		InputStream:  inputReader,
		OutputStream: outputBuffer,
	}

	//hacky, need to pipe this to a log
	outputBuffer.WriteTo(std_out)

	if err := imageCreator.client.BuildImage(buildImageOptions); err != nil {
		Log.Fatal(err)
		return err
	}



	//print the output stream

	return nil

}

//Tag the remote image with the given dockerInfo
func (imageCreator *ImageCreator) TagImage(dockerInfo *DockerInfo) {

}

//PushImage pushes the remotely tagged image to docker
func (imageCreator *ImageCreator) PushImage(dockerInfo *DockerInfo) {

}
