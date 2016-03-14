package shipyard

import (
	"github.com/fsouza/go-dockerclient"
	"os"
	// "io"
	"io"
)

//getImageName generate an image of the format {RepoName}/{ImageName}
func (dockerInfo *DockerInfo) getImageName() string {
	return dockerInfo.RepoName + "/" + dockerInfo.ImageName
}

//getTagName Get the anme of the tag of the format  {RepoName}/{ImageName}:{Revision}
func (dockerInfo *DockerInfo) getTagName() string {
	return dockerInfo.getImageName() + ":" + dockerInfo.Revision
}

//getRemoteTagName Get the remote tag name of the docker repo of the format {RemoteRepo}/{RepoName}/{ImageName}:{Revision}
func (dockerInfo *DockerInfo) getRemoteTagName(remoteRepo string) string {
	return remoteRepo + "/" + dockerInfo.getImageName()
}

//LocalImageCreator is a struct that holds our pointer to the docker client
type LocalImageCreator struct {
	//the client to docker
	client *docker.Client
	//the remote repository url
	remoteRepo string
}

//NewLocalImageCreator creates an instance of the LocalImageCreator from the docker environment variables, and returns the instance
func NewLocalImageCreator() (ImageCreator, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	imageCreator := LocalImageCreator{
		client:     client,
		remoteRepo: "localhost:5000",
	}

	//TODO, stop selecting all
	_, err = client.ListImages(docker.ListImagesOptions{All: false})

	if err != nil {
		return nil, err
	}

	return imageCreator, nil
}

//ListImages prints all images in the system, just here for show
func (imageCreator LocalImageCreator) ListImages() ([]docker.APIImages, error) {
	opts := docker.ListImagesOptions{All: false}

	return imageCreator.client.ListImages(opts)
}

//SearchImages return all images with matching labels.  The label name is the key, the values are the value strings
func (imageCreator LocalImageCreator) SearchImages(search *ImageSearch) ([]docker.APIImages, error) {

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

	filter := map[string][]string{
		"label": filters,
	}

	opts := docker.ListImagesOptions{All: false, Filters: filter}

	return imageCreator.client.ListImages(opts)
}

//BuildImage creates a docker tar from the specified dockerInfo to the specified repo, image, and version.  Returns the reader stream or an error
func (imageCreator LocalImageCreator) BuildImage(dockerInfo *DockerBuild, logs io.Writer) error {

	name := dockerInfo.getTagName()

	Log.Printf("Started uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	inputReader, err := os.Open(dockerInfo.TarFile)

	if err != nil {
		Log.Fatal("Unable to open tar file "+dockerInfo.TarFile+"for input", err)
		return err
	}

	//make an output buffer with 1m
	buildImageOptions := docker.BuildImageOptions{
		Name:         name,
		InputStream:  inputReader,
		OutputStream: logs,
	}

	if err := imageCreator.client.BuildImage(buildImageOptions); err != nil {
		Log.Fatal(err)
		return err
	}

	Log.Printf("Completed uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	return nil

}

//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
func (imageCreator LocalImageCreator) PushImage(dockerInfo *DockerInfo, logs io.Writer) error {

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

	pushOts := docker.PushImageOptions{
		Name:         remoteTag,
		Registry:     remoteRepo,
		Tag:          revision,
		OutputStream: logs,
	}

	authConfig := docker.AuthConfiguration{}

	err = imageCreator.client.PushImage(pushOts, authConfig)

	if err != nil {
		return err
	}

	return nil
}

//PullImage pull the specified image to our the docker runtime
func (imageCreator LocalImageCreator) PullImage(dockerInfo *DockerInfo, logs io.Writer) error {

	remoteRepo := imageCreator.remoteRepo
	remoteTag := dockerInfo.getRemoteTagName(remoteRepo)
	revision := dockerInfo.Revision

	pullOpts := docker.PullImageOptions{
		Repository:   remoteTag,
		Registry:     remoteRepo,
		Tag:          revision,
		OutputStream: logs,
	}

	authConfig := docker.AuthConfiguration{}

	return imageCreator.client.PullImage(pullOpts, authConfig)
}
