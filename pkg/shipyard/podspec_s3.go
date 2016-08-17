package shipyard

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

//S3PodSpec the ec2 pod spec
type S3PodSpec struct {
	//pointer to the s3 instance
	s3 *s3.S3

	//the stream of the bucket
	bucket string
}

//NewS3PodSpec create a new EC2 pod spec to store pod specs in the given aws region and bucket
func NewS3PodSpec(awsRegion string, bucketName string) (*S3PodSpec, error) {
	//instanciate the bucket

	svc := s3.New(session.New(&aws.Config{Region: aws.String(awsRegion)}))

	head, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: &bucketName,
	})

	if err != nil {
		return nil, err
	}

	//check for no such bucket
	if head == nil {

		result, err := svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: &bucketName,
		})

		if err != nil || result == nil {
			return nil, err
		}

		err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: &bucketName})

		if err != nil {
			return nil, err
		}
	}

	//now return the client

	return &S3PodSpec{
		s3:     svc,
		bucket: bucketName,
	}, nil
}

//WritePodSpec write the pod spec
func (s3PodSpec *S3PodSpec) WritePodSpec(namespace string, application string, revision string, podspec string) error {

  if valid, reason, err := ValidatePTS(podspec); err != nil {
    return err
  } else if !valid {
    return fmt.Errorf("Provided pod template spec is invalid: %s.", reason)
  }

	key := createS3Path(namespace, application, revision)

	//upload the json
	_, err := s3PodSpec.s3.PutObject(&s3.PutObjectInput{
		Body:   strings.NewReader(podspec),
		Bucket: &s3PodSpec.bucket,
		Key:    &key,
	})

	if err != nil {
		return err
	}

	return nil

}

//ReadPodSpec read the pod spec and return it as a string
func (s3PodSpec *S3PodSpec) ReadPodSpec(namespace string, application string, revision string) (*string, error) {

	key := createS3Path(namespace, application, revision)

	//upload the json
	downloadResult, err := s3PodSpec.s3.GetObject(&s3.GetObjectInput{
		Bucket: &s3PodSpec.bucket,
		Key:    &key,
	})

	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(downloadResult.Body)

	if err != nil {
		return nil, err
	}

	returnValue := string(bytes)
	return &returnValue, nil
}

//createPath create the path of the json
func createS3Path(namespace string, application string, revision string) string {
  // Changed this from %s/%s/%s
	return fmt.Sprintf("%s/%s:%s", namespace, application, revision)
}