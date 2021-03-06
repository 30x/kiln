package kiln

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"fmt"

	uuid "github.com/nu7hatch/gouuid"
)

const DEFAULT_TMP_DIR = "/tmp"
const SHIPYARD_ENV_VARIABLE = "SHIPYARD_TMP_DIR"
const DEFAULT_FILE_MODE = 0775

const TAG_REPO = "com.github.30x.kiln.repo"
const TAG_APPLICATION = "com.github.30x.kiln.app"
const TAG_REVISION = "com.github.30x.kiln.revision"

const DefaultNodeBaseImage = "mhart/alpine-node:4"
const NodeImageRepo = "mhart/alpine-node"

//SourceInfo The description of the source package
type SourceInfo struct {
	//the root directory containing the zip, source and target tar
	RootDirectory string
	//The name of the zip file to create and stream data to
	SourceZipFile string
	//The directory of extracted source, in our case, NODE.JS code
	SourceDirectory string
	//The directory of the docker file
	DockerFile string
	//the targer tar file name
	TargetTarName string
}

//CreateNewWorkspace Creates a new tmp directory and return a source directory struct
func CreateNewWorkspace() (*SourceInfo, error) {

	tmpDir, exists := os.LookupEnv(SHIPYARD_ENV_VARIABLE)

	if !exists {
		tmpDir = DEFAULT_TMP_DIR
	}

	uuidBinary, err := uuid.NewV4()

	if err != nil {
		return nil, err
	}

	uuid := uuidBinary.String()

	rootDirectory := tmpDir + "/kiln/" + uuid
	sourceDirectory := rootDirectory + "/source"
	dockerFile := sourceDirectory + "/Dockerfile"
	zipFileName := rootDirectory + "/input.zip"
	targetTarName := rootDirectory + "/docker-data.tar"

	sourceInfo := &SourceInfo{
		RootDirectory:   rootDirectory,
		SourceZipFile:   zipFileName,
		SourceDirectory: sourceDirectory,
		DockerFile:      dockerFile,
		TargetTarName:   targetTarName,
	}

	//create the directory

	//implicitly creates the root directory
	mkdirError := os.MkdirAll(sourceInfo.SourceDirectory, DEFAULT_FILE_MODE)

	if mkdirError != nil {
		return nil, mkdirError
	}

	return sourceInfo, nil

}

//ExtractZipFile Extracts a zip file.  Returns an error if the system is unable to extract a zip
func (sourceInfo *SourceInfo) ExtractZipFile() error {
	reader, err := zip.OpenReader(sourceInfo.SourceZipFile)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(sourceInfo.SourceDirectory, DEFAULT_FILE_MODE); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(sourceInfo.SourceDirectory, file.Name)
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(path, file.Mode())
			if err != nil {
				return err
			}

			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		//NOTE defer has intentionally been omitted here.  It would cause close calls to stack up
		//when reading large resources
		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			//can't open the file for writing, close the read handle on the existing file
			fileReader.Close()
			return err
		}

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			//failed to copy, close both, ignore errors on close
			fileReader.Close()
			targetFile.Close()
			return err
		}

		//close our I/O
		fileReader.Close()
		targetFile.Close()
	}

	return nil
}

//WriteZipFileData write the zip file data from a reader
func (sourceInfo *SourceInfo) WriteZipFileData(reader io.Reader) error {
	file, err := os.Create(sourceInfo.SourceZipFile)

	if err != nil {
		LogError.Printf("Unable to write bytes to file at %s.  Error is %s", sourceInfo.SourceZipFile, err)
		return err
	}

	defer file.Close()

	count, err := io.Copy(file, reader)

	if err != nil {
		LogError.Printf("Unable to write bytes to file at %s.  Error is %s", sourceInfo.SourceZipFile, err)
		return err
	}

	LogInfo.Printf("Wrote %d bytes to file %s", count, sourceInfo.SourceZipFile)

	return nil

}

//BuildTarFile Copies the docker file into the current working directory, tars the source and returns
func (sourceInfo *SourceInfo) BuildTarFile() error {

	target := filepath.Join(sourceInfo.TargetTarName)
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(sourceInfo.SourceDirectory)
	if err != nil {
		return err
	}

	var baseDir string
	if !info.IsDir() {
		return errors.New("You cannot tar a file, you must tar a directory")
	}

	//
	return filepath.Walk(sourceInfo.SourceDirectory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, sourceInfo.SourceDirectory))

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}

			_, err = io.Copy(tarball, file)
			file.Close()

			return err
		})

}

//Clean removes the tar file and temporary work directory.  Returns an error if an error occurred, nil otherwise.
//Note that you will probably want to invoke via defer sourceInfo.Clean() to ensure cleanup
func (sourceInfo *SourceInfo) Clean() error {

	dirError := os.RemoveAll(sourceInfo.RootDirectory)

	if dirError != nil {
		return dirError
	}

	return nil
}

//The TEMPLATE for generating a go file
const templateString = `FROM {{.BaseImage}}

ADD . .
RUN apk add --no-cache git && \
    npm install && \
    apk del git

LABEL ` + TAG_REPO + `={{.RepoName}}
LABEL ` + TAG_APPLICATION + `={{.ImageName}}
LABEL ` + TAG_REVISION + `={{.Revision}}

CMD ["npm", "start"]

{{range .EnvVars}}
ENV {{.}}
{{end}}
`

//constant that's initialized below.  Constants must only be primitive types
var dockerTemplate *template.Template

//init Initializes the docker template once for performance
func init() {
	dockerTemplate = template.Must(template.New("Dockerfile").Parse(templateString))
}

//CreateDockerFile creates a dockerfile at the specified location from the specfiied dockerInfo
func (sourceInfo *SourceInfo) CreateDockerFile(dockerInfo *DockerInfo) error {

	parentPath := filepath.Dir(sourceInfo.DockerFile)

	//create the parent path
	err := os.MkdirAll(parentPath, DEFAULT_FILE_MODE)

	if err != nil {
		return err
	}

	//create the docker file and close after we exit
	dockerFile, err := os.Create(sourceInfo.DockerFile)

	defer dockerFile.Close()

	if err != nil {
		return err
	}

	err = dockerTemplate.Execute(dockerFile, dockerInfo)

	if err != nil {
		return err
	}

	return nil
}

// GetExampleDockerfile is used to create an example Dockerfile with the given information
func GetExampleDockerfile(dockerInfo *DockerInfo) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}

	err = dockerTemplate.Execute(buf, dockerInfo)

	if err != nil {
		return nil, err
	}

	return buf, nil
}

// DetermineBaseImage consumes a given runtime selection and determines the base image for the Dockerfile
func DetermineBaseImage(runtimeSelection string) (baseImage string, err error) {
	runtimeSplit := strings.Split(runtimeSelection, ":")

	switch runtimeSplit[0] {
	case "node":
		if len(runtimeSplit) > 1 {
			baseImage = fmt.Sprintf("%s:%s", NodeImageRepo, runtimeSplit[1])
		} else {
			baseImage = DefaultNodeBaseImage
		}
		break
	default:
		return "", fmt.Errorf("Unsupported runtime selection: %s", runtimeSelection)
	}

	return
}
