package shipyard

/**

This implementation supports the local docker API, as well as the docker provided remote registry
**/
import (
	"os"
	// "io"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
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
	_, err = client.ImageList(types.ImageListOptions{All: false})

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
	var transport *http.Transport
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	dockerTLSVerify := os.Getenv("DOCKER_TLS_VERIFY")

	if dockerCertPath != "" && dockerTLSVerify != "" {
		tlsc := &tls.Config{}

		cert, err := tls.LoadX509KeyPair(filepath.Join(dockerCertPath, "cert.pem"), filepath.Join(dockerCertPath, "key.pem"))
		if err != nil {
			return nil, fmt.Errorf("Error loading x509 key pair: %s", err)
		}

		tlsc.Certificates = append(tlsc.Certificates, cert)
		tlsc.InsecureSkipVerify = dockerTLSVerify == ""
		transport = &http.Transport{
			TLSClientConfig: tlsc,
		}
	}

	dockerHost := os.Getenv("DOCKER_HOST")

	if dockerHost == "" {
		LogError.Fatalf("You must sent the DOCKER_HOST env variable")
	}

	parts := strings.SplitN(dockerHost, "://", 2)

	if len(parts) != 2 {
		LogError.Fatalf("You must specify a protocol on your server.  tcp:// or http:// is required")
	}

	return client.NewClient(dockerHost, os.Getenv("DOCKER_API_VERSION"), transport, nil)
}

//SearchLocalImages return all images with matching labels.  The label name is the key, the values are the value strings
func (imageCreator LocalImageCreator) SearchLocalImages(search *DockerInfo) (*[]types.Image, error) {

	filter := createFilter(search)

	opts := types.ImageListOptions{All: false, Filters: *filter}

	images, err := imageCreator.client.ImageList(opts)

	return &images, err
}

//SearchRemoteImages search remote images
func (imageCreator LocalImageCreator) SearchRemoteImages(search *DockerInfo) (*[]types.Image, error) {
	//TODO connect to remote repo here

	filter := createFilter(search)

	opts := types.ImageListOptions{All: false, Filters: *filter}

	images, err := imageCreator.client.ImageList(opts)

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
func (imageCreator LocalImageCreator) BuildImage(dockerInfo *DockerBuild, logs io.Writer) error {

	name := dockerInfo.GetTagName()

	LogInfo.Printf("Started uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	inputReader, err := os.Open(dockerInfo.TarFile)

	if err != nil {
		LogError.Printf("Unable to open tar file "+dockerInfo.TarFile+"for input", err)
		return err
	}

	imageBuildOptions := types.ImageBuildOptions{
		Context: inputReader,
		Tags:    []string{name},
	}

	response, err := imageCreator.client.ImageBuild(imageBuildOptions)

	if err != nil {
		LogInfo.Fatal(err)
		return err
	}

	io.Copy(logs, response.Body)
	response.Body.Close()

	LogInfo.Printf("Completed uploading image with name %s and tar file %s", name, dockerInfo.TarFile)

	return nil

}

//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
func (imageCreator LocalImageCreator) PushImage(dockerInfo *DockerInfo, logs io.Writer) error {

	localTag := dockerInfo.GetTagName()
	remoteRepo := imageCreator.remoteRepo
	remoteTag := dockerInfo.GetRemoteTagName(remoteRepo)
	revision := dockerInfo.Revision

	imageTagOptions := types.ImageTagOptions{
		Force:          true,
		ImageID:        localTag,
		RepositoryName: remoteTag,
		Tag:            revision,
	}

	err := imageCreator.client.ImageTag(imageTagOptions)

	if err != nil {
		return err
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

	reader, err := imageCreator.client.ImagePush(imagePushOptions, privledgedFunction)

	if err != nil {
		return err
	}

	io.Copy(logs, reader)

	reader.Close()

	return nil
}

//PullImage pull the specified image to our the docker runtime
func (imageCreator LocalImageCreator) PullImage(dockerInfo *DockerInfo, logs io.Writer) error {

	remoteRepo := imageCreator.remoteRepo
	remoteTag := dockerInfo.GetRemoteTagName(remoteRepo)
	revision := dockerInfo.Revision

	imagePullOpts := types.ImagePullOptions{
		ImageID: remoteTag,
		Tag:     revision,
	}

	//this call on every invocation is deliberate.  Our keys should rotate and we want to continue to function when
	//that happens
	authConfig := generateAuthConfiguration(imageCreator.remoteRepo)

	imagePullOpts.RegistryAuth = authConfig

	privledgedFunction := func() (string, error) {
		return authConfig, nil
	}

	response, error := imageCreator.client.ImagePull(imagePullOpts, privledgedFunction)

	if error != nil {
		return error
	}

	io.Copy(logs, response)

	response.Close()

	return nil
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
