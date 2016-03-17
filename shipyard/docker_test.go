package shipyard_test

import (
	. "github.com/30x/shipyard/shipyard"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"github.com/docker/engine-api/types"
	"os"
	"strings"
)

var _ = Describe("docker", func() {

	var imageCreator ImageCreator

	//Tests for all image operations
	Describe("Image Operations", func() {

		AssertImageTests := func() {
			// It("Should create image successfully", func() {
			// 	createImage(imageCreator)
			// })

			It("Tag and Push", func() {
				_, dockerInfo := createImage(imageCreator)

				err := imageCreator.PushImage(dockerInfo, os.Stdout)

				Expect(err).Should(BeNil(), "Unable to push image", err)
				imageSearch := &DockerInfo{
					RepoName:  dockerInfo.RepoName,
					ImageName: dockerInfo.ImageName,
				}

				images, err := imageCreator.SearchRemoteImages(imageSearch)

				Expect(err).Should(BeNil(), "Unable to list images", err)

				printImages(&images)

				dockerTag := dockerInfo.GetTagName()

				Expect(imageExists(&images, dockerTag)).Should(Equal(true), "Could not find image with the docker tags", dockerTag)

				err = imageCreator.PullImage(dockerInfo, os.Stdout)

				Expect(err).Should(BeNil(), "Could not pull image from remote repo, upload may have failed", err)
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

		// Context("Amazon ECS", func() {
		// 	BeforeEach(func() {
		// 		var error error
		// 		imageCreator, error = NewEcsImageCreator("977777657611.dkr.ecr.us-east-1.amazonaws.com", "us-east-1")

		// 		Expect(error).Should(BeNil(), "Could not create local docker image creator")

		// 	})

		// 	AssertImageTests()
		// })

	})

})

//helper functions called within the tests
func createImage(imageCreator ImageCreator) (*SourceInfo, *DockerInfo) {

	const validTestZip = "../testresources/echo-test.zip"

	workspace, dockerInfo := doSetup(validTestZip)

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

	images, err := imageCreator.SearchLocalImages(&DockerInfo{})

	Expect(err).Should(BeNil(), "Unable to list images", err)

	printImages(&images)

	dockerTag := dockerImage.GetTagName()

	Expect(imageExists(&images, dockerTag)).Should(Equal(true), "Could not find image with the docker tags", dockerTag)

	//pull by label

	search := &DockerInfo{
		RepoName:  dockerInfo.RepoName,
		ImageName: dockerInfo.ImageName,
		Revision:  dockerInfo.Revision,
	}

	images, err = imageCreator.SearchLocalImages(search)

	Expect(err).Should(BeNil(), "Unable to list images", err)

	printImages(&images)

	Expect(imageExists(&images, dockerTag)).Should(Equal(true), "Could not find image with the docker tags", dockerTag)

	return workspace, dockerImage.DockerInfo

}

//DoSetup Copies the specified inputZip file into the source directory and adds the docker file to it
func doSetup(inputZip string) (*SourceInfo, *DockerInfo) {

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

	dockerFile := &DockerFile{
		ParentImage: "node:4.3.0-onbuild",
		DockerInfo: DockerInfo{
			RepoName:  "test" + UUIDString(),
			ImageName: "test",
			Revision:  "v1.0",
		},
	}

	err = workspace.CreateDockerFile(dockerFile)

	Expect(err).Should(BeNil(), "Could not find asset ", err)

	if stat, err := os.Stat(workspace.DockerFile); err != nil || stat == nil {
		Fail("Could not find docker file " + workspace.DockerFile + " " + err.Error())
	}

	//now tar it up
	err = workspace.BuildTarFile()

	Expect(err).Should(BeNil(), "Unable to create tar file")

	return workspace, &dockerFile.DockerInfo

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
