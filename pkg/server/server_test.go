package server_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/30x/shipyard/pkg/server"
	"github.com/30x/shipyard/pkg/shipyard"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server Test", func() {

	//Use our test provider for jwt tokens
	os.Setenv("JWTTOKENIMPL", "test")

	ServerTests := func(testServer *server.Server, hostBase string) {

		It("Get Namespaces ", func() {

			httpResponse, namespaces, err := getNamespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			//we explicity don't test since other images might be present in the docker registry

			//now create a new images

			namespace := "test" + shipyard.UUIDString()
			application := "application"
			revision := "v1.0"

			response, err := newFileUploadRequest(hostBase, namespace, application, revision, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			bytes, err := ioutil.ReadAll(response.Body)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			//check build response
			lastLine := getLastLogLine(string(bytes))

			Expect(strings.Index(lastLine, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

			httpResponse, namespaces, err = getNamespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			assertContainsNamespace(namespaces, namespace)

		})

		FIt("Create Duplicate Application ", func() {
			//upload the first image
			namespace := "test" + shipyard.UUIDString()
			application := "application"
			revision := "v1.0"

			response, err := newFileUploadRequest(hostBase, namespace, application, revision, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, applications, err := getApplications(hostBase, namespace)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			assertContainsApplication(applications, application)

			//now try to post again, should get a 409

			response, err = newFileUploadRequest(hostBase, namespace, application, revision, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(409), "409 should be returned")

		})

		It("Test Application Images", func() {
			//upload the first image
			namespace := "test" + shipyard.UUIDString()
			application := "application"
			revision := "v1.0"

			response, err := newFileUploadRequest(hostBase, namespace, application, revision, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, images, err := getImages(hostBase, namespace, application)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(images) >= 1).Should(BeTrue(), "Images should be >= than length 1")

			Expect(images[0].Created).ShouldNot(BeNil())

			Expect(images[0].ImageID).ShouldNot(BeNil())

			Expect(strings.Index(images[0].ImageID, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

			//now try to post again, with a new revision

			revision2 := "v1.1"

			response, err = newFileUploadRequest(hostBase, namespace, application, revision2, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, images, err = getImages(hostBase, namespace, application)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(images) >= 2).Should(BeTrue(), "Images should be of length >= 2")

			Expect(images[0].Created).ShouldNot(BeNil())

			Expect(images[0].ImageID).ShouldNot(BeNil())

			Expect(strings.Index(images[0].ImageID, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

			Expect(images[1].Created).ShouldNot(BeNil())

			Expect(images[1].ImageID).ShouldNot(BeNil())
			Expect(strings.Index(images[1].ImageID, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

		})

	}

	Context("Local Docker", func() {
		//set up the provider
		os.Setenv("DOCKER_PROVIDER", "docker")
		os.Setenv("DOCKER_REGISTRY_URL", "localhost:5000")
		os.Setenv("POD_PROVIDER", "local")
		os.Setenv("LOCAL_DIR", "/tmp/podspecs")

		server, hostBase, err := doSetup(5280)

		if err != nil {
			Fail(fmt.Sprintf("Could not start server %s", err))
		}

		ServerTests(server, hostBase)
	})

	Context("ECR Docker", func() {

		//set up the provider
		os.Setenv("DOCKER_PROVIDER", "ecr")
		os.Setenv("DOCKER_REGISTRY_URL", "977777657611.dkr.ecr.us-east-1.amazonaws.com")
		os.Setenv("ECR_REGION", "us-east-1")
		os.Setenv("POD_PROVIDER", "s3")
		os.Setenv("S3_REGION", "us-east-1")
		os.Setenv("S3_BUCKET", "podspectestbucket")

		server, hostBase, err := doSetup(5281)

		if err != nil {
			Fail(fmt.Sprintf("Could not start server %s", err))
		}

		ServerTests(server, hostBase)
	})

})

//create a new instance of the server based on the env vars and the image creator.  Return them to be tested
func doSetup(port int) (*server.Server, string, error) {
	imageCreator, err := shipyard.NewImageCreatorFromEnv()

	if err != nil {
		return nil, "", err
	}

	podSpecProvider, err := shipyard.NewPodSpecIoFromEnv()

	if err != nil {
		return nil, "", err
	}

	testServer := server.NewServer(imageCreator, podSpecProvider)

	//start server in the background
	go func() {
		//start  the server and produce it to the start channel
		testServer.Start(port, 10*time.Second)
	}()

	//wait for it to start

	hostBase := fmt.Sprintf("http://localhost:%d/beeswax/images/api/v1", port)

	started := false

	//wait for host to start for 10 seconds
	for i := 0; i < 20; i++ {

		host := fmt.Sprintf("localhost:%d", port)

		conn, err := net.Dial("tcp", host)

		//done waiting, continue
		if err == nil {
			conn.Close()
			started = true
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if !started {
		return nil, "", errors.New("Server did not start")
	}

	return testServer, hostBase, nil
}

//assertContainsNamespace check if namespace array has the expected namespace in it
func assertContainsNamespace(namespaces []*server.Namespace, expectedNamespace string) {

	for _, namespace := range namespaces {
		if namespace.Name == expectedNamespace {
			return
		}
	}

	Fail(fmt.Sprintf("Could not find namespace %s in list returned", expectedNamespace))
}

func assertContainsApplication(applications []*server.Application, expectedApplication string) {

	for _, application := range applications {
		if application.Name == expectedApplication {
			return
		}
	}

	Fail(fmt.Sprintf("Could not find application %s in list returned", expectedApplication))
}

//getNamespaces perform a get request on namespaces
func getNamespaces(hostBase string) (*http.Response, []*server.Namespace, error) {

	url := fmt.Sprintf("%s/namespaces/", hostBase)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return nil, nil, err
	}

	repositories := []*server.Namespace{}

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, err
	}

	json.Unmarshal(bytes, &repositories)

	return response, repositories, err

}

//getApplications get the applications
func getApplications(hostBase string, namespace string) (*http.Response, []*server.Application, error) {
	url := getApplicationsURL(hostBase, namespace)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	repositories := []*server.Application{}

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, err
	}

	json.Unmarshal(bytes, &repositories)

	return response, repositories, err

}

//getImages get the images from the response
func getImages(hostBase string, namespace string, application string) (*http.Response, []*server.Image, error) {

	url := getImagesURL(hostBase, namespace, application)

	shipyard.LogInfo.Printf("Invoking get at URL %s", url)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	images := []*server.Image{}

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, err
	}

	body := string(bytes)

	shipyard.LogInfo.Printf("Response is %s", body)

	json.Unmarshal(bytes, &images)

	return response, images, err

}

//newfileUploadRequest upload a file form request
func newFileUploadRequest(hostBase string, namespace string, application string, revision string, path string) (*http.Response, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))

	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	if err != nil {
		return nil, err
	}

	writer.WriteField("namespace", namespace)
	writer.WriteField("application", application)
	writer.WriteField("revision", revision)

	//set the content type
	writer.FormDataContentType()

	err = writer.Close()

	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s/builds", hostBase)

	request, err := http.NewRequest("POST", uri, body)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	return client.Do(request)
}

//getApplicationsURL get the appplicationsUrl
func getApplicationsURL(hostBase string, namespace string) string {

	applicationsURL := fmt.Sprintf("%s/namespaces/%s/applications", hostBase, namespace)

	// shipyard.LogInfo.Printf("Creating URL %s", applicationsURL)

	return applicationsURL
}

func getApplicationURL(hostBase string, namespace string, application string) string {
	return fmt.Sprintf("%s/%s", getApplicationsURL(hostBase, namespace), application)
}

func getImagesURL(hostBase string, namespace string, application string) string {
	return fmt.Sprintf("%s/images/", getApplicationURL(hostBase, namespace, application))
}

func getLastLogLine(buildResponseBody string) string {

	scanner := bufio.NewScanner(strings.NewReader(buildResponseBody))

	scanner.Split(bufio.ScanLines)

	var line string

	for scanner.Scan() {

		line = scanner.Text()
	}

	return line

}
