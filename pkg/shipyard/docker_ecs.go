package shipyard

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/docker/engine-api/types"
)

//EcsImageCreator is a struct that holds our pointer to the docker client
type EcsImageCreator struct {
	//the client aws

	awsClient *ecr.ECR

	//the remote repository url
	dockerCreator ImageCreator
}

//NewEcsImageCreator creates an instance of the EcsImageCreator from the docker environment variables, and returns the instance
func NewEcsImageCreator(repo string, region string) (ImageCreator, error) {
	//
	awsClient := ecr.New(session.New(), &aws.Config{Region: aws.String(region)})

	//only try to pull a single repository as a test
	describeRequest := &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(1),
	}

	_, err := awsClient.DescribeRepositories(describeRequest)

	if err != nil {
		return nil, err
	}

	localDocker, err := NewLocalImageCreator(repo)

	if err != nil {
		return nil, err
	}

	return &EcsImageCreator{
		awsClient:     awsClient,
		dockerCreator: localDocker,
	}, nil
}

//SearchRemoteImages return all images with matching labels.  The label name is the key, the values are the value strings
func (imageCreator EcsImageCreator) SearchRemoteImages(search *DockerInfo) (*[]types.Image, error) {

	//revision exists, perform a search for this revision
	if search.Revision != "" {
		//  search.Revision
		repoString := search.GetImageName()

		LogInfo.Printf("Searching for revision %s in repo %s", search.Revision, repoString)

		//revi
		return &[]types.Image{}, nil
	}

	//initialize the filter map
	describeRequest := &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(100),
		// RepositoryNames: []*string{
		// 	aws.String(searchString),
		// },
	}

	results := []types.Image{}

	for {

		response, err := imageCreator.awsClient.DescribeRepositories(describeRequest)

		//call failed, bail
		if err != nil {
			return nil, err
		}

		for _, repository := range response.Repositories {

			awsImages, err := imageCreator.getImages(repository.RepositoryName)

			if err != nil {
				return nil, err
			}

			for _, awsImage := range awsImages {

				if awsImage.ImageDigest == nil || awsImage.ImageTag == nil {
					LogWarn.Printf("Was unable to marshall response from aws image, skipping")
					continue
				}

				repoTag := *repository.RepositoryName + ":" + *awsImage.ImageTag

				dockerImage := types.Image{
					ID:       *awsImage.ImageDigest,
					RepoTags: []string{repoTag},
				}

				results = append(results, dockerImage)
			}
		}

		//no token, break out of the loop
		if response.NextToken == nil {
			break
		}

		//otherwise reset our describe for the next request
		describeRequest.NextToken = response.NextToken

	}

	return &results, nil
}

//SearchLocalImages return all images with matching labels.  The label name is the key, the values are the value strings
func (imageCreator EcsImageCreator) SearchLocalImages(search *DockerInfo) (*[]types.Image, error) {
	return imageCreator.dockerCreator.SearchLocalImages(search)
}

//BuildImage creates a docker tar from the specified dockerInfo to the specified repo, image, and version.  Returns the reader stream or an error
func (imageCreator EcsImageCreator) BuildImage(dockerInfo *DockerBuild, logs io.Writer) error {

	return imageCreator.dockerCreator.BuildImage(dockerInfo, logs)

}

//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
func (imageCreator EcsImageCreator) PushImage(dockerInfo *DockerInfo, logs io.Writer) error {

	//check if it exists on ecs, if not create it first
	imageName := dockerInfo.GetImageName()

	if !imageCreator.imageExists(imageName) {
		err := imageCreator.createImage(imageName)

		if err != nil {
			return err
		}
	}

	return imageCreator.dockerCreator.PushImage(dockerInfo, logs)
}

//PullImage pull the specified image to our the docker runtime
func (imageCreator EcsImageCreator) PullImage(dockerInfo *DockerInfo, logs io.Writer) error {

	return nil
}

//return true if the image exists
func (imageCreator EcsImageCreator) imageExists(imageName string) bool {
	describeRequest := &ecr.DescribeRepositoriesInput{
		MaxResults:      aws.Int64(1),
		RepositoryNames: []*string{&imageName},
	}

	response, err := imageCreator.awsClient.DescribeRepositories(describeRequest)

	return err == nil && len(response.Repositories) == 1

}

func (imageCreator EcsImageCreator) createImage(imageName string) error {
	describeRequest := &ecr.CreateRepositoryInput{
		RepositoryName: &imageName,
	}

	_, err := imageCreator.awsClient.CreateRepository(describeRequest)

	return err

}

//getImages get all images for the repositoryName
func (imageCreator EcsImageCreator) getImages(repositoryName *string) ([]*ecr.ImageIdentifier, error) {

	listRequest := &ecr.ListImagesInput{
		MaxResults:     aws.Int64(100),
		RepositoryName: repositoryName,
	}

	results := []*ecr.ImageIdentifier{}

	for {

		response, err := imageCreator.awsClient.ListImages(listRequest)

		//call failed, bail
		if err != nil {
			return nil, err
		}

		for _, image := range response.ImageIds {
			results = append(results, image)
		}

		//no token, break out of the loop
		if response.NextToken == nil {
			break
		}

		//otherwise reset our describe for the next request
		listRequest.NextToken = response.NextToken

	}

	return results, nil
}
