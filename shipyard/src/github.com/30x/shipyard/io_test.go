package shipyard

import (
	"testing"
	"os"
)

//TestCreateWorkspace Tests creating the temporary working directory
func TestCreateWorkspace(t *testing.T) {
	workspace, err := CreateNewWorkspace()

	if(err != nil){
		t.Fatal("Should not return an error creating a valid workspace")
	}

	//if could not find directory, it's a fail
	if _, err := os.Stat(workspace.sourceDirectory); os.IsNotExist(err) {
		t.Fatal("Could not find directory " + workspace.sourceDirectory)
	}

	//otherwise success
}

//TestNoPermissions Tests an error is correctly thrown when the system cant' create the directory
func TestNoPermissions(t *testing.T) {

	os.Setenv(SHIPYARD_ENV_VARIABLE, "/usr/ishouldntbecreated")

	workspace, err := CreateNewWorkspace()

	if(workspace !=nil && err == nil){
		t.Fatal("Should not have been able to create the directory")
	}

	//unset variable
	os.Setenv(SHIPYARD_ENV_VARIABLE, "")

}
