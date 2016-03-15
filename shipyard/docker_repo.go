package shipyard

/**

This implementation supports the local docker API, as well as the docker provided remote registry
**/
import (
	"github.com/fsouza/go-dockerclient"
	"os"
	// "io"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/engine-api/types"
	"io"
	"strings"
)

//LocalImageCreator is a struct that holds our pointer to the docker client
type LocalImageCreator struct {
	//the client to docker
	client *docker.Client
	//the remote repository url
	remoteRepo string
}

//NewLocalImageCreator creates an instance of the LocalImageCreator from the docker environment variables, and returns the instance
func NewLocalImageCreator(repo string) (ImageCreator, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	imageCreator := LocalImageCreator{
		client:     client,
		remoteRepo: repo,
	}

	//TODO, stop selecting all
	_, err = client.ListImages(docker.ListImagesOptions{All: false})

	if err != nil {
		return nil, err
	}

	return imageCreator, nil
}

//SearchLocalImages return all images with matching labels.  The label name is the key, the values are the value strings
func (imageCreator LocalImageCreator) SearchLocalImages(search *DockerInfo) ([]docker.APIImages, error) {

	filter := createFilter(search)

	opts := docker.ListImagesOptions{All: false, Filters: *filter}

	return imageCreator.client.ListImages(opts)
}

//SearchRemoteImages search remote images
func (imageCreator LocalImageCreator) SearchRemoteImages(search *DockerInfo) ([]docker.APIImages, error) {
	//TODO connect to remote repo here
	filter := createFilter(search)

	opts := docker.ListImagesOptions{All: false, Filters: *filter}

	return imageCreator.client.ListImages(opts)

}

//createFilter generate filter from search
func createFilter(search *DockerInfo) *map[string][]string {
	filters := []string{}

	//append filters as required based on the input
	if search.RepoName != "" {
		newFilter := TAG_REPO + "=" + search.RepoName
		filters = append(filters, newFilter)
	}

	if search.ImageName != "" {
		newFilter := TAG_APPLICATION + "=" + search.ImageName
		filters = append(filters, newFilter)
	}

	if search.Revision != "" {
		newFilter := TAG_REVISION + "=" + search.Revision
		filters = append(filters, newFilter)
	}

	filter := map[string][]string{
		"label": filters,
	}

	return &filter
}

//BuildImage creates a docker tar from the specified dockerInfo to the specified repo, image, and version.  Returns the reader stream or an error
func (imageCreator LocalImageCreator) BuildImage(dockerInfo *DockerBuild, logs io.Writer) error {

	name := dockerInfo.GetTagName()

	LogInfo.Printf("Started uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	inputReader, err := os.Open(dockerInfo.TarFile)

	if err != nil {
		LogError.Printf("Unable to open tar file "+dockerInfo.TarFile+"for input", err)
		return err
	}

	//make an output buffer with 1m
	buildImageOptions := docker.BuildImageOptions{
		Name:         name,
		InputStream:  inputReader,
		OutputStream: logs,
	}

	if err := imageCreator.client.BuildImage(buildImageOptions); err != nil {
		LogInfo.Fatal(err)
		return err
	}

	LogInfo.Printf("Completed uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	return nil

}

//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
func (imageCreator LocalImageCreator) PushImage(dockerInfo *DockerInfo, logs io.Writer) error {

	localTag := dockerInfo.GetTagName()
	remoteRepo := imageCreator.remoteRepo
	remoteTag := dockerInfo.GetRemoteTagName(remoteRepo)
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

	//this call on every invocation is deliberate.  Our keys should rotate and we want to continue to function when
	//that happens
	authConfig := generateAuthConfiguration(imageCreator.remoteRepo)

	err = imageCreator.client.PushImage(pushOts, *authConfig)

	if err != nil {
		return err
	}

	return nil
}

//PullImage pull the specified image to our the docker runtime
func (imageCreator LocalImageCreator) PullImage(dockerInfo *DockerInfo, logs io.Writer) error {

	remoteRepo := imageCreator.remoteRepo
	remoteTag := dockerInfo.GetRemoteTagName(remoteRepo)
	revision := dockerInfo.Revision

	pullOpts := docker.PullImageOptions{
		Repository:   remoteTag,
		Registry:     remoteRepo,
		Tag:          revision,
		OutputStream: logs,
	}

	//this call on every invocation is deliberate.  Our keys should rotate and we want to continue to function when
	//that happens
	authConfig := generateAuthConfiguration(imageCreator.remoteRepo)

	return imageCreator.client.PullImage(pullOpts, *authConfig)
}

//generateAuthConfiguration Create an auth configuration from the environment variables
func generateAuthConfiguration(remoteRepo string) *docker.AuthConfiguration {

	authConfig, exists := getAuthConfig(remoteRepo)

	if !exists {
		LogWarn.Printf("Could not find repo %s in auth configuration.  Returning empty auth", authConfig)
		return &docker.AuthConfiguration{}
	}

	clientAuthConfig := &docker.AuthConfiguration{
		Email:         authConfig.Email,
		Password:      authConfig.Password,
		ServerAddress: authConfig.ServerAddress,
		Username:      authConfig.Username,
	}

	return clientAuthConfig
}

//getAuthConfig return the auth config if it exists.  Nil and false otherwise
func getAuthConfig(remoteRepo string) (*types.AuthConfig, bool) {

	configFile, e := cliconfig.Load(cliconfig.ConfigDir())
	if e != nil {
		LogWarn.Printf("Error loading config file:%v\n", e)

		//no auth, return an empty auth
		return nil, false
	}

	for hostName, config := range configFile.AuthConfigs {

		if strings.Contains(hostName, remoteRepo) {
			return &config, true
		}
	}

	return nil, false

}
