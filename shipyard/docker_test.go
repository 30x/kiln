package shipyard

import (
	// "fmt"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"io"
	"os"
	"strings"
	"testing"
)

const remoteURL = "localhost:5000"

//tests creating an image, then assets it is present in docker
func TestCreateTar(t *testing.T) {

	createImage(t)

}

//tests pushing to a remote repo, and ensuring we can pull the image
func TestPushImage(t *testing.T) {

	_, dockerInfo, imageCreator := createImage(t)

	stream, err := imageCreator.PushImage(dockerInfo)

	if err != nil {
		t.Fatal("Unable to push image", err)
	}

	io.Copy(os.Stdout, stream)

	images, err := imageCreator.ListImages()

	if err != nil {
		t.Fatal("Unable to list images", err)
	}

	printImages(&images)

	dockerTag := remoteURL + "/" + dockerInfo.getTagName()

	if !imageExists(&images, dockerTag) {
		t.Fatal("Could not find image with the docker tags", dockerTag)
	}

	err = imageCreator.PullImage(dockerInfo)

	if err != nil {
		t.Fatal("Could not pull image from remote repo, upload may have failed", err)
	}

}

//createImage creates an image and validates it exists in docker.  Assumes you have a docker registry API running at localhost:5000
func createImage(t *testing.T) (*SourceInfo, *DockerInfo, *ImageCreator) {

	imageCreator, error := NewImageCreator(remoteURL)

	if error != nil {
		t.Fatal("Could not create image", error)
	}

	const validTestZip = "../testresources/echo-test.zip"

	workspace, dockerInfo := DoSetup(validTestZip, t)

	//clean up the workspace after the test.  Comment this out for debugging
	//defer workspace.Clean()

	dockerImage := &DockerBuild{
		TarFile:    workspace.TargetTarName,
		DockerInfo: dockerInfo,
	}

	//copy over our docker file.  These tests assume io has been tested and works properly

	stream, err := imageCreator.BuildImage(dockerImage)

	if err != nil {
		t.Fatal("Unable to build image", err)
	}

	io.Copy(os.Stdout, stream)

	//get the image from docker and ensure it exists

	images, err := imageCreator.ListImages()

	if err != nil {

		t.Fatal("Unable to list images", err)
	}

	printImages(&images)

	dockerTag := dockerImage.getTagName()

	if !imageExists(&images, dockerTag) {
		t.Fatal("Could not find image with the docker tags", dockerTag)
	}

	//pull by label

    
    search := &ImageSearch{
        Repository: dockerInfo.RepoName,
        Application: dockerInfo.ImageName,
        Revision: dockerInfo.Revision,
    }
    
	images, err = imageCreator.SearchImages(search)

	if err != nil {

		t.Fatal("Unable to list images", err)
	}

	printImages(&images)

	if !imageExists(&images, dockerTag) {
		t.Fatal("Could not find image with the docker tags", dockerTag)
	}

	return workspace, dockerImage.DockerInfo, imageCreator

}

func printImages(images *[]docker.APIImages) {
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
func imageExists(images *[]docker.APIImages, repoTagName string) bool {
	for _, img := range *images {

		for _, tag := range img.RepoTags {
			if strings.Contains(tag, repoTagName) {
				return true
			}
		}

	}

	return false
}

//DoSetup Copies the specified inputZip file into the source directory and adds the docker file to it
func DoSetup(inputZip string, t *testing.T) (*SourceInfo, *DockerInfo) {

	//copy over our docker file.  These tests assume io has been tested and works properly

	const dockerAsset = "resources/Dockerfile"

	workspace, err := CreateNewWorkspace()

	if err != nil {
		t.Fatal("Should not return an error creating a valid workspace")
	}

	//change the source zip input for extactraction
	CopyFile(inputZip, workspace.SourceZipFile)

	Log.Printf("Extracting zip file %s to %s", workspace.SourceZipFile, workspace.SourceDirectory)

	//now that the zip file is extracted, copy the docker file
	err = workspace.ExtractZipFile()

	dockerInfo := &DockerInfo{
		RepoName:  "test" + UUIDString(),
		ImageName: "test",
		Revision:  "v1.0",
	}

	err = workspace.CreateDockerFile(dockerInfo)

	if err != nil {
		t.Fatal("Could not find asset ", err)
	}

	if stat, err := os.Stat(workspace.DockerFile); err != nil || stat == nil {
		t.Fatal("Could not find docker file "+workspace.DockerFile, err)
	}

	//now tar it up
	err = workspace.BuildTarFile()

	if err != nil {
		t.Fatal("Unable to create tar file")
	}

	return workspace, dockerInfo

}
