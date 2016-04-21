package shipyard

import (
	"errors"
	"os"
)

//NewPodSpecIoFromEnv instanciate the pod spect storage from the environment variables.  If an env is missing, an error will be logged and execution will halt
func NewPodSpecIoFromEnv() (PodspecIo, error) {

	creatorValue := os.Getenv("POD_PROVIDER")

	if creatorValue == "" {
		return nil, errors.New("You most specifity the POD_PROVIDER environment variable.  Valid values are 'docker' or 'ecr'")
	}

	switch creatorValue {
	case "local":
		return newLocalPodSpecFromEnv()
	case "s3":
		return newS3PodSpecFromEnv()
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

func newS3PodSpecFromEnv() (*S3PodSpec, error) {
	s3Region := os.Getenv("EC2_REGION")

	if s3Region == "" {
		LogError.Fatalf("You must specify the EC2_REGION variable. An example value is 'us-east-1'")
	}

	s3BucketName := os.Getenv("EC2_BUCKET")

	if s3BucketName == "" {
		LogError.Fatalf("You must specify the EC2_BUCKET variable. An example value is 'testbeeswaxbucket'")
	}

	return NewS3PodSpec(s3Region, s3BucketName)
}
