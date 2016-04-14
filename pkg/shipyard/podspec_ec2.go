package shipyard

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

//Ec2PodSpec the ec2 pod spec
type Ec2PodSpec struct {
	//pointer to the s3 instance
	s3 *s3.S3

	//the stream of the bucket
	bucket string
}

//NewEc2PodSpec create a new EC2 pod spec to store pod specs in the given aws region and bucket
func NewEc2PodSpec(awsRegion string, bucketName string) (*Ec2PodSpec, error) {
	//instanciate the bucket

	svc := s3.New(session.New(&aws.Config{Region: aws.String(awsRegion)}))

	head, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: &bucketName,
	})

	//check for no such bucket
	if head == nil || err != nil {

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

	return &Ec2PodSpec{
		s3:     svc,
		bucket: bucketName,
	}, nil
}

//WritePodSpec write the pod spec
func (ec2PodSpec *Ec2PodSpec) WritePodSpec(namespace string, application string, revision string, podspec string) error {

	key := createPath(namespace, application, revision)

	//upload the json
	_, err := ec2PodSpec.s3.PutObject(&s3.PutObjectInput{
		Body:   strings.NewReader(podspec),
		Bucket: &ec2PodSpec.bucket,
		Key:    &key,
	})

	if err != nil {
		return err
	}

	return nil

}

//ReadPodSpec read the pod spec and return it as a string
func (ec2PodSpec *Ec2PodSpec) ReadPodSpec(namespace string, application string, revision string) (*string, error) {

	key := createPath(namespace, application, revision)

	//upload the json
	downloadResult, err := ec2PodSpec.s3.GetObject(&s3.GetObjectInput{
		Bucket: &ec2PodSpec.bucket,
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
func createPath(namespace string, application string, revision string) string {
	return fmt.Sprintf("%s/%s/%s", namespace, application, revision)
}
