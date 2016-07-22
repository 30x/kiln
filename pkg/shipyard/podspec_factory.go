package shipyard

import (
	"errors"
	"os"
)

//NewPodSpecIoFromEnv instanciate the pod spect storage from the environment variables.  If an env is missing, an error will be logged and execution will halt
func NewPodSpecIoFromEnv() (PodspecIo, error) {

	creatorValue := os.Getenv("POD_PROVIDER")

	if creatorValue == "" {
		return nil, errors.New("You most specifity the POD_PROVIDER environment variable.  Valid values are 'local' or 's3'")
	}

	switch creatorValue {
	case "local":
		return newLocalPodSpecFromEnv()
	default:
		return nil, errors.New("You most specifity the POD_PROVIDER environment variable.  Valid values are 'local' or 's3'")
	}
}

func newLocalPodSpecFromEnv() (*LocalPodSpec, error) {

	localDirectory := os.Getenv("LOCAL_DIR")

	if localDirectory == "" {
		LogError.Fatalf("You must specify the LOCAL_DIR variable.")
	}

	return NewLocalPodSpec(localDirectory)
}
