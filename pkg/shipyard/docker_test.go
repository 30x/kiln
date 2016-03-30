package shipyard_test

import (
	. "github.com/30x/shipyard/pkg/shipyard"
	"github.com/docker/engine-api/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"os"
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

				err := imageCreator.PushImage(dockerInfo, os.Stdout)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				imageSearch := &DockerInfo{
					RepoName:  dockerInfo.RepoName,
					ImageName: dockerInfo.ImageName,
				}

				assertRemoteImageExists(imageCreator, dockerInfo, imageSearch)

			})

			It("Test Search", func() {

				//push first image
				repoName := "test" + UUIDString()
				imageName1 := "test1"
				revision10 := "v1.0"

				_, dockerInfo10 := createImage(imageCreator, repoName, imageName1, revision10)

				err := imageCreator.PushImage(dockerInfo10, os.Stdout)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				revision11 := "v1.1"

				_, dockerInfo11 := createImage(imageCreator, repoName, imageName1, revision11)

				err = imageCreator.PushImage(dockerInfo11, os.Stdout)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				//push second image
				imageName2 := "test2"

				_, dockerInfo2 := createImage(imageCreator, repoName, imageName2, revision10)

				err = imageCreator.PushImage(dockerInfo2, os.Stdout)

				Expect(err).Should(BeNil(), "Unable to push image", err)

				//perform a search on all 3 images since they have the same repo name and ensure they're returned
				imageSearch := &DockerInfo{
					RepoName: dockerInfo2.RepoName,
				}

				assertLocalImageExists(imageCreator, dockerInfo10, imageSearch)
				assertLocalImageExists(imageCreator, dockerInfo11, imageSearch)
				assertLocalImageExists(imageCreator, dockerInfo2, imageSearch)

				assertRemoteImageExists(imageCreator, dockerInfo10, imageSearch)
				assertRemoteImageExists(imageCreator, dockerInfo11, imageSearch)
				assertRemoteImageExists(imageCreator, dockerInfo2, imageSearch)

				//refine to repo and app

				imageSearch = &DockerInfo{
					RepoName:  dockerInfo10.RepoName,
					ImageName: dockerInfo10.ImageName,
				}

				assertLocalImageExists(imageCreator, dockerInfo10, imageSearch)
				assertLocalImageExists(imageCreator, dockerInfo11, imageSearch)
				assertNoLocalImageExists(imageCreator, dockerInfo2, imageSearch)

				assertRemoteImageExists(imageCreator, dockerInfo10, imageSearch)
				assertRemoteImageExists(imageCreator, dockerInfo11, imageSearch)
				assertNoRemoteImageExists(imageCreator, dockerInfo2, imageSearch)

				//now search for specific revision
				imageSearch = &DockerInfo{
					RepoName:  dockerInfo11.RepoName,
					ImageName: dockerInfo11.ImageName,
					Revision:  dockerInfo11.Revision,
				}

				assertNoLocalImageExists(imageCreator, dockerInfo10, imageSearch)
				assertLocalImageExists(imageCreator, dockerInfo11, imageSearch)
				assertNoLocalImageExists(imageCreator, dockerInfo2, imageSearch)

				assertNoRemoteImageExists(imageCreator, dockerInfo10, imageSearch)
				assertRemoteImageExists(imageCreator, dockerInfo11, imageSearch)
				assertNoRemoteImageExists(imageCreator, dockerInfo2, imageSearch)

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

	err := imageCreator.BuildImage(dockerImage, os.Stdout)

	Expect(err).Should(BeNil(), "Unable to build image", err)

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

	dockerFile := &DockerFile{
		ParentImage: "node:4.3.0-onbuild",
		DockerInfo: &DockerInfo{
			RepoName:  repoName,
			ImageName: appName,
			Revision:  revision,
		},
	}

	err = workspace.CreateDockerFile(dockerFile)

	Expect(err).Should(BeNil(), "Could not find asset ", err)

	Expect(workspace.DockerFile).Should(BeAnExistingFile(), "Could not find docker file "+workspace.DockerFile)

	//now tar it up
	err = workspace.BuildTarFile()

	Expect(err).Should(BeNil(), "Unable to create tar file")

	return workspace, dockerFile.DockerInfo

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

//assertLocalImageExists search local images and ensures they exist
func assertLocalImageExists(imageCreator ImageCreator, dockerInfo *DockerInfo, imageSearch *DockerInfo) {

	result := searchLocalImages(imageCreator, dockerInfo, imageSearch)

	Expect(result).Should(BeTrue(), "Could not find image with the docker tags", imageSearch.GetTagName())

}

func assertNoLocalImageExists(imageCreator ImageCreator, dockerInfo *DockerInfo, imageSearch *DockerInfo) {
	result := searchLocalImages(imageCreator, dockerInfo, imageSearch)

	Expect(result).Should(BeFalse(), "Shouldn't  find image with the docker tags", imageSearch.GetTagName())
}

//returns true if the image exists, false otherwise
func searchLocalImages(imageCreator ImageCreator, dockerInfo *DockerInfo, imageSearch *DockerInfo) bool {

	images, err := imageCreator.SearchLocalImages(imageSearch)

	Expect(err).Should(BeNil(), "Unable to list images", err)

	dockerTag := dockerInfo.GetTagName()

	return imageExists(images, dockerTag)

}

//assertRemoteImageExists Searches remote images and ensures they exist
func assertRemoteImageExists(imageCreator ImageCreator, dockerInfo *DockerInfo, imageSearch *DockerInfo) {

	result := searchRemoteImageExists(imageCreator, dockerInfo, imageSearch)

	Expect(result).Should(BeTrue(), "Could not find image with the docker tags", imageSearch.GetTagName())

}

func assertNoRemoteImageExists(imageCreator ImageCreator, dockerInfo *DockerInfo, imageSearch *DockerInfo) {

	result := searchRemoteImageExists(imageCreator, dockerInfo, imageSearch)

	Expect(result).Should(BeFalse(), "Should not find image with the docker tags", imageSearch.GetTagName())

}

func searchRemoteImageExists(imageCreator ImageCreator, dockerInfo *DockerInfo, imageSearch *DockerInfo) bool {

	images, err := imageCreator.SearchRemoteImages(imageSearch)

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
