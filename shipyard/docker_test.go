package shipyard

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"os"
	"strings"
	"testing"
"github.com/30x/shipyard/util"
)

func TestCreateTar(t *testing.T) {

	const remoteUrl = "http://localhost:5000"

	imageCreator, error := NewImageCreator(remoteUrl)

	if error != nil {
		t.Fatal("Could not create image", error)
	}

	const validTestZip = "../testresources/echo-test.zip"

	workspace := DoSetup(validTestZip, t)

	//clean up the workspace after the test.  Comment this out for debugging
	//defer workspace.Clean()

	dockerImage := &DockerInfo{
		TarFile:   workspace.TargetTarName,
		RepoName:  "test" + util.UUIDString(),
		ImageName: "test",
		Revision:   "v1.0",
	}

	//copy over our docker file.  These tests assume io has been tested and works properly

	imageCreator.BuildImage(dockerImage)

	//get the image from docker and ensure it exists

	images, err := imageCreator.ListImages()

	if err != nil {

		t.Fatal("Unable to list images", err)
	}

	printImages(&images)

	dockerTag := dockerImage.RepoName + "/" + dockerImage.ImageName + ":" + dockerImage.Revision

	if !imageExists(&images, dockerTag) {
		t.Fatal("Could not find image with the docker tags", dockerTag)
	}

	//
	//imageCreator.TagImage(dockerImage)
	//imageCreator.PushImage(dockerImage)

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

//doSetup Copies the specified inputZip file into the source directory and adds the docker file to it
func DoSetup(inputZip string, t *testing.T) *SourceInfo {

	//copy over our docker file.  These tests assume io has been tested and works properly

	const dockerAsset = "resources/Dockerfile"

	workspace, err := CreateNewWorkspace()

	if err != nil {
		t.Fatal("Should not return an error creating a valid workspace")
	}

	//change the source zip input for extactraction

	util.CopyFile(inputZip, workspace.SourceZipFile)

	Log.Printf("Extracting zip file %s to %s", workspace.SourceZipFile, workspace.SourceDirectory)

	//now that the zip file is extracted, copy the docker file
	err = workspace.ExtractZipFile()

	data, err := Asset(dockerAsset)

	if err != nil {
		t.Fatal("Could not find asset ", err)
	}

	ioutil.WriteFile(workspace.DockerFile, data, 770)

	if stat, err := os.Stat(workspace.DockerFile); err != nil || stat == nil {
		t.Fatal("Could not find docker file "+workspace.DockerFile, err)
	}

	//now tar it up
	workspace.BuildTarFile()

	return workspace

}
