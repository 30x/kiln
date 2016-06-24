package shipyard_test

import (
	"time"

	. "github.com/30x/shipyard/pkg/shipyard"
	"github.com/docker/engine-api/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"strings"
	// "github.com/docker/engine-api/types"
)

var _ = Describe("docker", func() {

	var imageCreator ImageCreator

	//Tests for all image operations
	Describe("Image Operations", func() {

		AssertImageTests := func() {
			It("Should create image successfully", func() {
				repoName := "test" + UUIDString()
				imageName := "test"
				revision := "v1.0"

				createImage(imageCreator, repoName, imageName, revision)
			})

			It("Tag and Push", func() {
				repoName := "test" + UUIDString()
				imageName := "test"
				revision := "v1.0"

				_, dockerInfo := createImage(imageCreator, repoName, imageName, revision)

				stream, err := imageCreator.PushImage(dockerInfo)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				channelToOutput(stream)

				assertImageExists(imageCreator, dockerInfo)

			})

			It("Test Search", func() {

				//push first image
				repoName := "test" + UUIDString()
				imageName1 := "test1"
				revision10 := "v1.0"

				_, dockerInfo10 := createImage(imageCreator, repoName, imageName1, revision10)

				stream, err := imageCreator.PushImage(dockerInfo10)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				channelToOutput(stream)

				revision11 := "v1.1"

				_, dockerInfo11 := createImage(imageCreator, repoName, imageName1, revision11)

				stream, err = imageCreator.PushImage(dockerInfo11)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				channelToOutput(stream)

				//push second image
				imageName2 := "test2"

				_, dockerInfo2 := createImage(imageCreator, repoName, imageName2, revision10)

				stream, err = imageCreator.PushImage(dockerInfo2)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				channelToOutput(stream)

				//applications
				assertApplicationsExist(imageCreator, dockerInfo2.RepoName, dockerInfo10.ImageName, dockerInfo11.ImageName, dockerInfo2.ImageName)

				//images
				assertImageExists(imageCreator, dockerInfo10)
				assertImageExists(imageCreator, dockerInfo11)
				assertImageExists(imageCreator, dockerInfo2)
			})

			FIt("Test Cross Namespace", func() {

				//push first image
				repoName1 := "test" + UUIDString()
				repoName2 := "test" + UUIDString()
				imageName1 := "test1"
				imageName2 := "test2"
				revision := "v1.0"

				_, dockerInfo1 := createImage(imageCreator, repoName1, imageName1, revision)

				stream, err := imageCreator.PushImage(dockerInfo1)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				channelToOutput(stream)

				_, dockerInfo2 := createImage(imageCreator, repoName2, imageName2, revision)

				stream, err = imageCreator.PushImage(dockerInfo2)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				channelToOutput(stream)

				applications, err := imageCreator.GetApplications(repoName1)

				Expect(err).Should(BeNil(), "Unable to get applications", err)

				Expect(len(*applications)).Should(Equal(1), "Only 1 application should be returned")

				appName := (*applications)[0]

				Expect(appName).Should(Equal(imageName1))

				applications, err = imageCreator.GetApplications(repoName2)

				Expect(err).Should(BeNil(), "Unable to get applications", err)

				Expect(len(*applications)).Should(Equal(1), "Only 1 application should be returned")

				appName = (*applications)[0]
				Expect(appName).Should(Equal(imageName2))

			})

			It("Reap", func() {
				repoName := "test" + UUIDString()
				imageName := "test"
				revision := "v1.0"

				_, dockerInfo := createImage(imageCreator, repoName, imageName, revision)

				exists := searchLocalImages(imageCreator, dockerInfo, dockerInfo)

				Expect(exists).Should(BeTrue(), "Image should exist locally")

				//now reap it

				err := Reap(time.Duration(0)*time.Second, imageCreator)

				Expect(err).Should(BeNil(), "Unable to reap images.  Error is %s", err)

				//now it should be deleted, see if it exists
				exists = searchLocalImages(imageCreator, dockerInfo, dockerInfo)

				Expect(exists).Should(BeFalse(), "Image should not exist locally")

			})

		}

		//test the local machine impl
		Context("Local Docker Machine", func() {
			BeforeEach(func() {
				var error error
				imageCreator, error = NewLocalImageCreator("localhost:5000")

				Expect(error).Should(BeNil(), "Could not create local docker image creator")

			})

			AssertImageTests()
		})

		Context("Amazon ECS", func() {
			BeforeEach(func() {
				var error error
				imageCreator, error = NewEcsImageCreator("977777657611.dkr.ecr.us-east-1.amazonaws.com", "us-east-1")

				Expect(error).Should(BeNil(), "Could not create local docker image creator")

			})

			AssertImageTests()
		})

	})

})

//helper functions called within the tests
func createImage(imageCreator ImageCreator, repoName string, appName string, revision string) (*SourceInfo, *DockerInfo) {

	const validTestZip = "../../testresources/echo-test.zip"

	workspace, dockerInfo := doSetup(validTestZip, repoName, appName, revision)

	//clean up the workspace after the test.  Comment this out for debugging
	//defer workspace.Clean()

	dockerImage := &DockerBuild{
		TarFile:    workspace.TargetTarName,
		DockerInfo: dockerInfo,
	}

	//copy over our docker file.  These tests assume io has been tested and works properly

	stream, err := imageCreator.BuildImage(dockerImage)

	Expect(err).Should(BeNil(), "Unable to build image", err)

	channelToOutput(stream)

	//get the image from docker and ensure it exists

	assertLocalImageExists(imageCreator, dockerImage.DockerInfo, &DockerInfo{})

	//pull by label

	search := &DockerInfo{
		RepoName:  dockerInfo.RepoName,
		ImageName: dockerInfo.ImageName,
		Revision:  dockerInfo.Revision,
	}

	assertLocalImageExists(imageCreator, dockerImage.DockerInfo, search)

	return workspace, dockerImage.DockerInfo

}

//DoSetup Copies the specified inputZip file into the source directory and adds the docker file to it
func doSetup(inputZip string, repoName string, appName string, revision string) (*SourceInfo, *DockerInfo) {

	//copy over our docker file.  These tests assume io has been tested and works properly

	workspace, err := CreateNewWorkspace()

	if err != nil {
		Fail("Should not return an error creating a valid workspace")
	}

	//change the source zip input for extactraction
	CopyFile(inputZip, workspace.SourceZipFile)

	LogInfo.Printf("Extracting zip file %s to %s", workspace.SourceZipFile, workspace.SourceDirectory)

	//now that the zip file is extracted, copy the docker file
	err = workspace.ExtractZipFile()

	if err != nil {
		Fail("Unable to extract workspace" + err.Error())
	}

	dockerInfo := &DockerInfo{
		RepoName:  repoName,
		ImageName: appName,
		Revision:  revision,
	}

	err = workspace.CreateDockerFile(dockerInfo)

	Expect(err).Should(BeNil(), "Could not find asset ", err)

	Expect(workspace.DockerFile).Should(BeAnExistingFile(), "Could not find docker file "+workspace.DockerFile)

	//now tar it up
	err = workspace.BuildTarFile()

	Expect(err).Should(BeNil(), "Unable to create tar file")

	return workspace, dockerInfo

}

func printImages(images *[]types.Image) {
	for _, img := range *images {
		fmt.Println("ID: ", img.ID)
		fmt.Println("RepoTags: ", img.RepoTags)
		fmt.Println("Created: ", img.Created)
		fmt.Println("Size: ", img.Size)
		fmt.Println("VirtualSize: ", img.VirtualSize)
		fmt.Println("ParentId: ", img.ParentID)
	}
}

func assertApplicationsExist(imageCreator ImageCreator, repoName string, expectedValues ...string) {

	apps, err := imageCreator.GetApplications(repoName)

	Expect(err).Should(BeNil(), "An error occured getting applications %s", err)

	for _, expected := range expectedValues {
		found := stringExists(apps, expected)

		Expect(found).Should(BeTrue(), "Could not find application %s for repo %s", expected, repoName)
	}
}

func assertApplicationVersionsExist(imageCreator ImageCreator, repoName string, applicationName string, versions ...string) {

	images, err := imageCreator.GetImages(repoName, applicationName)

	Expect(err).Should(BeNil(), "An Error occured getting images")

	for _, version := range versions {

		//loop through the images and verify them.

		found := false

		for _, image := range *images {

			for _, repoTag := range image.RepoTags {

				expected := fmt.Sprintf("%s/%s:%s", repoName, applicationName, version)

				if repoTag == expected {
					found = true
					break
				}
			}

			if found {
				break
			}

		}

		Expect(found).Should(BeTrue(), "Expected to find version %s for repo %s and applicationName %s", version, repoName, applicationName)

	}
}

func assertImageExists(imageCreator ImageCreator, expectedImage *DockerInfo) {
	image, err := imageCreator.GetImageRevision(expectedImage)

	Expect(err).Should(BeNil(), "Could not retrieve image.  %s", err)
	Expect(image).ShouldNot(BeNil(), "Image should be returned")

	found := false

	expected := expectedImage.GetTagName()

	for _, tag := range image.RepoTags {
		if tag == expected {
			found = true
			break
		}
	}

	Expect(found).Should(BeTrue(), "Image image should be %s", expectedImage.GetTagName())
}

//assertLocalImageExists search local images and ensures they exist
func assertLocalImageExists(imageCreator ImageCreator, dockerInfo *DockerInfo, imageSearch *DockerInfo) {

	result := searchLocalImages(imageCreator, dockerInfo, imageSearch)

	Expect(result).Should(BeTrue(), "Could not find image with the docker tags", imageSearch.GetTagName())

}

//returns true if the image exists, false otherwise
func searchLocalImages(imageCreator ImageCreator, dockerInfo *DockerInfo, imageSearch *DockerInfo) bool {

	images, err := imageCreator.GetLocalImages()

	Expect(err).Should(BeNil(), "Unable to list images", err)

	dockerTag := dockerInfo.GetTagName()

	return imageExists(images, dockerTag)

}

//imageExists.  Returns true if an image has been tagged with the specified repo name
func imageExists(images *[]types.Image, repoTagName string) bool {
	for _, img := range *images {

		for _, tag := range img.RepoTags {
			if strings.Contains(tag, repoTagName) {
				return true
			}
		}

	}

	return false
}

//stringExists return true if the st
func stringExists(array *[]string, search string) bool {

	for _, string := range *array {
		if search == string {
			return true
		}
	}

	return false
}

func channelToOutput(messages chan (string)) {

	for {
		message, ok := <-messages

		if !ok {
			break
		}

		fmt.Println(message)
	}
}
