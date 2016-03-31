package shipyard

import (
	"errors"
	"io"
	"strings"

	"fmt"

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

//GetRepositories get all remote repositories
func (imageCreator EcsImageCreator) GetRepositories() (*[]string, error) {

	repositoryResult := newRepositoryResult()

	err := imageCreator.getResults(repositoryResult)

	return repositoryResult.getResults(), err
}

//GetApplications get all remote application for the specified repository
func (imageCreator EcsImageCreator) GetApplications(repository string) (*[]string, error) {
	applicationResult := newApplicationResult()

	err := imageCreator.getResults(applicationResult)

	return applicationResult.getResults(), err

}

//GetImages get all the images for the specified repository and application
func (imageCreator EcsImageCreator) GetImages(repository string, application string) (*[]types.Image, error) {

	repositoryName := repository + "/" + application

	listRequest := &ecr.ListImagesInput{
		MaxResults:     aws.Int64(100),
		RepositoryName: &repositoryName,
	}

	results := []types.Image{}

	for {

		response, err := imageCreator.awsClient.ListImages(listRequest)

		//call failed, bail
		if err != nil {
			return nil, err
		}

		for _, awsImage := range response.ImageIds {
			if awsImage.ImageDigest == nil || awsImage.ImageTag == nil {
				LogWarn.Printf("Was unable to marshall response from aws image, skipping")
				continue
			}

			repoTag := repository + ":" + *awsImage.ImageTag

			dockerImage := types.Image{
				ID:       *awsImage.ImageDigest,
				RepoTags: []string{repoTag},
			}

			results = append(results, dockerImage)
		}

		//no token, break out of the loop
		if response.NextToken == nil {
			break
		}

		//otherwise reset our describe for the next request
		listRequest.NextToken = response.NextToken

	}

	return &results, nil

}

//GetImageRevision get the image revision
func (imageCreator EcsImageCreator) GetImageRevision(repository string, application string, revision string) (*types.Image, error) {
	repositoryName := repository + "/" + application

	imageID := &ecr.ImageIdentifier{
		ImageTag: &revision,
	}

	imageRequest := ecr.BatchGetImageInput{
		RepositoryName: &repositoryName,
		ImageIds:       []*ecr.ImageIdentifier{imageID},
	}

	response, err := imageCreator.awsClient.BatchGetImage(&imageRequest)

	//call failed, bail
	if err != nil {
		return nil, err
	}

	if len(response.Images) < 1 {
		errorMsg := fmt.Sprintf("Could not find images with repository %s, application %s, and revision %s", repository, application, revision)

		return nil, errors.New(errorMsg)
	}

	awsImage := response.Images[0].ImageId

	if awsImage.ImageDigest == nil || awsImage.ImageTag == nil {
		LogWarn.Printf("Was unable to marshall response from aws image, skipping")
		return nil, nil
	}

	repoTag := repositoryName + ":" + *awsImage.ImageTag

	dockerImage := types.Image{
		ID:       *awsImage.ImageDigest,
		RepoTags: []string{repoTag},
	}

	return &dockerImage, nil
}

//GetLocalImages return all local images
func (imageCreator EcsImageCreator) GetLocalImages() (*[]types.Image, error) {
	return imageCreator.dockerCreator.GetLocalImages()
}

//iterate all the results and add them to the result builder  TODO paging client side
func (imageCreator EcsImageCreator) getResults(resultBuilder resultBuilder) error {
	describeRequest := &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(100),
	}

	for {

		response, err := imageCreator.awsClient.DescribeRepositories(describeRequest)

		//call failed, bail
		if err != nil {
			return err
		}

		for _, repository := range response.Repositories {

			name := repository.RepositoryName

			resultBuilder.add(name)

		}

		//no token, break out of the loop
		if response.NextToken == nil {
			break
		}

		//otherwise reset our describe for the next request
		describeRequest.NextToken = response.NextToken
	}

	return nil
}

type resultBuilder interface {
	//possibly add the repository to the results
	add(repositoryName *string)

	//Get the results
	getResults() *[]string
}

type repositoryResult struct {
	repoSet *StringSet
}

//create a new instance of the repository result
func newRepositoryResult() *repositoryResult {
	return &repositoryResult{
		repoSet: NewStringSet(),
	}
}

func (repositoryResult *repositoryResult) add(repositoryName *string) {
	parts := strings.Split(*repositoryName, "/")

	//not the right length, drop it
	if len(parts) != 2 {
		return
	}

	repositoryResult.repoSet.Add(parts[0])
}

//Get the results
func (repositoryResult *repositoryResult) getResults() *[]string {
	return repositoryResult.repoSet.AsSlice()
}

type applicationResult struct {
	repoSet *StringSet
}

//create a new instance of the application result
func newApplicationResult() *applicationResult {
	return &applicationResult{
		repoSet: NewStringSet(),
	}
}

func (applicationResult *applicationResult) add(repositoryName *string) {
	parts := strings.Split(*repositoryName, "/")

	//not the right length, drop it
	if len(parts) != 2 {
		return
	}

	applicationResult.repoSet.Add(parts[1])
}

//Get the results
func (applicationResult *applicationResult) getResults() *[]string {
	return applicationResult.repoSet.AsSlice()
}

//BuildImage build the image
func (imageCreator EcsImageCreator) BuildImage(dockerInfo *DockerBuild, logs io.Writer) error {
	return imageCreator.dockerCreator.BuildImage(dockerInfo, logs)
}

//PullImage pull the specified image to our the docker runtime
func (imageCreator EcsImageCreator) PullImage(dockerInfo *DockerInfo, logs io.Writer) error {
	return imageCreator.dockerCreator.PullImage(dockerInfo, logs)
}

//PushImage pushes the remotely tagged image to docker. Returns a reader of the stream, or an error
func (imageCreator EcsImageCreator) PushImage(dockerInfo *DockerInfo, logs io.Writer) error {

	//check if it exists on ecs, if not create it first
	imageName := dockerInfo.GetImageName()

	exists, err := imageCreator.imageExists(imageName)

	if err != nil {
		return err
	}

	if !exists {
		err := imageCreator.createImage(imageName)

		if err != nil {
			return err
		}
	}

	return imageCreator.dockerCreator.PushImage(dockerInfo, logs)
}

//return true if the image exists
func (imageCreator EcsImageCreator) imageExists(imageName string) (bool, error) {
	describeRequest := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []*string{&imageName},
	}

	response, err := imageCreator.awsClient.DescribeRepositories(describeRequest)

	LogInfo.Printf("Received response %v with error %s with type", response, err)

	// RepositoryNotFoundException
	if err != nil && isNotFoundError(err) {
		return false, nil
	}

	return err == nil && len(response.Repositories) == 1, err

}

func isNotFoundError(err error) bool {
	errorString := err.Error()

	//feels hacky, but we can't cast this error type because it's a private type
	return strings.Contains(errorString, "RepositoryNotFoundException")
}

func (imageCreator EcsImageCreator) createImage(imageName string) error {
	describeRequest := &ecr.CreateRepositoryInput{
		RepositoryName: &imageName,
	}

	_, err := imageCreator.awsClient.CreateRepository(describeRequest)

	return err

}
