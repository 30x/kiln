package shipyard

import (
	"bytes"
	"github.com/fsouza/go-dockerclient"
	"os"
	// "io"
)

//DockerInfo is a struct that holds information for creating a docker container
type DockerInfo struct {
	RepoName  string
	ImageName string
	Revision  string
}

//DockerBuild a type for building a docker image docker
type DockerBuild struct {
	TarFile string
	*DockerInfo
}

func (dockerInfo *DockerInfo) getImageName() string {
	return dockerInfo.RepoName + "/" + dockerInfo.ImageName
}

//getTagName Get the anme of the tag
func (dockerInfo *DockerInfo) getTagName() string {
	return dockerInfo.getImageName() + ":" + dockerInfo.Revision
}

//getRemoteTagName Get the remote tag name of the docker repo
func (dockerInfo *DockerInfo) getRemoteTagName(remoteRepo string) string {
	return remoteRepo + "/" + dockerInfo.getImageName()
}

//ImageCreator is a struct that holds our pointer to the docker client
type ImageCreator struct {
	//the client to docker
	client *docker.Client
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

	//TODO, stop selecting all
	_, err = client.ListImages(docker.ListImagesOptions{All: false})

	if err != nil {
		return nil, err
	}

	return imageCreator, nil
}

//ListImages prints all images in the system, just here for show
func (imageCreator *ImageCreator) ListImages() ([]docker.APIImages, error) {
	// use client

	opts := docker.ListImagesOptions{All: false}

	return imageCreator.client.ListImages(opts)

}

//BuildImage creates a docker tar from the specified dockerInfo to the specified repo, image, and version
func (imageCreator *ImageCreator) BuildImage(dockerInfo *DockerBuild) error {

	name := dockerInfo.getTagName()

	Log.Printf("Uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	inputReader, err := os.Open(dockerInfo.TarFile)

	if err != nil {
		Log.Fatal("Unable to open tar file " + dockerInfo.TarFile + "for input")
		return err
	}

	//make an output buffer with 1m
	outputBuffer := &bytes.Buffer{}

	buildImageOptions := docker.BuildImageOptions{
		Name:         name,
		InputStream:  inputReader,
		OutputStream: outputBuffer,
	}

	//hacky, need to pipe this to a log
	//outputBuffer.WriteTo(std_out)

	Log.Printf("Started build image with options %s", buildImageOptions)

	if err := imageCreator.client.BuildImage(buildImageOptions); err != nil {
		Log.Fatal(err)
		return err
	}

	Log.Printf("Completed build image with options %s", buildImageOptions)

	// TODO fix this
	// outputBuffer.WriteTo(os.Stdout)

	//print the output stream

	return nil

}

//PushImage pushes the remotely tagged image to docker
func (imageCreator *ImageCreator) PushImage(dockerInfo *DockerInfo) error {

	localTag := dockerInfo.getTagName()
	remoteRepo := imageCreator.remoteRepo
	remoteTag := dockerInfo.getRemoteTagName(remoteRepo)
	revision := dockerInfo.Revision

	tagOptions := docker.TagImageOptions{
		Repo: remoteTag,
		Tag:  revision,
	}

	err := imageCreator.client.TagImage(localTag, tagOptions)

	if err != nil {
		return err
	}

	//now push the image
	outputBuffer := &bytes.Buffer{}

	pushOts := docker.PushImageOptions{
		Name:         remoteTag,
		Registry:     remoteRepo,
		Tag:          revision,
		OutputStream: outputBuffer,
	}

	authConfig := docker.AuthConfiguration{}

	err = imageCreator.client.PushImage(pushOts, authConfig)

	if err != nil {
		return err
	}

	return nil
}

//PullImage pull the specified image to our the docker runtime
func (imageCreator *ImageCreator) PullImage(dockerInfo *DockerInfo) error {

	remoteRepo := imageCreator.remoteRepo
	remoteTag := dockerInfo.getRemoteTagName(remoteRepo)
	revision := dockerInfo.Revision

	outputBuffer := &bytes.Buffer{}

	pullOpts := docker.PullImageOptions{
		Repository:   remoteTag,
		Registry:     remoteRepo,
		Tag:          revision,
		OutputStream: outputBuffer,
	}

	authConfig := docker.AuthConfiguration{}

	return imageCreator.client.PullImage(pullOpts, authConfig)
}
