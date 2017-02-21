package kiln

import (
	"errors"
	"os"

	"github.com/30x/kiln/pkg/registry"
)

//NewImageCreatorFromEnv instanciate the image creator from the environment variables.  If an env is missing, an error will be logged and execution will halt
func NewImageCreatorFromEnv() (ImageCreator, error) {

	creatorValue := os.Getenv("DOCKER_PROVIDER")

	if creatorValue == "" {
		return nil, errors.New("You most specifity the DOCKER_PROVIDER environment variable.  Valid values are 'docker' or 'ecr'")
	}

	repoURL := os.Getenv("DOCKER_REGISTRY_URL")

	if repoURL == "" {
		return nil, errors.New("You must set the DOCKER_REGISTRY_URL environment variable.  An example would be localhost:5000")
	}

	switch creatorValue {
	case "private":
		return getPrivateImpl(repoURL)
	case "gcr":
		return getGcrImpl(repoURL)
	case "docker":
		return NewLocalImageCreator(repoURL)
	case "ecr":
		return getEcrImpl(repoURL)
	default:
		return nil, errors.New("You most specifity the DOCKER_PROVIDER environment variable.  Valid values are 'docker' or 'ecr'")
	}
}

//get the ecr image impl
func getEcrImpl(dockerRegistryURL string) (ImageCreator, error) {

	ecrRegion := os.Getenv("ECR_REGION")

	if ecrRegion == "" {
		return nil, errors.New("You must set the ECR_REGION environment variable.  An example would be 'us-east-1'")
	}

	return NewEcsImageCreator(dockerRegistryURL, ecrRegion)
}

func getPrivateImpl(dockerRegistryURL string) (ImageCreator, error) {
	// registry API is exposed via a svc
	// e.g. running hosted reg in k8s, push images to localhost:5000, but API server is elsewhere
	regAPIServer := os.Getenv("REGISTRY_API_SERVER")

	if regAPIServer == "" {
		return nil, errors.New("You must set the REGISTRY_API_SERVER environment variable.  An example would be 'https://gcr.io'")
	}

	reg := registry.NewPrivateRegistryClient(regAPIServer)

	return NewRegistryImageCreator(dockerRegistryURL, reg)
}

func getGcrImpl(dockerRegistryURL string) (ImageCreator, error) {
	regAPIServer := os.Getenv("REGISTRY_API_SERVER")

	if regAPIServer == "" {
		return nil, errors.New("You must set the REGISTRY_API_SERVER environment variable.  An example would be 'https://gcr.io'")
	}

	reg, err := registry.NewGoogleContainerRegistryClient(regAPIServer)
	if err != nil {
		return nil, err
	}

	return NewRegistryImageCreator(dockerRegistryURL, reg)
}
