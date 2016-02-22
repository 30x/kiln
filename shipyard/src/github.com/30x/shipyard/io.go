package shipyard

import ("os"
	"github.com/nu7hatch/gouuid"
	"archive/tar"
)

const DEFAULT_TMP_DIR = "/tmp"
const SHIPYARD_ENV_VARIABLE = "SHIPYARD_TMP_DIR"
const defaultFileMode = 0777

type SourceInfo struct {
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

	uuid, err := uuid.NewV4()


	if(err != nil ){
		return nil, err
	}


	sourceDirectory := tmpDir + "/" + uuid.String()
	targetTarName := sourceDirectory + ".tar"



	sourceInfo := &SourceInfo{
		sourceDirectory:sourceDirectory,
		targetTarName: targetTarName,
	}

	//create the directory
	mkdirError := os.MkdirAll(sourceInfo.sourceDirectory,defaultFileMode )

	if(mkdirError != nil){
		return nil, err
	}


	return sourceInfo, nil;

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
