package shipyard

import (
	"io"

	"github.com/docker/engine-api/types"
)

//ImageCreator the interface an ImageCreator instance must implement
type ImageCreator interface {
	//GetRepositories get all remote Namespaces
	GetNamespaces() (*[]string, error)

	//GetApplications get all remote application for the specified repository
	GetApplications(repository string) (*[]string, error)

	//GetImages get all the images for the specified repository and application
	GetImages(repository string, application string) (*[]types.Image, error)

	//GetImageRevision get the image for the specified repository, application, and revision.  Nil is returned if one does not exist
	GetImageRevision(dockerInfo *DockerInfo) (*types.Image, error)

	//DeleteImageRevisionLocal Delete the image from the local machine repository.  Return an error if unable to do so.  Should not be called from outside the package
	deleteImageRevisionLocal(dockerInfo *DockerInfo) error

	//CleanImageRevision clean the image revision
	CleanImageRevision(dockerInfo *DockerInfo) error

	//GetLocalImages return all local images
	GetLocalImages() (*[]types.Image, error)

	//BuildImage creates a docker tar from the specified dockerInfo to the specified repo, image, and version.  Returns the reader stream or an error
	BuildImage(dockerInfo *DockerBuild, logs io.Writer) error

	//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
	PushImage(dockerInfo *DockerInfo, logs io.Writer) error

	//PullImage pull the specified image to our the docker runtime
	PullImage(dockerInfo *DockerInfo, logs io.Writer) error
}
