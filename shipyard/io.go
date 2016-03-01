package shipyard

import (
	"os"
	"archive/tar"
	"path/filepath"
	"strings"
	"io"
	"errors"
	"archive/zip"
	uuid "github.com/nu7hatch/gouuid"
)

const DEFAULT_TMP_DIR = "/tmp"
const SHIPYARD_ENV_VARIABLE = "SHIPYARD_TMP_DIR"
const DEFAULT_FILE_MODE = 0775

type SourceInfo struct {
	//the root directory containing the zip, source and target tar
	rootDirectory   string
	//The name of the zip file to create and stream data to
	sourceZipFile   string
	//The directory of extracted source, in our case, NODE.JS code
	sourceDirectory string
	//the targer tar file name
	targetTarName   string
}


//CreateNewWorkspace Creates a new tmp directory and return a source directory struct
func CreateNewWorkspace() (*SourceInfo, error) {

	tmpDir, exists := os.LookupEnv(SHIPYARD_ENV_VARIABLE)

	if (!exists) {
		tmpDir = DEFAULT_TMP_DIR
	}

	uuidBinary, err := uuid.NewV4()


	if (err != nil ) {
		return nil, err
	}

	uuid := uuidBinary.String()

	rootDirectory := tmpDir + "/shipyard/" + uuid
	sourceDirectory := rootDirectory + "/source"
	zipFileName := rootDirectory + "/input.zip"
	targetTarName := rootDirectory + "/docker-data.tar"

	sourceInfo := &SourceInfo{
		rootDirectory: rootDirectory,
		sourceZipFile: zipFileName,
		sourceDirectory:sourceDirectory,
		targetTarName: targetTarName,
	}

	//create the directory


	//implicity creates the root directory
	mkdirError := os.MkdirAll(sourceInfo.sourceDirectory, DEFAULT_FILE_MODE)

	if (mkdirError != nil) {
		return nil, err
	}

	return sourceInfo, nil;

}

//ExtractZipFile Extracts a zip file.  Returns an error if the system is unable to extract a zip
func (sourceInfo *SourceInfo)  ExtractZipFile() error {
	reader, err := zip.OpenReader(sourceInfo.sourceZipFile)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(sourceInfo.sourceDirectory, DEFAULT_FILE_MODE); err != nil {
			return err
		}

		for _, file := range reader.File {
			path := filepath.Join(sourceInfo.sourceDirectory, file.Name)
			if file.FileInfo().IsDir() {
				err := os.MkdirAll(path, file.Mode())
				if(err != nil){
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

//BuildTarFile.  Copies the docker file into the current working directory, tars the source and returns
func (sourceInfo *SourceInfo) BuildTarFile() error {

	target := filepath.Join(sourceInfo.targetTarName)
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(sourceInfo.sourceDirectory)
	if err != nil {
		return err
	}

	var baseDir string
	if !info.IsDir() {
		return errors.New("You cannot tar a file, you must tar a directory")
	}

	baseDir = filepath.Base(sourceInfo.sourceDirectory)

	//
	return filepath.Walk(sourceInfo.sourceDirectory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, sourceInfo.sourceDirectory))
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})

}


//Clean removes the tar file and temporary work directory.  Returns an error if an error occurred, nil otherwise.
//Note that you will probably want to invoke via defer sourceInfo.Clean() to ensure cleanup
func (sourceInfo *SourceInfo) Clean() error {

	dirError := os.RemoveAll(sourceInfo.rootDirectory)

	if (dirError != nil) {
		return dirError
	}

	return nil
}
