package shipyard

import (
	"log"
	"os"
	"strings"
	"testing"
)

//TestCreateWorkspace Tests creating the temporary working directory
func TestCreateWorkspace(t *testing.T) {
	workspace, err := CreateNewWorkspace()

	if err != nil {
		t.Fatal("Should not return an error creating a valid workspace")
	}

	//if could not find directory, it's a fail
	if _, err := os.Stat(workspace.SourceDirectory); os.IsNotExist(err) {
		t.Fatal("Could not find directory " + workspace.SourceDirectory)
	}

	if _, err := os.Stat(workspace.RootDirectory); os.IsNotExist(err) {
		t.Fatal("Could not find directory " + workspace.RootDirectory)
	}

	if workspace.SourceZipFile == "" {
		t.Fatal("sourceZipFile should be specified")
	}

	if workspace.TargetTarName == "" {
		t.Fatal("targetTarName should be specified")
	}

	if !strings.Contains(workspace.DockerFile, workspace.SourceDirectory) {
		t.Fatal("Docker file should be in the source directory")
	}

	subString := strings.Replace(workspace.DockerFile, workspace.SourceDirectory, "", 1)

	if subString != "/Dockerfile" {
		t.Fatal("Dockerfile was not in the correct location")
	}

	//otherwise success
}

//TestNoPermissions Tests an error is correctly thrown when the system cant' create the directory
func TestNoPermissions(t *testing.T) {

	os.Setenv(SHIPYARD_ENV_VARIABLE, "/usr/ishouldntbecreated")

	//unset variable
	defer os.Setenv(SHIPYARD_ENV_VARIABLE, DEFAULT_TMP_DIR)

	workspace, err := CreateNewWorkspace()

	if workspace != nil && err == nil {
		t.Fatal("Should not have been able to create the directory")
	}

}

//TestNoPermissions Tests an error is correctly thrown when the system cant' create the directory
func TestUnzip(t *testing.T) {

	const validTestZip = "../testresources/echo-test.zip"

	if _, err := os.Stat(validTestZip); os.IsNotExist(err) {
		t.Fatal("Could not find source file " + validTestZip)
	}

	workspace, err := CreateNewWorkspace()

	if err != nil {
		t.Fatal("Should not have been able to create the directory")
	}

	if workspace == nil {
		t.Fatal("Workspace should not be nil")
	}

	//create a symlink to a valid test zip into our zip workspace
	err = CopyFile(validTestZip, workspace.SourceZipFile)

	if err != nil {
		t.Fatal("Could not link test archive for verification of unzip", err)
	}

	err = workspace.ExtractZipFile()

	if err != nil {
		t.Fatal("Could not extract zip file ", err)
	}

	//now validate the file
	log.Printf("Testing for source files in " + workspace.SourceDirectory)

	testFile := workspace.SourceDirectory + "/index.js"

	if stat, err := os.Stat(testFile); err != nil || stat == nil {
		t.Fatal("Could not find source file " + testFile, err)
	}

	testFile = workspace.SourceDirectory + "/package.json"

	if stat, err := os.Stat(testFile); err != nil || stat == nil {
		t.Fatal("Could not find source file " + testFile, err)
	}

}

