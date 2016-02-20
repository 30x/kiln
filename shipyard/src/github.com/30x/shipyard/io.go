package shipyard

import ("os"
"github.com/nu7hatch/gouuid"
)

const DEFAULT_TMP_DIR = "/tmp"
const defaultFileMode = 0777

type SourceInfo struct {
	//The directory of extracted source, in our case, NODE.JS code
	sourceDirectory string
	//the targer tar file name
	targetTarName   string
}


//CreateNewWorkspace Creates a new tmp directory and return a source directory struct
func CreateNewWorkspace() (*SourceInfo) {
	tmpDir, exists := os.LookupEnv("TMP_DIR")

	if (!exists) {
		tmpDir = DEFAULT_TMP_DIR
	}

	uuid, err := uuid.NewV4()


	if(err != nil ){
		//TODO, something useful here
	}


	sourceDirectory := tmpDir + "/" + uuid.String()
	targetTarName := sourceDirectory + ".tar"



	sourceInfo := &SourceInfo{
		sourceDirectory:sourceDirectory,
		targetTarName: targetTarName,
	}

	//create the directory
	os.MkdirAll(sourceInfo.sourceDirectory,defaultFileMode )


	return sourceInfo;

}

//BuildTarFile.  Copies the docker file into the current working directory, tars the source and returns
func (sourceInfo *SourceInfo) BuildTarFile(){

}


//Clean removes the tar file and temporary work directory.  Returns an error if an error occured, nil otherwise
func (sourceInfo *SourceInfo) Clean() error{

	dirError:= os.RemoveAll(sourceInfo.sourceDirectory)

	if(dirError != nil){
		return dirError
	}

	fileError := os.Remove(sourceInfo.targetTarName)

	if(fileError != nil){
		return fileError
	}

	return nil
}
