package shipyard_test

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/30x/shipyard/pkg/shipyard"
)

var _ = Describe("Io", func() {

	It("Create workspace ", func() {
		workspace, err := CreateNewWorkspace()

		Expect(err).Should(BeNil(), "Should not return an error creating a valid workspace")

		//if could not find directory, it's a fail
		Expect(workspace.SourceDirectory).Should(BeADirectory(), "Could not find directory "+workspace.SourceDirectory)

		Expect(workspace.RootDirectory).Should(BeADirectory(), "Could not find directory "+workspace.RootDirectory)

		Expect(workspace.SourceZipFile).ShouldNot(BeEmpty(), "SourceZipFile should be specified")

		Expect(workspace.TargetTarName).ShouldNot(BeEmpty(), "TargetTarName should be specified")

		Expect(workspace.DockerFile).Should(ContainSubstring(workspace.SourceDirectory), "Docker file should be in the source directory")

		subString := strings.Replace(workspace.DockerFile, workspace.SourceDirectory, "", 1)

		Expect(subString).Should(Equal("/Dockerfile"), "Dockerfile was not in the correct location")

	})

	It("No Permissions", func() {
		os.Setenv(SHIPYARD_ENV_VARIABLE, "/usr/ishouldntbecreated")

		//unset variable after return
		defer os.Setenv(SHIPYARD_ENV_VARIABLE, DEFAULT_TMP_DIR)

		workspace, err := CreateNewWorkspace()

		Expect(workspace).Should(BeNil(), "Should have been able to create the directory")

		Expect(err).ShouldNot(BeNil(), "Should not have been able to create the directory")
	})

	It("Test unzip", func() {
		const validTestZip = "../../testresources/echo-test.zip"

		Expect(validTestZip).Should(BeAnExistingFile(), "Could not find source file ")

		workspace, err := CreateNewWorkspace()

		Expect(err).Should(BeNil(), "Should have been able to create the directory")

		Expect(workspace).ShouldNot(BeNil(), "Workspace should not be nil")

		//create a symlink to a valid test zip into our zip workspace
		err = CopyFile(validTestZip, workspace.SourceZipFile)

		Expect(err).Should(BeNil(), "Could not link test archive for verification of unzip")

		err = workspace.ExtractZipFile()

		Expect(err).Should(BeNil(), "Could not extract zip file ")

		//now validate the file
		log.Printf("Testing for source files in " + workspace.SourceDirectory)

		testFile := workspace.SourceDirectory + "/index.js"

		Expect(testFile).Should(BeAnExistingFile(), "Could not find source file ")

		testFile = workspace.SourceDirectory + "/package.json"

		Expect(testFile).Should(BeAnExistingFile(), "Could not find source file "+testFile)

	})

	It("Test Docker file", func() {
		sourceInfo, err := CreateNewWorkspace()

		Expect(err).Should(BeNil(), "Unable to create a workspace %s", err)

		dockerInfo := &DockerInfo{
			RepoName:  "testRepo",
			ImageName: "testImage",
			Revision:  "v1.0",
		}

		err = sourceInfo.CreateDockerFile(dockerInfo)

		Expect(err).Should(BeNil(), "Received an error creating template %s")

		//test they're the same

		bytes, err := ioutil.ReadFile(sourceInfo.DockerFile)

		Expect(err).Should(BeNil(), "Could not read file %s", err)

		expected :=
			`FROM mhart/alpine-node:4

      ADD . .
      RUN npm install
      #Taken from the runtime on start
      EXPOSE 9000

      LABEL com.github.30x.shipyard.repo=testRepo
      LABEL com.github.30x.shipyard.app=testImage
      LABEL com.github.30x.shipyard.revision=v1.0

      CMD ["npm", "start"]


      `

			//take out all the whitespace to ensure that the payloads match, don't test whitespace since this always fails
		expected = strings.Replace(expected, " ", "", -1)

		fileAsString := strings.Replace(string(bytes), " ", "", -1)

		Expect(fileAsString).Should(Equal(expected), "File is not as excepcted.  Received \n %s \n but expected \n %s \n ", fileAsString, expected)
	})

})
