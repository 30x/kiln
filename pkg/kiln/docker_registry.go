package kiln

import (
	"fmt"
	"os"
	"strings"

	"github.com/30x/kiln/pkg/registry"
	"github.com/docker/engine-api/types"
)

//RegistryImageCreator is a struct that holds our pointer to the docker client
type RegistryImageCreator struct {
	//the client to docker
	client registry.Registry
	//the remote repository url
	dockerCreator ImageCreator
}

//NewRegistryImageCreator creates an instance of the RegistryImageCreator from the docker environment variables, and returns the instance
func NewRegistryImageCreator(repo string, reg registry.Registry) (ImageCreator, error) {
	// necessary for using a GCR registry, the project name needs to be in the image tag
	// and tagging is done by the LocalImageCreator before push
	if proj := reg.GetProjectName(); proj != "" {
		repo = fmt.Sprintf("%s/%s", repo, proj)
	}

	localDocker, err := NewLocalImageCreator(repo)

	if err != nil {
		return nil, err
	}

	return &RegistryImageCreator{
		client:        reg,
		dockerCreator: localDocker,
	}, nil
}

// GetOrganizations retrieves image repos in the registry
func (imageCreator RegistryImageCreator) GetOrganizations() (*[]string, error) {
	return nil, fmt.Errorf("GetOrganizations is unsupported for private registries at this time")
}

//GetApplications get all remote application for the specified repository
func (imageCreator RegistryImageCreator) GetApplications(repository string) (*[]string, error) {
	apps, err := imageCreator.client.ListRepositories(repository)

	if err != nil {
		return nil, err
	}

	return &apps, nil
}

//GetLocalImages return all local images
func (imageCreator RegistryImageCreator) GetLocalImages() (*[]types.Image, error) {
	return imageCreator.dockerCreator.GetLocalImages()
}

//BuildImage build the image
func (imageCreator RegistryImageCreator) BuildImage(dockerInfo *DockerBuild) (chan (string), error) {
	return imageCreator.dockerCreator.BuildImage(dockerInfo)
}

// PushImage pushes the iamge
func (imageCreator RegistryImageCreator) PushImage(dockerInfo *DockerInfo) (chan (string), error) {
	return imageCreator.dockerCreator.PushImage(dockerInfo)
}

//DeleteApplication Delete all images of the application from the remote repository.  Return an error if unable to do so.
func (imageCreator RegistryImageCreator) DeleteApplication(dockerInfo *DockerInfo, images *[]types.Image) error {
	name := fmt.Sprintf("%s/%s", dockerInfo.RepoName, dockerInfo.ImageName)

	for _, image := range *images {
		// with GCR Docker Registry, you must delete the tag before the manifest
		if os.Getenv("DOCKER_PROVIDER") == "gcr" {
			tag := strings.Split(image.RepoTags[0], ":")[1]
			err := imageCreator.client.DeleteImageTag(name, tag)
			if err != nil {
				LogError.Printf("Error received deleting a tag for %s: %v", name, err)
				return err
			}
		}

		err := imageCreator.client.DeleteImageManifest(name, image.ID)
		if err != nil {
			LogError.Printf("Error received deleting a manifest for %s: %v", name, err)
			return err
		}
	}

	return nil
}

//DeleteImageRevisionLocal Delete the image revision from the local repo
func (imageCreator RegistryImageCreator) DeleteImageRevisionLocal(sha string) error {
	//just delegate to the local docker impl
	return imageCreator.dockerCreator.DeleteImageRevisionLocal(sha)
}

//GetImageRevision get the image revision
func (imageCreator RegistryImageCreator) GetImageRevision(dockerInfo *DockerInfo) (*types.Image, error) {

	name := fmt.Sprintf("%s/%s", dockerInfo.RepoName, dockerInfo.ImageName)
	tag := dockerInfo.Revision

	blobDigest, _, err := imageCreator.client.GetImageBlobDigest(name, tag)
	if err != nil {
		return nil, err
	}

	blob, err := imageCreator.client.GetImageBlob(name, blobDigest)
	if err != nil {
		return nil, err
	}

	return &types.Image{
		Created:  blob.Created.Unix(),
		RepoTags: []string{fmt.Sprintf("%s:%s", name, tag)},
		ID:       blobDigest,
		Labels:   blob.Config.Labels,
	}, nil

}

//GetImages get all the images for the specified repository and application
func (imageCreator RegistryImageCreator) GetImages(repository string, application string) (*[]types.Image, error) {

	name := fmt.Sprintf("%s/%s", repository, application)

	tags, err := imageCreator.client.ListImageTags(name)
	if err != nil {
		return nil, err
	}

	images := make([]types.Image, len(tags))
	for ndx, tag := range tags {
		blobDigest, manifestDigest, err := imageCreator.client.GetImageBlobDigest(name, tag)
		if err != nil {
			return nil, err
		}

		blob, err := imageCreator.client.GetImageBlob(name, blobDigest)
		if err != nil {
			return nil, err
		}

		images[ndx] = types.Image{
			Created:  blob.Created.Unix(),
			RepoTags: []string{fmt.Sprintf("%s:%s", name, tag)},
			ID:       manifestDigest, //we return this so that this can be used to delete apps
			Labels:   blob.Config.Labels,
		}
	}

	return &images, nil
}
