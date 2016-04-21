package shipyard

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

//LocalPodSpec the ec2 pod spec
type LocalPodSpec struct {
	//the root directory for storage
	rootDir string
}

//NewLocalPodSpec create a new local pod spec to store pod specs in the given aws region and bucket
func NewLocalPodSpec(rootDirName string) (*LocalPodSpec, error) {
	//validate the root dir exists

	dir, err := filepath.Abs(rootDirName)

	if err != nil {
		return nil, err
	}

	LogInfo.Printf("Creating directory %s", dir)

	err = os.MkdirAll(dir, 0755)

	if err != nil {
		return nil, err
	}

	//return the local pod spec
	return &LocalPodSpec{
		rootDir: dir,
	}, nil
}

//WritePodSpec write the pod spec
func (localPodSpec *LocalPodSpec) WritePodSpec(namespace string, application string, revision string, podspec string) error {

	localPath := localPodSpec.createLocalPath(namespace, application, revision)

	LogInfo.Printf("Writing file to path %s", localPath)

	parentDir := filepath.Dir(localPath)

	err := os.MkdirAll(parentDir, 0755)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(localPath, []byte(podspec), 0755)

	if err != nil {
		return err
	}

	return nil

}

//ReadPodSpec read the pod spec and return it as a string
func (localPodSpec *LocalPodSpec) ReadPodSpec(namespace string, application string, revision string) (*string, error) {

	localPath := localPodSpec.createLocalPath(namespace, application, revision)

	bytes, err := ioutil.ReadFile(localPath)

	if err != nil {
		return nil, err
	}

	json := string(bytes)

	return &json, nil
}

//createPath create the path of the json
func (localPodSpec *LocalPodSpec) createLocalPath(namespace string, application string, revision string) string {
	return fmt.Sprintf("%s/%s/%s/%s.json", localPodSpec.rootDir, namespace, application, revision)
}
