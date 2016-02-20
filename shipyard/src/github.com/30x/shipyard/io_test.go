package shipyard

import (
	"testing"
	"os"
)

func TestAuthLegacyConfig(t *testing.T) {
	workspace := CreateNewWorkspace()

	//if could not find directory, it's a fail
	if _, err := os.Stat(workspace.sourceDirectory); os.IsNotExist(err) {
		t.Fatal("Could not find directory " + workspace.sourceDirectory)
	}

	//otherwise success
}
