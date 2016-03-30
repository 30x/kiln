package shipyard

import (
	"io"
	"strings"

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

//PullImage pull the specified image to our the docker runtime
func (imageCreator EcsImageCreator) PullImage(dockerInfo *DockerInfo, logs io.Writer) error {

	return nil
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

//Search Each search will implement their ECS calls differently, as a result, we simply intanciate the correct type and delegate to it.  An impl of the command pattern
type Search interface {
	//Search
	search() (*[]types.Image, error)
}

//searchAll The search for all repositories
type searchAll struct {
	awsClient *ecr.ECR
}

func (searchAll *searchAll) search() (*[]types.Image, error) {

	//initialize the filter map
	describeRequest := &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(100),
	}

	results := []types.Image{}

	for {

		response, err := searchAll.awsClient.DescribeRepositories(describeRequest)

		//call failed, bail
		if err != nil {
			return nil, err
		}

		for _, repository := range response.Repositories {

			awsImages, err := searchAll.getImages(repository.RepositoryName)

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

type searchRepository struct {
	awsClient *ecr.ECR
}

func (searchRepository *searchRepository) search() (*[]types.Image, error) {

	//initialize the filter map
	describeRequest := &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(100),
		// RepositoryNames: []*string{
		// 	aws.String(searchString),
		// },
	}

	results := []types.Image{}

	for {

		response, err := searchRepository.awsClient.DescribeRepositories(describeRequest)

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

type searchApplication struct {
	awsClient *ecr.ECR
}

func (search *searchApplication) search() (*[]types.Image, error) {

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

type searchApplicationRevision struct {
}

func (search *searchApplicationRevision) search() (*[]types.Image, error) {

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
