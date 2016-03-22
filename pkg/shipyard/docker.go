package shipyard

import (
	"github.com/docker/engine-api/types"
	"io"
)

//ImageCreator the interface an ImageCreator instance must implement
type ImageCreator interface {

	//SearchLocalImages return all images with matching labels.  The label name is the key, the values are the value strings
	SearchLocalImages(search *DockerInfo) ([]types.Image, error)

	//SearchRemoteImages search remote images
	SearchRemoteImages(search *DockerInfo) ([]types.Image, error)

	//BuildImage creates a docker tar from the specified dockerInfo to the specified repo, image, and version.  Returns the reader stream or an error
	BuildImage(dockerInfo *DockerBuild, logs io.Writer) error

	//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
	PushImage(dockerInfo *DockerInfo, logs io.Writer) error

	//PullImage pull the specified image to our the docker runtime
	PullImage(dockerInfo *DockerInfo, logs io.Writer) error
}
