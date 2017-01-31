package kiln

/**

This implementation supports the local docker API, as well as the docker provided remote registry
**/
import (
	"os"

	"golang.org/x/net/context"
	// "io"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/cliconfig"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

//LocalImageCreator is a struct that holds our pointer to the docker client
type LocalImageCreator struct {
	//the client to docker
	client *client.Client
	//the remote repository url
	remoteRepo string
}

//NewLocalImageCreator creates an instance of the LocalImageCreator from the docker environment variables, and returns the instance
func NewLocalImageCreator(repo string) (ImageCreator, error) {

	client, err := newEnvClient()
	if err != nil {
		return nil, err
	}

	imageCreator := LocalImageCreator{
		client:     client,
		remoteRepo: repo,
	}

	//TODO, stop selecting all
	_, err = client.ImageList(context.Background(), types.ImageListOptions{All: false})

	if err != nil {
		return nil, err
	}

	return imageCreator, nil
}

// newEnvClient initializes a new API client based on environment variables.. Taken from github.com/docker/engine-api/client.go and fixes bug when no TLS is present
// as well as adds validation to the conneciton URL to give users a more useful error message
// Use DOCKER_HOST to set the url to the docker server.
// Use DOCKER_API_VERSION to set the version of the API to reach, leave empty for latest.
// Use DOCKER_CERT_PATH to load the tls certificates from.
// Use DOCKER_TLS_VERIFY to enable or disable TLS verification, off by default.
func newEnvClient() (*client.Client, error) {
	return client.NewEnvClient()
}

//GetOrganizations get all remote repositories
func (imageCreator LocalImageCreator) GetOrganizations() (*[]string, error) {
	opts := types.ImageListOptions{All: false}

	images, err := imageCreator.client.ImageList(context.Background(), opts)

	if err != nil {
		return nil, err
	}

	return getTags(&images, TAG_REPO), nil
}

//GetApplications get all remote application for the specified repository
func (imageCreator LocalImageCreator) GetApplications(repository string) (*[]string, error) {
	filters := filters.NewArgs()

	//append filters as required based on the input

	newFilter := TAG_REPO + "=" + repository
	filters.Add("label", newFilter)

	opts := types.ImageListOptions{All: false, Filters: filters}

	images, err := imageCreator.client.ImageList(context.Background(), opts)

	if err != nil {
		return nil, err
	}

	return getTags(&images, TAG_APPLICATION), nil
}

//GetImages get all the images for the specified repository and application
func (imageCreator LocalImageCreator) GetImages(repository string, application string) (*[]types.Image, error) {

	filters := filters.NewArgs()

	//append filters as required based on the input

	repoFilter := TAG_REPO + "=" + repository
	filters.Add("label", repoFilter)

	applicationFilter := TAG_APPLICATION + "=" + application
	filters.Add("label", applicationFilter)

	opts := types.ImageListOptions{All: false, Filters: filters}

	images, err := imageCreator.client.ImageList(context.Background(), opts)

	return &images, err
}

//GetImageRevision get the image revision
func (imageCreator LocalImageCreator) GetImageRevision(dockerInfo *DockerInfo) (*types.Image, error) {

	filters := filters.NewArgs()

	//append filters as required based on the input

	repoFilter := TAG_REPO + "=" + dockerInfo.RepoName
	filters.Add("label", repoFilter)

	applicationFilter := TAG_APPLICATION + "=" + dockerInfo.ImageName
	filters.Add("label", applicationFilter)

	revisionFilter := TAG_REVISION + "=" + dockerInfo.Revision
	filters.Add("label", revisionFilter)

	opts := types.ImageListOptions{All: false, Filters: filters}

	images, err := imageCreator.client.ImageList(context.Background(), opts)

	if err == nil && len(images) > 0 {
		return &images[0], err
	}

	return nil, err

}

//DeleteImageRevisionLocal Delete the image revision from the local repo
func (imageCreator LocalImageCreator) DeleteImageRevisionLocal(sha string) error {

	imageRemoveOptions := types.ImageRemoveOptions{
		Force:         true,
		ImageID:       sha,
		PruneChildren: false,
	}
	deleted, err := imageCreator.client.ImageRemove(context.Background(), imageRemoveOptions)

	if err != nil {
		return err
	}

	if deleted == nil || len(deleted) == 0 {
		return fmt.Errorf("Could not find an image with sha %s to delete", sha)
	}

	//otherwise we were successful
	return nil
}

//DeleteApplication Delete all images of the application from the remote repository.  Return an error if unable to do so.
func (imageCreator LocalImageCreator) DeleteApplication(dockerInfo *DockerInfo, images *[]types.Image) error {
	for _, image := range *images {

		_, err := imageCreator.client.ImageRemove(context.Background(), types.ImageRemoveOptions{
			Force:         true,
			ImageID:       image.ID,
			PruneChildren: true,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

//parse the tag out of the returned image
func getTags(images *[]types.Image, tagToParse string) *[]string {
	nameSet := NewStringSet()

	for _, image := range *images {
		repo := image.Labels[tagToParse]

		if repo == "" {
			continue
		}

		nameSet.Add(repo)
	}

	return nameSet.AsSlice()

}

//GetLocalImages return all local images
func (imageCreator LocalImageCreator) GetLocalImages() (*[]types.Image, error) {

	opts := types.ImageListOptions{All: false}

	images, err := imageCreator.client.ImageList(context.Background(), opts)

	return &images, err
}

//createFilter generate filter from search
func createFilter(search *DockerInfo) *filters.Args {
	filters := filters.NewArgs()

	//append filters as required based on the input
	if search.RepoName != "" {
		newFilter := TAG_REPO + "=" + search.RepoName
		filters.Add("label", newFilter)
	}

	if search.ImageName != "" {
		newFilter := TAG_APPLICATION + "=" + search.ImageName
		filters.Add("label", newFilter)
	}

	if search.Revision != "" {
		newFilter := TAG_REVISION + "=" + search.Revision
		filters.Add("label", newFilter)
	}

	return &filters
}

//BuildImage creates a docker tar from the specified dockerInfo to the specified repo, image, and version.  Returns the reader stream or an error
func (imageCreator LocalImageCreator) BuildImage(dockerInfo *DockerBuild) (chan (string), error) {

	name := dockerInfo.GetTagName()

	LogInfo.Printf("Started uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	inputReader, err := os.Open(dockerInfo.TarFile)

	if err != nil {
		LogError.Printf("Unable to open tar file "+dockerInfo.TarFile+"for input", err)
		return nil, err
	}

	imageBuildOptions := types.ImageBuildOptions{
		Context:     inputReader,
		Tags:        []string{name},
		ForceRemove: true,
		Remove:      true,
		NoCache:     true,
	}

	response, err := imageCreator.client.ImageBuild(context.Background(), imageBuildOptions)

	if err != nil {
		LogInfo.Fatal(err)
		return nil, err
	}

	LogInfo.Printf("Parsing response from uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	//conver it into a channel

	streamParser := NewBuildStreamParser(response.Body)

	//start consuming/emitting
	go streamParser.Parse()

	return streamParser.Channel(), nil

}

//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
func (imageCreator LocalImageCreator) PushImage(dockerInfo *DockerInfo) (chan (string), error) {

	localTag := dockerInfo.GetTagName()
	remoteRepo := imageCreator.remoteRepo
	remoteTag := dockerInfo.GetRemoteTagPath(remoteRepo)
	revision := dockerInfo.Revision

	imageTagOptions := types.ImageTagOptions{
		Force:          true,
		ImageID:        localTag,
		RepositoryName: remoteTag,
		Tag:            revision,
	}

	LogInfo.Printf("Tagging local image id of %s with remote tag of %s and tag %s", localTag, remoteTag, revision)

	err := imageCreator.client.ImageTag(context.Background(), imageTagOptions)

	if err != nil {
		return nil, err
	}

	imagePushOptions := types.ImagePushOptions{
		ImageID: remoteTag,
		Tag:     revision,
	}

	//this call on every invocation is deliberate.  Our keys should rotate and we want to continue to function when
	//that happens
	authConfig := generateAuthConfiguration(imageCreator.remoteRepo)

	imagePushOptions.RegistryAuth = authConfig

	//not sure why we need this when authconfig is already provided, but required at the API level
	privledgedFunction := func() (string, error) {
		return authConfig, nil
	}

	reader, err := imageCreator.client.ImagePush(context.Background(), imagePushOptions, privledgedFunction)

	if err != nil {
		return nil, err
	}

	streamParser := NewPushStreamParser(reader)

	//start consuming/emitting
	go streamParser.Parse()

	return streamParser.Channel(), nil
}

//GenerateRepoURI generate the repo uri
func (imageCreator LocalImageCreator) GenerateRepoURI(dockerInfo *DockerInfo) string {

	return dockerInfo.GetRemoteTagName(imageCreator.remoteRepo)
}

//generateAuthConfiguration Create an auth configuration from the environment variables
func generateAuthConfiguration(remoteRepo string) string {

	authConfig, exists := getAuthConfig(remoteRepo)

	if !exists {
		LogWarn.Printf("Could not find repo %s in auth configuration.  Returning empty auth", authConfig)
		authConfig = &types.AuthConfig{}
	}

	encoded, err := encodeAuthToBase64(authConfig)

	if err != nil {
		LogWarn.Printf("Could not marshall credentials for encoding, returning empty string")
		return "{}"
	}

	return encoded
}

func encodeAuthToBase64(authConfig *types.AuthConfig) (string, error) {
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

//getAuthConfig return the auth config if it exists.  Nil and false otherwise
func getAuthConfig(remoteRepo string) (*types.AuthConfig, bool) {

	configFile, e := cliconfig.Load(cliconfig.ConfigDir())

	//ou
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
