package shipyard

import (
	"testing"
	"os"
	"io"
)

//TestCreateWorkspace Tests creating the temporary working directory
func TestCreateWorkspace(t *testing.T) {
	workspace, err := CreateNewWorkspace()

	if (err != nil) {
		t.Fatal("Should not return an error creating a valid workspace")
	}

	//if could not find directory, it's a fail
	if _, err := os.Stat(workspace.sourceDirectory); os.IsNotExist(err) {
		t.Fatal("Could not find directory " + workspace.sourceDirectory)
	}

	if _, err := os.Stat(workspace.rootDirectory); os.IsNotExist(err) {
		t.Fatal("Could not find directory " + workspace.rootDirectory)
	}

	if (workspace.sourceZipFile == "") {
		t.Fatal("sourceZipFile should be specified")
	}

	if (workspace.targetTarName == "") {
		t.Fatal("targetTarName should be specified")
	}


	//otherwise success
}

//TestNoPermissions Tests an error is correctly thrown when the system cant' create the directory
func TestNoPermissions(t *testing.T) {

	os.Setenv(SHIPYARD_ENV_VARIABLE, "/usr/ishouldntbecreated")

	//unset variable
	defer os.Setenv(SHIPYARD_ENV_VARIABLE, DEFAULT_TMP_DIR)

	workspace, err := CreateNewWorkspace()

	if (workspace != nil && err == nil) {
		t.Fatal("Should not have been able to create the directory")
	}



}



//TestNoPermissions Tests an error is correctly thrown when the system cant' create the directory
func TestUnzip(t *testing.T) {

	const validTestZip = "testresources/echo-test.zip"

	if _, err := os.Stat(validTestZip); os.IsNotExist(err) {
		t.Fatal("Could not find source file " + validTestZip)
	}

	workspace, err := CreateNewWorkspace()

	if (err != nil) {
		t.Fatal("Should not have been able to create the directory")
	}

	if( workspace == nil){
		t.Fatal("Workspace should not be nil")
	}



	//create a symlink to a valid test zip into our zip workspace
	err = copyFile(validTestZip, workspace.sourceZipFile)

	if (err != nil) {
		t.Fatal("Could not link test archive for verification of unzip", err)
	}

	err = workspace.ExtractZipFile()

	if ( err != nil) {
		t.Fatal("Could not extract zip file ", err)
	}

	//now validate the file

	testFile := workspace.sourceDirectory + "/index.js"


	if _, err := os.Stat(testFile); err != nil {
		t.Fatal("Could not find source file " + testFile, err)
	}

	testFile = workspace.sourceDirectory + "/package.json"

	if _, err := os.Stat(testFile); err != nil {
		t.Fatal("Could not find source file " + testFile, err)
	}

}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	//defer closing
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	//defer closing
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return err
}
