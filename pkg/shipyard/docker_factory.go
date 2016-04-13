package shipyard

import (
	"errors"
	"os"
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
